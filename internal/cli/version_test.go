package cli_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DO3B/gotmojify/internal/cli"
)

func TestVersionCmd(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	stdout := new(bytes.Buffer)
	root := cli.NewRootCmd(cli.BuildInfo{
		Version: "1.2.3",
		Commit:  "abc1234",
		Date:    "2026-11-20T21:00:00Z",
	})
	root.SetOut(stdout)
	root.SetArgs([]string{"version"})

	must.NoError(root.Execute())
	is.Contains(stdout.String(), "gotmojify 1.2.3 (commit: abc1234, built: 2026-11-20T21:00:00Z)")
	is.Contains(stdout.String(), "go: go")
}
