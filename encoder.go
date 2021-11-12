package logfjournald

import (
	"encoding/base64"
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/ssgreg/journald"
	"github.com/ssgreg/logf"
)

// NewEncoder creates the new instance of Encoder for journal Encode with
// the given EncoderConfig and TypeEncoderFactory for non-basic types.
var NewEncoder = jsonEncoderGetter(
	func(c EncoderConfig, mf logf.TypeEncoderFactory) logf.Encoder {
		return &encoder{c.WithDefaults(), mf, nil, logf.NewCache(100)}
	},
)

// NewTypeEncoderFactory creates the new instance of TypeEncoderFactory
// for journal message Encode with the given EncoderConfig and another
// TypeEncoderFactory for non-basic types.
var NewTypeEncoderFactory = jsonTypeEncoderFactoryGetter(
	func(c EncoderConfig, mf logf.TypeEncoderFactory) logf.TypeEncoderFactory {
		return &encoder{c.WithDefaults(), mf, nil, nil}
	},
)

type jsonEncoderGetter func(c EncoderConfig, mf logf.TypeEncoderFactory) logf.Encoder

func (c jsonEncoderGetter) Default() logf.Encoder {
	return c(EncoderConfig{}, logf.NewJSONTypeEncoderFactory.Default())
}

type jsonTypeEncoderFactoryGetter func(c EncoderConfig, mf logf.TypeEncoderFactory) logf.TypeEncoderFactory

func (c jsonTypeEncoderFactoryGetter) Default() logf.TypeEncoderFactory {
	return c(EncoderConfig{}, logf.NewJSONTypeEncoderFactory.Default())
}

type encoder struct {
	EncoderConfig
	mf logf.TypeEncoderFactory

	buf   *logf.Buffer
	cache *logf.Cache
}

// TypeEncoder conforms to TypeEncoderFactory interface.
func (f *encoder) TypeEncoder(buf *logf.Buffer) logf.TypeEncoder {
	f.buf = buf

	return f
}

// Encode conforms to Encoder interface.
func (f *encoder) Encode(buf *logf.Buffer, e logf.Entry) error {
	f.buf = buf

	// There are messages in buffer already. Add message separator.
	if f.buf.Len() != 0 {
		f.buf.AppendByte('\n')
	}

	// PRIORITY.
	if !f.DisableFieldPriority {
		f.EncodeFieldInt64(DefaultFieldKeyPriority, int64(f.levelToPriority(e.Level)))
	}

	// Level.
	if !f.DisableFieldLevel {
		f.addKey(f.FieldKeyLevel)
		f.EncodeLevel(e.Level, f)
	}

	// MESSAGE.
	f.EncodeFieldString(DefaultFieldKeyMessage, e.Text)

	// Time.
	if !f.DisableFieldTime {
		f.EncodeFieldTime(f.FieldKeyTime, e.Time)
	}

	// Logger name.
	if !f.DisableFieldName && e.LoggerName != "" {
		f.EncodeFieldString(f.FieldKeyName, e.LoggerName)
	}

	// Caller.
	if !f.DisableFieldCaller && e.Caller.Specified {
		f.addKey(f.FieldKeyCaller)
		f.EncodeCaller(e.Caller, f)
	}

	// Logger fields.
	if bytes, ok := f.cache.Get(e.LoggerID); ok {
		buf.AppendBytes(bytes)
	} else {
		le := buf.Len()
		for _, field := range e.DerivedFields {
			field.Accept(f)
		}

		bf := make([]byte, buf.Len()-le)
		copy(bf, buf.Data[le:])
		f.cache.Set(e.LoggerID, bf)
	}

	// Entry's fields.
	for _, field := range e.Fields {
		field.Accept(f)
	}

	return nil
}

func (f *encoder) EncodeFieldAny(k string, v interface{}) {
	f.addKey(k)
	f.EncodeTypeAny(v)
}

func (f *encoder) EncodeFieldBool(k string, v bool) {
	f.addKey(k)
	f.EncodeTypeBool(v)
}

func (f *encoder) EncodeFieldInt64(k string, v int64) {
	f.addKey(k)
	f.EncodeTypeInt64(v)
}

func (f *encoder) EncodeFieldInt32(k string, v int32) {
	f.addKey(k)
	f.EncodeTypeInt32(v)
}

func (f *encoder) EncodeFieldInt16(k string, v int16) {
	f.addKey(k)
	f.EncodeTypeInt16(v)
}

func (f *encoder) EncodeFieldInt8(k string, v int8) {
	f.addKey(k)
	f.EncodeTypeInt8(v)
}

