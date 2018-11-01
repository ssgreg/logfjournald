package logfjournald

import "github.com/ssgreg/logf"

const (
	DefaultFieldKeyLevel  = "LEVEL"
	DefaultFieldKeyTime   = "TS"
	DefaultFieldKeyName   = "LOGGER"
	DefaultFieldKeyCaller = "CALLER"
)

type EncoderConfig struct {
	FieldKeyTime   string
	FieldKeyLevel  string
	FieldKeyName   string
	FieldKeyCaller string

	DisableFieldTime   bool
	DisableFieldLevel  bool
	DisableFieldName   bool
	DisableFieldCaller bool

	EncodeTime     logf.TimeEncoder
	EncodeDuration logf.DurationEncoder
	EncodeError    logf.ErrorEncoder
	EncodeLevel    logf.LevelEncoder
	EncodeCaller   logf.CallerEncoder
}

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
		c.EncodeTime = logf.RFC3339TimeEncoder
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
