package logfjournald

import (
	"github.com/ssgreg/journald"
	"github.com/ssgreg/logf"
)

// NewAppender creates the new instance of journal appender with the given
// Encoder.
func NewAppender(enc logf.Encoder) logf.Appender {
	return &appender{
		j:   &journald.Journal{},
		enc: enc,
		buf: logf.NewBufferWithCapacity(logf.PageSize * 2),
	}
}

type appender struct {
	j   *journald.Journal
	enc logf.Encoder
	buf *logf.Buffer
}

func (a *appender) Append(entry logf.Entry) error {
	err := a.enc.Encode(a.buf, entry)
	if err != nil {
		return err
	}
	if a.buf.Len() > logf.PageSize {
		a.Flush()
	}

	return nil
}

func (a *appender) Sync() (err error) {
	return a.Flush()
}

func (a *appender) Flush() error {
	if a.buf.Len() != 0 {
		defer a.buf.Reset()

		return a.j.WriteMsg(a.buf.Bytes())
	}

	return nil
}

func (a *appender) Close() error {
	defer func() {
		a.j = nil
	}()

	return a.j.Close()
}