func (f *encoder) EncodeFieldUint64(k string, v uint64) {
	f.addKey(k)
	f.EncodeTypeUint64(v)
}

func (f *encoder) EncodeFieldUint32(k string, v uint32) {
	f.addKey(k)
	f.EncodeTypeUint32(v)
}

func (f *encoder) EncodeFieldUint16(k string, v uint16) {
	f.addKey(k)
	f.EncodeTypeUint16(v)
}

func (f *encoder) EncodeFieldUint8(k string, v uint8) {
	f.addKey(k)
	f.EncodeTypeUint8(v)
}

func (f *encoder) EncodeFieldFloat64(k string, v float64) {
	f.addKey(k)
	f.EncodeTypeFloat64(v)
}

func (f *encoder) EncodeFieldFloat32(k string, v float32) {
	f.addKey(k)
	f.EncodeTypeFloat32(v)
}

func (f *encoder) EncodeFieldString(k string, v string) {
	f.addKey(k)
	f.EncodeTypeString(v)
}

func (f *encoder) EncodeFieldDuration(k string, v time.Duration) {
	f.addKey(k)
	f.EncodeTypeDuration(v)
}

func (f *encoder) EncodeFieldError(k string, v error) {
	// The only exception that has no EncodeTypeX function. EncodeError can add
	// new fields by itself.
	f.EncodeError(k, v, f)
}

func (f *encoder) EncodeFieldTime(k string, v time.Time) {
	f.addKey(k)
	f.EncodeTypeTime(v)
}

func (f *encoder) EncodeFieldArray(k string, v logf.ArrayEncoder) {
	f.addKey(k)
	f.EncodeTypeArray(v)
}

func (f *encoder) EncodeFieldObject(k string, v logf.ObjectEncoder) {
	f.addKey(k)
	f.EncodeTypeObject(v)
}

func (f *encoder) EncodeFieldBytes(k string, v []byte) {
	f.addKey(k)
	f.EncodeTypeBytes(v)
}

func (f *encoder) EncodeFieldBools(k string, v []bool) {
	f.addKey(k)
	f.EncodeTypeBools(v)
}

func (f *encoder) EncodeFieldStrings(k string, v []string) {
	f.addKey(k)
	f.EncodeTypeStrings(v)
}

func (f *encoder) EncodeFieldInts64(k string, v []int64) {
	f.addKey(k)
	f.EncodeTypeInts64(v)
}

func (f *encoder) EncodeFieldInts32(k string, v []int32) {
	f.addKey(k)
	f.EncodeTypeInts32(v)
}

func (f *encoder) EncodeFieldInts16(k string, v []int16) {
	f.addKey(k)
	f.EncodeTypeInts16(v)
}

func (f *encoder) EncodeFieldInts8(k string, v []int8) {
	f.addKey(k)
	f.EncodeTypeInts8(v)
}

func (f *encoder) EncodeFieldUints64(k string, v []uint64) {
	f.addKey(k)
	f.EncodeTypeUints64(v)
}

func (f *encoder) EncodeFieldUints32(k string, v []uint32) {
	f.addKey(k)
	f.EncodeTypeUints32(v)
}

func (f *encoder) EncodeFieldUints16(k string, v []uint16) {
	f.addKey(k)
	f.EncodeTypeUints16(v)
}

func (f *encoder) EncodeFieldUints8(k string, v []uint8) {
	f.addKey(k)
	f.EncodeTypeUints8(v)
}

func (f *encoder) EncodeFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.EncodeTypeFloats64(v)
}

func (f *encoder) EncodeFieldFloats32(k string, v []float32) {
	f.addKey(k)
	f.EncodeTypeFloats32(v)
}

func (f *encoder) EncodeFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.EncodeTypeDurations(v)
}

func (f *encoder) EncodeTypeAny(v interface{}) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeAny(v)
	})
}

func (f *encoder) EncodeTypeByte(v byte) {
	f.withValue(func() {
		logf.AppendInt(f.buf, int64(v))
	})
}

func (f *encoder) EncodeTypeUnsafeBytes(v unsafe.Pointer) {
	f.withValue(func() {
		f.buf.AppendBytes(*(*[]byte)(v))
	})
}

func (f *encoder) EncodeTypeBool(v bool) {
	f.withValue(func() {
		logf.AppendBool(f.buf, v)
	})
}

func (f *encoder) EncodeTypeString(v string) {
	f.withValue(func() {
		f.buf.AppendString(v)
	})
}

