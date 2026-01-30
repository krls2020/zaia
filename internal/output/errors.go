package output

import (
	"github.com/zeropsio/zaia/internal/platform"
)

// ZaiaError is the base error type returned by Err().
type ZaiaError struct {
	Message string `json:"error"`
	Code    string `json:"code"`
}

func (e *ZaiaError) Error() string {
	return e.Message
}

// ExitCode returns the process exit code for this error.
func (e *ZaiaError) ExitCode() int {
	return platform.ExitCodeForError(e.Code)
}
