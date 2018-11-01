package logfjournald

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/ssgreg/journald"
	"github.com/ssgreg/logf"
)

// NewEncoder creates the new instance of Encoder for journal format with
// the given FormatterConfig and TypeMarshallerFactory for non-basic types.
func NewEncoder(c *logf.FormatterConfig, mf logf.TypeMarshallerFactory) logf.Encoder {
	return &encoder{c, mf, nil, logf.NewCache(100)}
}

// NewTypeMarshallerFactory creates the new instance of TypeMarshallerFactory
// for journal message format with the given FormatterConfig and another
// TypeMarshallerFactory for non-basic types.
func NewTypeMarshallerFactory(c *logf.FormatterConfig, mf logf.TypeMarshallerFactory) logf.TypeMarshallerFactory {
	return &encoder{c, mf, nil, nil}
}

type encoder struct {
	*logf.FormatterConfig
	mf logf.TypeMarshallerFactory

	buf   *logf.Buffer
	cache *logf.Cache
}

// TypeMarshaller conforms to TypeMarshallerFactory interface.
func (f *encoder) TypeMarshaller(buf *logf.Buffer) logf.TypeMarshaller {
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

	if !f.DisableFieldLevel {
		// Both PRIORITY and FieldKeyLevel fields are usable.
		// 	- PRIORITY allows to use journal native features such as
		// filtering and color highlighting.
		// 	- FieldKeyLevel allows to check original severity level.

		// Add journal-compatible priority field on the base of entry's
		// severity level.
		f.MarshalFieldInt64("PRIORITY", int64(f.levelToPriority(e.Level)))
		if f.FieldKeyLevel != "" {
			// Add logf severity level using the given field name if
			// specified.
			f.MarshalFieldString(f.FieldKeyLevel, e.Level.String())
		}
	}
	if !f.DisableFieldMsg {
		// Ignore FieldKeyMsg in favor of journal-compatible MESSAGE.
		f.MarshalFieldString("MESSAGE", e.Text)
	}
	if !f.DisableFieldTime && f.FieldKeyTime != "" {
		f.MarshalFieldTime(f.FieldKeyTime, e.Time)
	}
	if !f.DisableFieldName && f.FieldKeyName != "" && e.LoggerName != "" {
		f.MarshalFieldString(f.FieldKeyName, e.LoggerName)
	}
	if !f.DisableFieldCaller && f.FieldKeyCaller != "" && e.Caller.Specified {
		f.addKey(f.FieldKeyCaller)
		f.FormatCaller(e.Caller, f)
	}

	for _, field := range e.Fields {
		field.Accept(f)
	}

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

	return nil
}

func (f *encoder) MarshalFieldAny(k string, v interface{}) {
	f.addKey(k)
	f.MarshalAny(v)
}

func (f *encoder) MarshalFieldBool(k string, v bool) {
	f.addKey(k)
	f.MarshalBool(v)
}

func (f *encoder) MarshalFieldInt64(k string, v int64) {
	f.addKey(k)
	f.MarshalInt64(v)
}

func (f *encoder) MarshalFieldInt32(k string, v int32) {
	f.addKey(k)
	f.MarshalInt32(v)
}

func (f *encoder) MarshalFieldInt16(k string, v int16) {
	f.addKey(k)
	f.MarshalInt16(v)
}

func (f *encoder) MarshalFieldInt8(k string, v int8) {
	f.addKey(k)
	f.MarshalInt8(v)
}

func (f *encoder) MarshalFieldUint64(k string, v uint64) {
	f.addKey(k)
	f.MarshalUint64(v)
}

func (f *encoder) MarshalFieldUint32(k string, v uint32) {
	f.addKey(k)
	f.MarshalUint32(v)
}

func (f *encoder) MarshalFieldUint16(k string, v uint16) {
	f.addKey(k)
	f.MarshalUint16(v)
}

func (f *encoder) MarshalFieldUint8(k string, v uint8) {
	f.addKey(k)
	f.MarshalUint8(v)
}

func (f *encoder) MarshalFieldFloat64(k string, v float64) {
	f.addKey(k)
	f.MarshalFloat64(v)
}

func (f *encoder) MarshalFieldFloat32(k string, v float32) {
	f.addKey(k)
	f.MarshalFloat32(v)
}

