package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const commitEditMsgFile = "COMMIT_EDITMSG"

// ErrGitDirNotFound indicates no .git directory or gitdir pointer was found in parent directories.
var ErrGitDirNotFound = errors.New("cli: .git directory not found")

// FindGitDir walks upward from cwd to find the repository's git directory, supporting worktree .git files.
func FindGitDir(cwd string) (string, error) {
	dir := filepath.Clean(cwd)
	for {
		candidate := filepath.Join(dir, ".git")

		info, err := os.Stat(candidate)
		if err == nil {
			if info.IsDir() {
				return candidate, nil
			}

			resolved, resolveErr := resolveGitDirFile(dir, candidate)
			if resolveErr != nil {
				return "", resolveErr
			}

			return resolved, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrGitDirNotFound
		}

		dir = parent
	}
}

func resolveGitDirFile(baseDir, gitFile string) (string, error) {
	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return "", fmt.Errorf("opening directory %s: %w", baseDir, err)
	}

	defer func() {
		_ = root.Close()
	}()

	raw, err := root.ReadFile(filepath.Base(gitFile))
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", gitFile, err)
	}

	line := strings.TrimSpace(string(raw))

	target, ok := strings.CutPrefix(line, "gitdir:")
	if !ok {
		return "", fmt.Errorf("%w: invalid .git file format", ErrGitDirNotFound)
	}

	target = strings.TrimSpace(target)
	if !filepath.IsAbs(target) {
		target = filepath.Join(baseDir, target)
	}

	return filepath.Clean(target), nil
}

// ReadCommitEditMsg reads COMMIT_EDITMSG safely from gitDir using os.Root.
func ReadCommitEditMsg(gitDir string) (string, error) {
	root, err := os.OpenRoot(gitDir)
	if err != nil {
		return "", fmt.Errorf("opening git directory: %w", err)
	}

	defer func() {
		_ = root.Close()
	}()

	data, err := root.ReadFile(commitEditMsgFile)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", commitEditMsgFile, err)
	}

	return string(data), nil
}

// WriteCommitEditMsg writes updated commit content safely to COMMIT_EDITMSG inside gitDir using os.Root.
func WriteCommitEditMsg(gitDir, content string) error {
	root, err := os.OpenRoot(gitDir)
	if err != nil {
		return fmt.Errorf("opening git directory: %w", err)
	}

	writeErr := root.WriteFile(commitEditMsgFile, []byte(content), 0o600)
	closeErr := root.Close()

	if writeErr != nil {
		return errors.Join(fmt.Errorf("writing %s: %w", commitEditMsgFile, writeErr), closeErr)
	}

	if closeErr != nil {
		return fmt.Errorf("closing git directory: %w", closeErr)
	}

	return nil
}
