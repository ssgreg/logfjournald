package main

import (
	"runtime"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logfjournald"
)

func main() {
	// Create journald Appender with default journald Encoder.
	appender, appenderClose := logfjournald.NewAppender(logfjournald.NewEncoder.Default())
	defer appenderClose()

	// Create ChannelWriter with journald Encoder.
	writer, writerClose := logf.NewChannelWriter(logf.ChannelWriterConfig{
		Appender: appender,
	})
	defer writerClose()

	// Create Logger with ChannelWriter.
	logger := logf.NewLogger(logf.LevelInfo, writer)

	logger.Info("got cpu info", logf.Int("count", runtime.NumCPU()))
}
