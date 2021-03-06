package logfjournald

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ssgreg/logf"
	"github.com/stretchr/testify/require"
)

type encoderTestCase struct {
	Name   string
	Entry  []logf.Entry
	Golden []byte
}

// Special error that returns full text using fmt.Formatter.
// logf.DefaultErrorFormatter adds it as an additional verbose field
// to the encoded message.
func newError(short string, full string) error {
	return &myError{short, full}
}

type myError struct {
	short string
	full  string
}

func (e *myError) Error() string {
	return e.short
}

func (e *myError) Format(f fmt.State, c rune) {
	f.Write([]byte(e.full))
}

type user struct {
	Name string `json:"name"`
}
type users []*user

func (u *user) EncodeLogfObject(enc logf.FieldEncoder) error {
	enc.EncodeFieldString("name", u.Name)

	return nil
}

func (u users) EncodeLogfArray(enc logf.TypeEncoder) error {
	for i := range u {
		enc.EncodeTypeObject(u[i])
	}

	return nil
}

type MyInt int
type MyUint uint
type MyBool bool
type MyFloat float64
type MyString string

func TestEncoder(t *testing.T) {
	testCases := []encoderTestCase{
		{
			"Simple",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelInfo,
					Text:     "message",
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '6', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'i', 'n', 'f', 'o', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
			},
		},
		{
			"WithFields",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelError,
					Text:     "message",
					Fields: []logf.Field{
						logf.String("str", "sv"),
					},
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '3', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, 'e', 'r', 'r', 'o', 'r', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'S', 'T', 'R', '\n', 0x2, 0, 0, 0, 0, 0, 0, 0, 's', 'v', '\n',
			},
		},
		{
			"WithFieldsAndDerivedFields",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelDebug,
					Text:     "message",
					Fields: []logf.Field{
						logf.ConstInts("ints", []int{0, 1}),
					},
					DerivedFields: []logf.Field{
						logf.ConstBytes("bytes", []byte("!")),
					},
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '7', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, 'd', 'e', 'b', 'u', 'g', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'B', 'Y', 'T', 'E', 'S', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'I', 'Q', '=', '=', '\n',
				'I', 'N', 'T', 'S', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
			},
		},
		{
			"WithFieldsAndDerivedFieldsAndCaller",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelWarn,
					Text:     "message",
					Fields: []logf.Field{
						logf.ConstFloats32("fts", []float32{0.1, 9}),
					},
					DerivedFields: []logf.Field{
						logf.ConstDurations("drs", []time.Duration{time.Second}),
					},
					Caller: logf.EntryCaller{
						PC:        0,
						File:      "/a/b/c/f.go",
						Line:      6,
						Specified: true,
					},
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '4', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'w', 'a', 'r', 'n', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'C', 'A', 'L', 'L', 'E', 'R', '\n', 0x8, 0, 0, 0, 0, 0, 0, 0, 'c', '/', 'f', '.', 'g', 'o', ':', '6', '\n',
				'D', 'R', 'S', '\n', 0x6, 0, 0, 0, 0, 0, 0, 0, '[', '"', '1', 's', '"', ']', '\n',
				'F', 'T', 'S', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, '[', '0', '.', '1', ',', '9', ']', '\n',
			},
		},
		{
			"WithFieldsAndDerivedFieldsAndName",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelWarn,
					Text:     "message",
					Fields: []logf.Field{
						logf.ConstUints32("uts", []uint32{8}),
					},
					DerivedFields: []logf.Field{
						logf.NamedError("err", newError("s", "f")),
					},
					LoggerName: "n",
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '4', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'w', 'a', 'r', 'n', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'L', 'O', 'G', 'G', 'E', 'R', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, 'n', '\n',
				'E', 'R', 'R', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, 's', '\n',
				'E', 'R', 'R', '_', 'V', 'E', 'R', 'B', 'O', 'S', 'E', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, 'f', '\n',
				'U', 'T', 'S', '\n', 0x3, 0, 0, 0, 0, 0, 0, 0, '[', '8', ']', '\n',
			},
		},
		{
			"WithArrayAndObjectFields",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelWarn,
					Text:     "message",
					Fields: []logf.Field{
						logf.Object("o", &user{"n"}),
						logf.Array("a", users{{"n1"}, {"n2"}}),
					},
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '4', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'w', 'a', 'r', 'n', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'O', '\n', 0xc, 0, 0, 0, 0, 0, 0, 0, '{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 'n', '"', '}', '\n',
				'A', '\n', 0x1d, 0, 0, 0, 0, 0, 0, 0, '[', '{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 'n', '1', '"', '}', ',', '{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 'n', '2', '"', '}', ']', '\n',
			},
		},
		{
			"DoubleMesage",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelInfo,
					Text:     "m1",
				},
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelInfo,
					Text:     "m2",
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '6', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'i', 'n', 'f', 'o', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x2, 0, 0, 0, 0, 0, 0, 0, 'm', '1', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'\n',
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '6', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'i', 'n', 'f', 'o', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x2, 0, 0, 0, 0, 0, 0, 0, 'm', '2', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
			},
		},
		{
			"WithAnyFields",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelWarn,
					Text:     "message",
					Fields: []logf.Field{
						logf.Any("i", 1),
						logf.Any("u", &user{"n"}),
					},
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '4', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'w', 'a', 'r', 'n', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'I', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '\n', 0xc, 0, 0, 0, 0, 0, 0, 0, '{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 'n', '"', '}', '\n',
			},
		},
		{
			"WithCompleteListOfAnyFields",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelWarn,
					Text:     "message",
					Fields: []logf.Field{
						logf.Any("s", "3"),
						logf.Any("b", true),
						logf.Any("i", int(1)),
						logf.Any("i8", int8(1)),
						logf.Any("i16", int16(1)),
						logf.Any("i32", int32(1)),
						logf.Any("i64", int64(1)),
						logf.Any("u", uint(1)),
						logf.Any("u8", uint8(1)),
						logf.Any("u16", uint16(1)),
						logf.Any("u32", uint32(1)),
						logf.Any("u64", uint64(1)),
						logf.Any("f32", float32(1)),
						logf.Any("f64", float64(1)),
						logf.Any("e", errors.New("1")),
						logf.Any("ri", MyInt(1)),
						logf.Any("ru", MyUint(1)),
						logf.Any("rb", MyBool(true)),
						logf.Any("rs", MyString("2")),
						logf.Any("rf", MyFloat(1)),
					},
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '4', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'w', 'a', 'r', 'n', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'S', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '3', '\n',
				'B', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 't', 'r', 'u', 'e', '\n',
				'I', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'I', '8', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'I', '1', '6', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'I', '3', '2', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'I', '6', '4', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '8', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '1', '6', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '3', '2', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '6', '4', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'F', '3', '2', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'F', '6', '4', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'E', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'R', 'I', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'R', 'U', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'R', 'B', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 't', 'r', 'u', 'e', '\n',
				'R', 'S', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '2', '\n',
				'R', 'F', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
			},
		},
		{
			"WithCompleteListOfFields",
			[]logf.Entry{
				{
					LoggerID: int32(rand.Int()),
					Level:    logf.LevelWarn,
					Text:     "message",
					Fields: []logf.Field{
						logf.String("s", "1"),
						logf.Bool("b", true),
						logf.Int("i", 1),
						logf.Int8("i8", 1),
						logf.Int16("i16", 1),
						logf.Int32("i32", 1),
						logf.Int64("i64", 1),
						logf.Uint("u", 1),
						logf.Uint8("u8", 1),
						logf.Uint16("u16", 1),
						logf.Uint32("u32", 1),
						logf.Uint64("u64", 1),
						logf.Float32("f32", 1),
						logf.Float64("f64", 1),
						logf.Duration("d", 1),
						logf.ConstBools("bs", []bool{true, false}),
						logf.ConstInts("is", []int{0, 1}),
						logf.ConstInts8("is8", []int8{0, 1}),
						logf.ConstInts16("is16", []int16{0, 1}),
						logf.ConstInts32("is32", []int32{0, 1}),
						logf.ConstInts64("is64", []int64{0, 1}),
						logf.ConstUints("us", []uint{0, 1}),
						logf.ConstUints8("us8", []uint8{0, 1}),
						logf.ConstUints16("us16", []uint16{0, 1}),
						logf.ConstUints32("us32", []uint32{0, 1}),
						logf.ConstUints64("us64", []uint64{0, 1}),
						logf.ConstFloats32("fs32", []float32{0, 1}),
						logf.ConstFloats64("fs64", []float64{0, 1}),
					},
				},
			},
			[]byte{
				'P', 'R', 'I', 'O', 'R', 'I', 'T', 'Y', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '4', '\n',
				'L', 'E', 'V', 'E', 'L', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 'w', 'a', 'r', 'n', '\n',
				'M', 'E', 'S', 'S', 'A', 'G', 'E', '\n', 0x7, 0, 0, 0, 0, 0, 0, 0, 'm', 'e', 's', 's', 'a', 'g', 'e', '\n',
				'T', 'S', '\n', 0x14, 0, 0, 0, 0, 0, 0, 0, '0', '0', '0', '1', '-', '0', '1', '-', '0', '1', 'T', '0', '0', ':', '0', '0', ':', '0', '0', 'Z', '\n',
				'S', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'B', '\n', 0x4, 0, 0, 0, 0, 0, 0, 0, 't', 'r', 'u', 'e', '\n',
				'I', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'I', '8', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'I', '1', '6', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'I', '3', '2', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'I', '6', '4', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '8', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '1', '6', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '3', '2', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'U', '6', '4', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'F', '3', '2', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'F', '6', '4', '\n', 0x1, 0, 0, 0, 0, 0, 0, 0, '1', '\n',
				'D', '\n', 0x3, 0, 0, 0, 0, 0, 0, 0, '1', 'n', 's', '\n',
				'B', 'S', '\n', 0xc, 0, 0, 0, 0, 0, 0, 0, '[', 't', 'r', 'u', 'e', ',', 'f', 'a', 'l', 's', 'e', ']', '\n',
				'I', 'S', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'I', 'S', '8', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'I', 'S', '1', '6', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'I', 'S', '3', '2', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'I', 'S', '6', '4', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'U', 'S', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'U', 'S', '8', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'U', 'S', '1', '6', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'U', 'S', '3', '2', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'U', 'S', '6', '4', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'F', 'S', '3', '2', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
				'F', 'S', '6', '4', '\n', 0x5, 0, 0, 0, 0, 0, 0, 0, '[', '0', ',', '1', ']', '\n',
			},
		},
	}

	enc := NewEncoder.Default()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			b := logf.NewBuffer()
			for _, e := range tc.Entry {
				enc.Encode(b, e)
			}

			require.EqualValues(t, tc.Golden, b.Bytes())
		})
	}
}
