package main

import (
	"github.com/pkg/errors"
	"github.com/ssgreg/logf"
	"github.com/ssgreg/logfjournald"
)

func main() {
	c := logf.SetFormatterConfigDefaults(&logf.FormatterConfig{})
	a := logfjournald.NewAppender(logfjournald.NewEncoder(c, logf.NewJSONTypeMarshallerFactory(c)))

	e := logf.Entry{
		Text: "greg test",
		Fields: []logf.Field{
			logf.AnError("_error", errors.Wrap(errors.New("error"), "internal error")),
			logf.ConstInts("ints", []int{123, 2342, 234}),
			logf.ConstInts("ints", []int{123, 2342, 234}),
			logf.ConstBytes("bytes", []byte(`byte array`)),
		},
	}

	a.Append(e)
	a.Append(e)
	a.Flush()
}