func (f *encoder) MarshalFieldString(k string, v string) {
	f.addKey(k)
	f.MarshalString(v)
}

func (f *encoder) MarshalFieldDuration(k string, v time.Duration) {
	f.addKey(k)
	f.MarshalDuration(v)
}

func (f *encoder) MarshalFieldError(k string, v error) {
	// The only exception that has no MarshalX function. FormatError can add
	// new fields by itself.
	f.FormatError(k, v, f)
}

func (f *encoder) MarshalFieldTime(k string, v time.Time) {
	f.addKey(k)
	f.MarshalTime(v)
}

func (f *encoder) MarshalFieldArray(k string, v logf.ArrayMarshaller) {
	f.addKey(k)
	f.MarshalArray(v)
}

func (f *encoder) MarshalFieldObject(k string, v logf.ObjectMarshaller) {
	f.addKey(k)
	f.MarshalObject(v)
}

func (f *encoder) MarshalFieldBytes(k string, v []byte) {
	f.addKey(k)
	f.MarshalBytes(v)
}

func (f *encoder) MarshalFieldBools(k string, v []bool) {
	f.addKey(k)
	f.MarshalBools(v)
}

func (f *encoder) MarshalFieldInts64(k string, v []int64) {
	f.addKey(k)
	f.MarshalInts64(v)
}

func (f *encoder) MarshalFieldInts32(k string, v []int32) {
	f.addKey(k)
	f.MarshalInts32(v)
}

func (f *encoder) MarshalFieldInts16(k string, v []int16) {
	f.addKey(k)
	f.MarshalInts16(v)
}

func (f *encoder) MarshalFieldInts8(k string, v []int8) {
	f.addKey(k)
	f.MarshalInts8(v)
}

func (f *encoder) MarshalFieldUints64(k string, v []uint64) {
	f.addKey(k)
	f.MarshalUints64(v)
}

func (f *encoder) MarshalFieldUints32(k string, v []uint32) {
	f.addKey(k)
	f.MarshalUints32(v)
}

func (f *encoder) MarshalFieldUints16(k string, v []uint16) {
	f.addKey(k)
	f.MarshalUints16(v)
}

func (f *encoder) MarshalFieldUints8(k string, v []uint8) {
	f.addKey(k)
	f.MarshalUints8(v)
}

func (f *encoder) MarshalFieldFloats64(k string, v []float64) {
	f.addKey(k)
	f.MarshalFloats64(v)
}

func (f *encoder) MarshalFieldFloats32(k string, v []float32) {
	f.addKey(k)
	f.MarshalFloats32(v)
}

func (f *encoder) MarshalFieldDurations(k string, v []time.Duration) {
	f.addKey(k)
	f.MarshalDurations(v)
}

func (f *encoder) MarshalAny(v interface{}) {
	f.withValue(func() {
		if !knownTypeToBuf(f.buf, v) {
			f.mf.TypeMarshaller(f.buf).MarshalAny(v)
		}
	})
}

func (f *encoder) MarshalByte(v byte) {
	f.withValue(func() {
		logf.AppendInt(f.buf, int64(v))
	})
}

func (f *encoder) MarshalUnsafeBytes(v unsafe.Pointer) {
	f.withValue(func() {
		f.buf.AppendBytes(*(*[]byte)(v))
	})
}

func (f *encoder) MarshalBool(v bool) {
	f.withValue(func() {
		logf.AppendBool(f.buf, v)
	})
}

func (f *encoder) MarshalString(v string) {
	f.withValue(func() {
		f.buf.AppendString(v)
	})
}

func (f *encoder) MarshalInt64(v int64) {
	f.withValue(func() {
		logf.AppendInt(f.buf, v)
	})
}
func (f *encoder) MarshalInt32(v int32) {
	f.withValue(func() {
		logf.AppendInt(f.buf, int64(v))
	})
}

func (f *encoder) MarshalInt16(v int16) {
	f.withValue(func() {
		logf.AppendInt(f.buf, int64(v))
	})
}

func (f *encoder) MarshalInt8(v int8) {
	f.withValue(func() {
		logf.AppendInt(f.buf, int64(v))
	})
}

func (f *encoder) MarshalUint64(v uint64) {
	f.withValue(func() {
		logf.AppendUint(f.buf, v)
	})
}

func (f *encoder) MarshalUint32(v uint32) {
	f.withValue(func() {
		logf.AppendUint(f.buf, uint64(v))
	})
}

