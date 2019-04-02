package logfjournald

import "github.com/ssgreg/logf"

// Default field keys.
const (
	DefaultFieldKeyLevel  = "LEVEL"
	DefaultFieldKeyTime   = "TS"
	DefaultFieldKeyName   = "LOGGER"
	DefaultFieldKeyCaller = "CALLER"

	// Systemd journal dependent field keys.
	DefaultFieldKeyPriority = "PRIORITY"
	DefaultFieldKeyMessage  = "MESSAGE"
)

// EncoderConfig allows to configure journal Encoder.
//
// Note that PRIORITY and MESSAGE field names could not be configured.
// Both PRIORITY and FieldKeyLevel fields are usable.
// 	- PRIORITY allows to use journal native features such as
// filtering and color highlighting.
// 	- FieldKeyLevel allows to check original severity level.
//
type EncoderConfig struct {
	FieldKeyTime   string
	FieldKeyLevel  string
	FieldKeyName   string
	FieldKeyCaller string

	// DisableFieldLevel disabled the Time field.
	// Native journal's time field (when the message was added to the
	// journal) stayes enabled.
	DisableFieldTime bool

	// DisableFieldLevel disables the Level field.
	DisableFieldLevel bool

	// DisableFieldPriority disables the native journal's PRIORITY field.
	DisableFieldPriority bool

	// DisableFieldName disables the logger name filed.
	DisableFieldName bool

	// DisableFieldCaller disables the caller field.
	DisableFieldCaller bool

	EncodeTime     logf.TimeEncoder
	EncodeDuration logf.DurationEncoder
	EncodeError    logf.ErrorEncoder
	EncodeLevel    logf.LevelEncoder
	EncodeCaller   logf.CallerEncoder
}

// WithDefaults returns the new config in which all uninitialized fields are
// filled with their default values.
func (c EncoderConfig) WithDefaults() EncoderConfig {
	// Handle default for predefined field names.
	if c.FieldKeyLevel == "" {
		c.FieldKeyLevel = DefaultFieldKeyLevel
	}
	if c.FieldKeyTime == "" {
		c.FieldKeyTime = DefaultFieldKeyTime
	}
	if c.FieldKeyName == "" {
		c.FieldKeyName = DefaultFieldKeyName
	}
	if c.FieldKeyCaller == "" {
		c.FieldKeyCaller = DefaultFieldKeyCaller
	}

	// Handle defaults for type encoder.
	if c.EncodeDuration == nil {
		c.EncodeDuration = logf.StringDurationEncoder
	}
	if c.EncodeTime == nil {
		c.EncodeTime = logf.RFC3339NanoTimeEncoder
	}
	if c.EncodeError == nil {
		c.EncodeError = logf.DefaultErrorEncoder
	}
	if c.EncodeLevel == nil {
		c.EncodeLevel = logf.DefaultLevelEncoder
	}
	if c.EncodeCaller == nil {
		c.EncodeCaller = logf.ShortCallerEncoder
	}

	return c
}
