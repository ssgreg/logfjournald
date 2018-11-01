package logfjournald

import (
	"unicode/utf8"

	"github.com/ssgreg/logf"
)

// appendNormalizedKey appends normalized key to the buf. The journal key
// name must be in uppercase and consist only of characters, numbers and
// underscores, and may not begin with an underscore.
func appendNormalizedKey(buf *logf.Buffer, s string) {
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case ((c >= 0x30 && c <= 0x39) || (c >= 0x41 && c <= 0x5a)):
			buf.AppendByte(c)
			i++
		case c >= 0x61 && c <= 0x7a:
			buf.AppendByte(c - 0x20)
			i++
		default:
			if i == 0 {
				buf.AppendString("LOGF")
			}
			buf.AppendByte('_')
			_, wd := utf8.DecodeRuneInString(s[i:])
			i += wd
		}
	}
}
