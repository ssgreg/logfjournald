package logfjournald

import (
	"testing"

	"github.com/ssgreg/logf"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	Name    string
	Testing string
	Golden  string
}

func TestAppendNormalizedKey(t *testing.T) {

	testCases := []testCase{
		{"NoChanges", "ABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789", "ABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789"},
		{"StartWithUnderscore", "_F", "LOGF_F"},
		{"StartWithInvalidChar", "!F", "LOGF_F"},
		{"Uppercasing", "abcdefghijklmnopqrstuvwxyz", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"InvalidChars", "!@#$%^&*()-=\n\r\a<>?ГЗ:;'\\|?.,~[]{}", "LOGF_________________________________"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			b := logf.NewBuffer()
			appendNormalizedKey(b, tc.Testing)

			require.EqualValues(t, tc.Golden, b.String())
		})
	}
}
