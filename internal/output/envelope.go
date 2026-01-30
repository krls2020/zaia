package output

import (
	"encoding/json"
	"io"
	"os"

	"github.com/zeropsio/zaia/internal/platform"
)

// SyncResponse represents an immediate response with data.
type SyncResponse struct {
	Type   string      `json:"type"`   // always "sync"
	Status string      `json:"status"` // always "ok"
	Data   interface{} `json:"data"`
}

// AsyncResponse represents an initiated operation with processes.
type AsyncResponse struct {
	Type      string          `json:"type"`   // always "async"
	Status    string          `json:"status"` // always "initiated"
	Processes []ProcessOutput `json:"processes"`
}

// ErrorResponse represents an error.
type ErrorResponse struct {
	Type       string      `json:"type"` // always "error"
	Code       string      `json:"code"`
	Error      string      `json:"error"`
	Suggestion string      `json:"suggestion,omitempty"`
	Context    interface{} `json:"context,omitempty"`
}

// ProcessOutput represents an async process in the output envelope.
type ProcessOutput struct {
	ProcessID       string  `json:"processId"`
	ActionName      string  `json:"actionName"`
	ServiceHostname string  `json:"serviceHostname,omitempty"`
	ServiceID       string  `json:"serviceId,omitempty"`
	Status          string  `json:"status"`
	Created         string  `json:"created,omitempty"`
	Finished        *string `json:"finished,omitempty"`
	FailureReason   *string `json:"failureReason,omitempty"`
}

// writer is the output destination. Defaults to os.Stdout.
// Can be overridden for testing.
var writer io.Writer = os.Stdout

// SetWriter sets the output writer (for testing).
func SetWriter(w io.Writer) {
	writer = w
}

// ResetWriter resets the output writer to os.Stdout.
func ResetWriter() {
	writer = os.Stdout
}

// Sync outputs a sync response to stdout.
func Sync(data interface{}) error {
	return writeJSON(SyncResponse{
		Type:   "sync",
		Status: "ok",
		Data:   data,
	})
}

// SyncTo writes a sync response to a specific writer.
func SyncTo(w io.Writer, data interface{}) error {
	return writeJSONTo(w, SyncResponse{
		Type:   "sync",
		Status: "ok",
		Data:   data,
	})
}

// Async outputs an async response to stdout.
func Async(processes []ProcessOutput) error {
	return writeJSON(AsyncResponse{
		Type:      "async",
		Status:    "initiated",
		Processes: processes,
	})
}

// AsyncTo writes an async response to a specific writer.
func AsyncTo(w io.Writer, processes []ProcessOutput) error {
	return writeJSONTo(w, AsyncResponse{
		Type:      "async",
		Status:    "initiated",
		Processes: processes,
	})
}

// Err outputs an error response to stdout and returns a ZaiaError.
func Err(code, message, suggestion string, ctx interface{}) error {
	resp := ErrorResponse{
		Type:       "error",
		Code:       code,
		Error:      message,
		Suggestion: suggestion,
		Context:    ctx,
	}
	_ = writeJSON(resp)
	return &ZaiaError{Message: message, Code: code}
}

// ErrTo writes an error response to a specific writer and returns a ZaiaError.
func ErrTo(w io.Writer, code, message, suggestion string, ctx interface{}) error {
	resp := ErrorResponse{
		Type:       "error",
		Code:       code,
		Error:      message,
		Suggestion: suggestion,
		Context:    ctx,
	}
	_ = writeJSONTo(w, resp)
	return &ZaiaError{Message: message, Code: code}
}

// MapProcessToOutput converts a platform.Process to an output ProcessOutput.
// Maps API status names to ZAIA status names.
func MapProcessToOutput(p *platform.Process, hostname string) ProcessOutput {
	out := ProcessOutput{
		ProcessID:     p.ID,
		ActionName:    p.ActionName,
		Status:        mapProcessStatus(p.Status),
		Created:       p.Created,
		Finished:      p.Finished,
		FailureReason: p.FailReason,
	}

	// Use hostname from process service stacks if available
	if hostname != "" {
		out.ServiceHostname = hostname
	}
	if len(p.ServiceStacks) > 0 {
		if out.ServiceHostname == "" {
			out.ServiceHostname = p.ServiceStacks[0].Name
		}
		out.ServiceID = p.ServiceStacks[0].ID
	}

	return out
}

// mapProcessStatus maps API process status to ZAIA output status.
func mapProcessStatus(apiStatus string) string {
	switch apiStatus {
	case "DONE":
		return "FINISHED"
	case "CANCELLED":
		return "CANCELED"
	default:
		return apiStatus // PENDING, RUNNING, FAILED pass through
	}
}

func writeJSON(v interface{}) error {
	return writeJSONTo(writer, v)
}

func writeJSONTo(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(v)
}