func (f *encoder) MarshalUint16(v uint16) {
	f.withValue(func() {
		logf.AppendUint(f.buf, uint64(v))
	})
}

func (f *encoder) MarshalUint8(v uint8) {
	f.withValue(func() {
		logf.AppendUint(f.buf, uint64(v))
	})
}

func (f *encoder) MarshalFloat64(v float64) {
	f.withValue(func() {
		logf.AppendFloat64(f.buf, v)
	})
}

func (f *encoder) MarshalFloat32(v float32) {
	f.withValue(func() {
		logf.AppendFloat32(f.buf, v)
	})
}

func (f *encoder) MarshalDuration(v time.Duration) {
	f.FormatDuration(v, f)
}

func (f *encoder) MarshalTime(v time.Time) {
	f.FormatTime(v, f)
}

func (f *encoder) MarshalBytes(v []byte) {
	f.withValue(func() {
		base64.StdEncoding.Encode(f.buf.ExtendBytes(base64.StdEncoding.EncodedLen(len(v))), v)
	})
}

func (f *encoder) MarshalBools(v []bool) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalBools(v)
	})
}

func (f *encoder) MarshalInts64(v []int64) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalInts64(v)
	})
}

func (f *encoder) MarshalInts32(v []int32) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalInts32(v)
	})
}

func (f *encoder) MarshalInts16(v []int16) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalInts16(v)
	})
}

func (f *encoder) MarshalInts8(v []int8) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalInts8(v)
	})
}

func (f *encoder) MarshalUints64(v []uint64) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalUints64(v)
	})
}

func (f *encoder) MarshalUints32(v []uint32) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalUints32(v)
	})
}

func (f *encoder) MarshalUints16(v []uint16) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalUints16(v)
	})
}

func (f *encoder) MarshalUints8(v []uint8) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalUints8(v)
	})
}

func (f *encoder) MarshalFloats64(v []float64) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalFloats64(v)
	})
}

func (f *encoder) MarshalFloats32(v []float32) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalFloats32(v)
	})
}

func (f *encoder) MarshalDurations(v []time.Duration) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalDurations(v)
	})
}

func (f *encoder) MarshalArray(v logf.ArrayMarshaller) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalArray(v)
	})
}

func (f *encoder) MarshalObject(v logf.ObjectMarshaller) {
	f.withValue(func() {
		f.mf.TypeMarshaller(f.buf).MarshalObject(v)
	})
}

func (f *encoder) addKey(k string) {
	appendNormalizedKey(f.buf, k)
}

func (f *encoder) withValue(fn func()) {
	// According to the format, if the value includes a newline
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

func knownTypeToBuf(buf *logf.Buffer, v interface{}) bool {
	switch rv := v.(type) {
	case string:
		buf.AppendString(rv)
	case bool:
		logf.AppendBool(buf, rv)
	case int:
		logf.AppendInt(buf, int64(rv))
	case int8:
		logf.AppendInt(buf, int64(rv))
	case int16:
		logf.AppendInt(buf, int64(rv))
	case int32:
		logf.AppendInt(buf, int64(rv))
	case int64:
		logf.AppendInt(buf, rv)
	case uint:
		logf.AppendUint(buf, uint64(rv))
	case uint8:
		logf.AppendUint(buf, uint64(rv))
	case uint16:
		logf.AppendUint(buf, uint64(rv))
	case uint32:
		logf.AppendUint(buf, uint64(rv))
	case uint64:
		logf.AppendUint(buf, rv)
	case float32:
		logf.AppendFloat32(buf, rv)
	case float64:
		logf.AppendFloat64(buf, rv)
	case fmt.Stringer:
		buf.AppendString(rv.String())
	case error:
		buf.AppendString(rv.Error())
	default:
		if rv == nil {
			return false
		}
		switch reflect.TypeOf(rv).Kind() {
		case reflect.Bool:
			logf.AppendBool(buf, reflect.ValueOf(rv).Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			logf.AppendInt(buf, reflect.ValueOf(rv).Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			logf.AppendUint(buf, reflect.ValueOf(rv).Uint())
		case reflect.Float32, reflect.Float64:
			logf.AppendFloat64(buf, reflect.ValueOf(rv).Float())
		default:
			return false
		}
	}

	return true
}
