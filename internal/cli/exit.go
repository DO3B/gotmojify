// Package cli implements the gotmojify Cobra command tree, flags, and terminal I/O.
package cli

import "errors"

// ExitError maps command failures to specific POSIX process exit codes without calling os.Exit inside RunE.
type ExitError struct {
	Code     int
	Err      error
	IsSilent bool
}

// Error returns the wrapped error string or a fallback exit code description.
func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}

	return "command failed"
}

// Unwrap exposes the underlying error for errors.Is / errors.As inspection.
func (e *ExitError) Unwrap() error {
	return e.Err
}

func newSilentExitError(code int) error {
	return &ExitError{
		Code:     code,
		IsSilent: true,
		Err:      errors.New("command exited"),
	}
}
