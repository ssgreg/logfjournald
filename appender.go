package logfjournald

import (
	"github.com/ssgreg/journald"
	"github.com/ssgreg/logf"
)

// NewAppender creates an instance of journal appender.
func NewAppender(e logf.Encoder) logf.Appender {
	return &appender{
		j:   &journald.Journal{},
		enc: e,
		buf: logf.NewBuffer(),
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
	defer a.buf.Reset()

	return a.j.WriteMsg(a.buf.Bytes())
}

func (a *appender) Sync() (err error) {
	return a.Flush()
}

func (a *appender) Flush() error {
	return nil
}
