package logfjournald

import (
	"testing"

	"github.com/ssgreg/logf"
	"github.com/stretchr/testify/require"
)

func TestAppender(t *testing.T) {
	baseApp := NewAppender(newTestEncoder())
	app, ok := baseApp.(*appender)
	require.True(t, ok)
	app.j.TestModeEnabled = true
	require.Empty(t, app.buf.Len())

	t.Run("AppendNotFlush", func(t *testing.T) {
		smallEntry := logf.Entry{}
		require.NoError(t, app.Append(smallEntry))
		require.True(t, app.buf.Len() > 0)
	})

	t.Run("Sync", func(t *testing.T) {
		require.NoError(t, app.Sync())
		require.Empty(t, app.buf.Len())
	})

	t.Run("FlushWithNoBuf", func(t *testing.T) {
		require.NoError(t, app.Flush())
		require.Empty(t, app.buf.Len())
	})

	t.Run("AppendAutoFlush", func(t *testing.T) {
		bigEntry := logf.Entry{Text: string(make([]byte, logf.PageSize))}
		require.NoError(t, app.Append(bigEntry))
		require.Empty(t, app.buf.Len())
	})

	t.Run("Close", func(t *testing.T) {
		require.NoError(t, app.Close())
		require.Empty(t, app.j)
	})
}
