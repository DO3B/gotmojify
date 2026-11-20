package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DO3B/gotmojify/internal/cli"
)

const (
	flagText     = "--text"
	flagFormat   = "--format"
	flagNoCommit = "--no-commit"
)

func runCLI(
	t *testing.T,
	workDir string,
	stdin string,
	stdinTTY bool,
	stdoutTTY bool,
	args ...string,
) (string, string, error) {
	t.Helper()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	root := cli.NewRootCmd(
		cli.BuildInfo{Version: "0.1.0", Commit: "test", Date: "2026-11-20"},
		cli.WithWorkDir(workDir),
		cli.WithStdinTTY(stdinTTY),
		cli.WithStdoutTTY(stdoutTTY),
	)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)

	err := root.ExecuteContext(t.Context())

	return stdout.String(), stderr.String(), err
}

func TestRootCmd_HelpAndVersion(t *testing.T) {
	t.Parallel()

	t.Run("help flag prints usage", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		out, _, err := runCLI(t, t.TempDir(), "", true, false, "--help")
		must.NoError(err)
		is.Contains(out, "Usage:")
		is.Contains(out, "--format")
	})

	t.Run("version flag prints version", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		out, _, err := runCLI(t, t.TempDir(), "", true, false, "--version")
		must.NoError(err)
		is.Contains(out, "gotmojify version 0.1.0")
	})

	t.Run("tty with no mode flags prints help and exits 1", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)

		out, _, err := runCLI(t, t.TempDir(), "", true, false)
		is.Contains(out, "Usage:")

		exitErr, ok := errors.AsType[*cli.ExitError](err)
		is.True(ok)
		is.Equal(1, exitErr.Code)
	})
}

func TestRootCmd_List(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	out, _, err := runCLI(t, t.TempDir(), "", true, false, "--list", "--no-color")
	must.NoError(err)
	is.Contains(out, "Available Devmoji")
	is.Contains(out, ":feat:")
	is.Contains(out, "feat: a new feature")
	is.Contains(out, ":chore-release:")
	is.Contains(out, "chore(release):")
}

func TestRootCmd_TextFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "default commit and unicode",
			args:     []string{flagText, "feat: added a new feature :smile:"},
			expected: "feat: ✨ added a new feature 😄\n",
		},
		{
			name:     "shortcode format without commit header processing",
			args:     []string{flagText, "This is a :test: of the first :release: :boom: ✨", flagNoCommit, flagFormat, "shortcode"},
			expected: "This is a :rotating_light: of the first :rocket: :boom: :sparkles:\n",
		},
		{
			name:     "unicode format without commit",
			args:     []string{flagText, "This is a :test: of the first :release: :boom: :sparkles:", flagNoCommit, flagFormat, "unicode"},
			expected: "This is a 🚨 of the first 🚀 💥 ✨\n",
		},
		{
			name:     "devmoji format without commit",
			args:     []string{flagText, "🚀 :boom: :sparkles:", flagNoCommit, flagFormat, "devmoji"},
			expected: ":chore-release: :breaking: :feat:\n",
		},
		{
			name:     "strip format without commit",
			args:     []string{flagText, "test 🚀 :boom: :sparkles: :security:", flagNoCommit, flagFormat, "strip"},
			expected: "test\n",
		},
		{
			name:     "chore release commit",
			args:     []string{flagText, "chore(release): 1.1.1"},
			expected: "chore(release): 🚀 1.1.1\n",
		},
		{
			name:     "fix security commit",
			args:     []string{flagText, "fix(security): upgraded lodash"},
			expected: "fix(security): 🐛 🔒️ upgraded lodash\n",
		},
		{
			name:     "log flag formats multiple lines",
			args:     []string{flagText, "abc feat: hello\ndef fix: world", "--log", "--no-color"},
			expected: "abc feat: ✨ hello\ndef fix: 🐛 world\n",
		},
		{
			name:    "invalid format errors",
			args:    []string{flagText, "hello", flagFormat, "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			must := require.New(t)

			out, _, err := runCLI(t, t.TempDir(), "", true, false, tt.args...)
			if tt.wantErr {
				is.Error(err)
				return
			}

			must.NoError(err)
			is.Equal(tt.expected, out)
		})
	}
}

func TestRootCmd_Lint(t *testing.T) {
	t.Parallel()

	t.Run("valid commit passes lint", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		out, errOut, err := runCLI(t, t.TempDir(), "", true, false, flagText, "feat(cli): valid commit", "--lint")
		must.NoError(err)
		is.Equal("feat(cli): ✨ valid commit\n", out)
		is.Empty(errOut)
	})

	t.Run("invalid commit fails lint with exit 1", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)

		_, errOut, err := runCLI(t, t.TempDir(), "", true, false, flagText, "Invalid: commit", "--lint", "--no-color")
		exitErr, ok := errors.AsType[*cli.ExitError](err)
		is.True(ok)
		is.Equal(1, exitErr.Code)
		is.Contains(errOut, "Type should be one of")
	})
}

