package main

import (
	"runtime"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logfjournald"
)

func newLogger() (*logf.Logger, logf.ChannelCloser) {
	channel := logf.NewBasicChannel(logf.ChannelConfig{
		Appender: logfjournald.NewAppender(
			logfjournald.NewEncoder(logfjournald.EncoderConfig{}, logf.NewJSONTypeEncoderFactory(logf.JSONEncoderConfig{})),
		),
	})

	return logf.NewLogger(logf.LevelInfo.Checker(), channel), channel
}

func main() {
	logger, channel := newLogger()
	defer channel.Close()

	logger.Info("got cpu info", logf.Int("count", runtime.NumCPU()))
}
