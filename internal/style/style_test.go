package style_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DO3B/gotmojify/internal/style"
)

func TestStyler(t *testing.T) {
	t.Parallel()

	t.Run("disabled returns plain text", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		st := style.New(false)

		is.False(st.IsEnabled())
		is.Equal("feat", st.Blue("feat"))
		is.Equal("(cli)", st.BoldBlue("(cli)"))
		is.Equal("Available Devmoji", st.BlueDimUnderline("Available Devmoji"))
		is.Equal("ok", st.Green("ok"))
		is.Equal("err", st.Red("err"))
		is.Equal("types", st.Grey("types"))
		is.Equal("desc", st.Dim("desc"))
	})

	t.Run("enabled wraps text in ANSI codes", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		st := style.New(true)

		is.True(st.IsEnabled())
		is.Contains(st.Blue("feat"), "\x1b[")
		is.Contains(st.Blue("feat"), "feat")
		is.Contains(st.BoldBlue("(cli)"), "\x1b[")
		is.Contains(st.BoldBlue("(cli)"), "(cli)")
		is.Contains(st.BlueDimUnderline("Available Devmoji"), "\x1b[")
		is.Contains(st.Green("ok"), "\x1b[")
		is.Contains(st.Red("err"), "\x1b[")
		is.Contains(st.Grey("types"), "\x1b[")
		is.Contains(st.Dim("desc"), "\x1b[")
	})
}
