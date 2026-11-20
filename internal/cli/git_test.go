package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DO3B/gotmojify/internal/cli"
)

func TestFindGitDir(t *testing.T) {
	t.Parallel()

	t.Run("discovers standard .git directory in parent", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		root := t.TempDir()
		gitDir := filepath.Join(root, ".git")
		nested := filepath.Join(root, "a", "b")

		must.NoError(os.MkdirAll(gitDir, 0o750))
		must.NoError(os.MkdirAll(nested, 0o750))

		got, err := cli.FindGitDir(nested)
		must.NoError(err)
		is.Equal(gitDir, got)
	})

	t.Run("discovers worktree gitdir file pointer", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		must := require.New(t)

		root := t.TempDir()
		actualGitDir := filepath.Join(root, "actual-git")
		worktree := filepath.Join(root, "worktree")

		must.NoError(os.MkdirAll(actualGitDir, 0o750))
		must.NoError(os.MkdirAll(worktree, 0o750))
		must.NoError(os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: ../actual-git\n"), 0o600))

		got, err := cli.FindGitDir(worktree)
		must.NoError(err)
		is.Equal(actualGitDir, got)
	})

	t.Run("returns ErrGitDirNotFound when missing", func(t *testing.T) {
		t.Parallel()

		is := assert.New(t)
		_, err := cli.FindGitDir(t.TempDir())
		is.ErrorIs(err, cli.ErrGitDirNotFound)
	})
}

func TestReadAndWriteCommitEditMsg(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	must := require.New(t)

	gitDir := t.TempDir()
	must.NoError(cli.WriteCommitEditMsg(gitDir, "feat: ✨ initial commit\n"))

	got, err := cli.ReadCommitEditMsg(gitDir)
	must.NoError(err)
	is.Equal("feat: ✨ initial commit\n", got)
}
