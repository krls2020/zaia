package commands

import (
	"io"

	"github.com/zeropsio/zaia/internal/output"
)

func getWriter() io.Writer {
	// We can't read the current writer, so we just return nil as a placeholder.
	// The real solution is to capture via the output package functions.
	return nil
}

func setWriter(w io.Writer) {
	if w == nil {
		output.ResetWriter()
	} else {
		output.SetWriter(w)
	}
}