func TestRootCmd_Edit(t *testing.T) {
	t.Parallel()

	t.Run("formats COMMIT_EDITMSG in git root", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		dir := t.TempDir()
		gitDir := filepath.Join(dir, ".git")
		must.NoError(os.MkdirAll(gitDir, 0o750))

		msgPath := filepath.Join(gitDir, "COMMIT_EDITMSG")
		must.NoError(os.WriteFile(msgPath, []byte("feat(cli): add edit mode\n"), 0o600))

		out, _, err := runCLI(t, dir, "", true, false, "--edit", "--lint", "--no-color")
		must.NoError(err)
		is.Contains(out, "✔ feat(cli): ✨ add edit mode")

		updated, err := cli.ReadCommitEditMsg(gitDir)
		must.NoError(err)
		is.Equal("feat(cli): ✨ add edit mode\n", updated)
	})

	t.Run("lint failure leaves COMMIT_EDITMSG untouched", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		dir := t.TempDir()
		gitDir := filepath.Join(dir, ".git")
		must.NoError(os.MkdirAll(gitDir, 0o750))

		msgPath := filepath.Join(gitDir, "COMMIT_EDITMSG")
		must.NoError(os.WriteFile(msgPath, []byte("bad commit message\n"), 0o600))

		_, errOut, err := runCLI(t, dir, "", true, false, "--edit", "--lint", "--no-color")
		exitErr, ok := errors.AsType[*cli.ExitError](err)
		is.True(ok)
		is.Equal(1, exitErr.Code)
		is.Contains(errOut, "Expecting a commit message")

		updated, readErr := cli.ReadCommitEditMsg(gitDir)
		must.NoError(readErr)
		is.Equal("bad commit message\n", updated)
	})

	t.Run("missing COMMIT_EDITMSG returns error", func(t *testing.T) {
		t.Parallel()

		must := require.New(t)
		_, _, err := runCLI(t, t.TempDir(), "", true, false, "--edit")
		must.Error(err)
		must.ErrorContains(err, "COMMIT_EDITMSG")
	})
}

func TestRootCmd_StdinPipe(t *testing.T) {
	t.Parallel()

	t.Run("formats commit header only on first line", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		stdin := "feat(cli): first :rocket:\nfix(body): second line\n"
		out, _, err := runCLI(t, t.TempDir(), stdin, false, false)
		must.NoError(err)
		is.Equal("feat(cli): ✨ first 🚀\nfix(body): second line\n", out)
	})

	t.Run("log mode formats every line", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		stdin := "* abc123 - feat(cli): first\n* def456 - fix(core): second\n"
		out, _, err := runCLI(t, t.TempDir(), stdin, false, false, "--log", "--no-color")
		must.NoError(err)
		is.Equal("* abc123 - feat(cli): ✨ first\n* def456 - fix(core): 🐛 second\n", out)
	})
}

func TestRootCmd_ConfigAndColor(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "custom.yaml")
	cfgContent := `types:
  - customtype
devmoji:
  - code: customtype
    emoji: star
    description: custom commit type
`
	must.NoError(os.WriteFile(cfgPath, []byte(cfgContent), 0o600))

	out, _, err := runCLI(
		t,
		dir,
		"",
		true,
		false,
		"--config",
		cfgPath,
		flagText,
		"customtype: works",
		"--lint",
		"--color",
	)
	must.NoError(err)
	is.Contains(out, "\x1b[")
	is.Contains(out, "⭐️ works")
}

func TestRootCmd_EnvPrecedence(t *testing.T) {
	is := assert.New(t)
	must := require.New(t)

	dir := t.TempDir()
	must.NoError(os.WriteFile(filepath.Join(dir, ".gotmojify.yaml"), []byte("format: devmoji\n"), 0o600))

	t.Setenv("GOTMOJIFY_FORMAT", "shortcode")

	out, _, err := runCLI(t, dir, "", true, false, flagText, "🚀 :feat:", flagNoCommit)
	must.NoError(err)
	is.Equal(":rocket: :sparkles:\n", out)

	// CLI flag overrides both env and config file.
	out, _, err = runCLI(t, dir, "", true, false, flagText, "🚀 :feat:", flagNoCommit, flagFormat, "devmoji")
	must.NoError(err)
	is.Equal(":chore-release: :feat:\n", out)

	// NO_COLOR disables TTY color output when --color is not explicitly passed.
	t.Setenv("NO_COLOR", "1")
	t.Setenv("GOTMOJIFY_FORMAT", "unicode")

	out, _, err = runCLI(t, dir, "", true, true, flagText, "feat: plain")
	must.NoError(err)
	is.NotContains(out, "\x1b[")
	is.Equal("feat: ✨ plain\n", out)
}