func (f *encoder) EncodeTypeInt64(v int64) {
	f.withValue(func() {
		logf.AppendInt(f.buf, v)
	})
}
func (f *encoder) EncodeTypeInt32(v int32) {
	f.withValue(func() {
		logf.AppendInt(f.buf, int64(v))
	})
}

func (f *encoder) EncodeTypeInt16(v int16) {
	f.withValue(func() {
		logf.AppendInt(f.buf, int64(v))
	})
}

func (f *encoder) EncodeTypeInt8(v int8) {
	f.withValue(func() {
		logf.AppendInt(f.buf, int64(v))
	})
}

func (f *encoder) EncodeTypeUint64(v uint64) {
	f.withValue(func() {
		logf.AppendUint(f.buf, v)
	})
}

func (f *encoder) EncodeTypeUint32(v uint32) {
	f.withValue(func() {
		logf.AppendUint(f.buf, uint64(v))
	})
}

func (f *encoder) EncodeTypeUint16(v uint16) {
	f.withValue(func() {
		logf.AppendUint(f.buf, uint64(v))
	})
}

func (f *encoder) EncodeTypeUint8(v uint8) {
	f.withValue(func() {
		logf.AppendUint(f.buf, uint64(v))
	})
}

func (f *encoder) EncodeTypeFloat64(v float64) {
	f.withValue(func() {
		logf.AppendFloat64(f.buf, v)
	})
}

func (f *encoder) EncodeTypeFloat32(v float32) {
	f.withValue(func() {
		logf.AppendFloat32(f.buf, v)
	})
}

func (f *encoder) EncodeTypeDuration(v time.Duration) {
	f.EncodeDuration(v, f)
}

func (f *encoder) EncodeTypeTime(v time.Time) {
	f.EncodeTime(v, f)
}

func (f *encoder) EncodeTypeBytes(v []byte) {
	f.withValue(func() {
		base64.StdEncoding.Encode(f.buf.ExtendBytes(base64.StdEncoding.EncodedLen(len(v))), v)
	})
}

func (f *encoder) EncodeTypeBools(v []bool) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeBools(v)
	})
}

func (f *encoder) EncodeTypeStrings(v []string) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeStrings(v)
	})
}

func (f *encoder) EncodeTypeInts64(v []int64) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeInts64(v)
	})
}

func (f *encoder) EncodeTypeInts32(v []int32) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeInts32(v)
	})
}

func (f *encoder) EncodeTypeInts16(v []int16) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeInts16(v)
	})
}

func (f *encoder) EncodeTypeInts8(v []int8) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeInts8(v)
	})
}

func (f *encoder) EncodeTypeUints64(v []uint64) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeUints64(v)
	})
}

func (f *encoder) EncodeTypeUints32(v []uint32) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeUints32(v)
	})
}

func (f *encoder) EncodeTypeUints16(v []uint16) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeUints16(v)
	})
}

func (f *encoder) EncodeTypeUints8(v []uint8) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeUints8(v)
	})
}

func (f *encoder) EncodeTypeFloats64(v []float64) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeFloats64(v)
	})
}

func (f *encoder) EncodeTypeFloats32(v []float32) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeFloats32(v)
	})
}

func (f *encoder) EncodeTypeDurations(v []time.Duration) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeDurations(v)
	})
}

func (f *encoder) EncodeTypeArray(v logf.ArrayEncoder) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeArray(v)
	})
}

func (f *encoder) EncodeTypeObject(v logf.ObjectEncoder) {
	f.withValue(func() {
		f.mf.TypeEncoder(f.buf).EncodeTypeObject(v)
	})
}

func (f *encoder) addKey(k string) {
	appendNormalizedKey(f.buf, k)
}

func (f *encoder) withValue(fn func()) {
	// According to the Encode, if the value includes a newline
	// need to write the field name, plus a newline, then the
	// size (64bit LE), the field data and a final newline.

	f.buf.AppendByte('\n')
	sizeBytes := f.buf.ExtendBytes(8)
	pos := f.buf.Len()

	fn()

	binary.LittleEndian.PutUint64(sizeBytes, uint64(f.buf.Len()-pos))
	f.buf.AppendByte('\n')
}

func (f *encoder) levelToPriority(lvl logf.Level) journald.Priority {
	switch lvl {
	case logf.LevelDebug:
		return journald.PriorityDebug
	case logf.LevelInfo:
		return journald.PriorityInfo
	case logf.LevelWarn:
		return journald.PriorityWarning
	case logf.LevelError:
		return journald.PriorityErr
	default:
		return journald.PriorityNotice
	}
}
