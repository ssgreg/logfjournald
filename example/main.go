package main

import (
	"os"
	"runtime"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logfjournald"
)

func newLogger() (*logf.Logger, logf.ChannelCloser) {
	c := logf.SetFormatterConfigDefaults(&logf.FormatterConfig{})

	channel := logf.NewBasicChannel(logf.ChannelConfig{
		Appender:      logfjournald.NewAppender(logfjournald.NewEncoder(c, logf.NewJSONTypeMarshallerFactory(c))),
		ErrorAppender: logf.NewWriteAppender(os.Stderr, logf.NewJSONEncoder(c)),
	})

	return logf.NewLogger(logf.LevelInfo.Checker(), channel), channel
}

func main() {
	logger, channel := newLogger()
	defer channel.Close()

	logger.Info("got cpu info", logf.Int("count", runtime.NumCPU()))
}
