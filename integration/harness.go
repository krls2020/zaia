package integration

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeropsio/zaia/internal/commands"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// Harness provides a test harness for integration testing ZAIA CLI commands.
// Each Run() creates a fresh root command (simulating a new CLI invocation).
type Harness struct {
	t           *testing.T
	storagePath string
	mock        *StatefulMock
	logFetcher  platform.LogFetcher
}

// NewHarness creates a new test harness with a StatefulMock and temp storage.
func NewHarness(t *testing.T) *Harness {
	t.Helper()
	return &Harness{
		t:           t,
		storagePath: filepath.Join(t.TempDir(), "zaia.data"),
		mock:        NewStatefulMock(),
		logFetcher:  platform.NewMockLogFetcher(),
	}
}

// Mock returns the underlying StatefulMock for configuration.
func (h *Harness) Mock() *StatefulMock {
	return h.mock
}

// SetLogFetcher overrides the log fetcher used by the harness.
func (h *Harness) SetLogFetcher(f platform.LogFetcher) {
	h.logFetcher = f
}

// StoragePath returns the temp storage directory path.
func (h *Harness) StoragePath() string {
	return h.storagePath
}

// Run executes a CLI command string and returns the result.
// The command string is split on spaces (e.g. "discover --service api").
func (h *Harness) Run(cmdLine string) *Result {
	h.t.Helper()

	args := splitArgs(cmdLine)

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	// ClientFactory for login — returns the same mock for any token/apiHost
	clientFactory := func(token, apiHost string) platform.Client {
		return h.mock
	}

	root := commands.NewRootForTest(commands.RootDeps{
		StoragePath:   h.storagePath,
		Client:        h.mock,
		ClientFactory: clientFactory,
		LogFetcher:    h.logFetcher,
	})
	root.SetArgs(args)

	err := commands.Execute(root)

	return &Result{
		t:        h.t,
		stdout:   stdout.Bytes(),
		err:      err,
		exitCode: exitCodeFromErr(err),
	}
}

// MustRun executes a command and fails the test if it returns an error.
func (h *Harness) MustRun(cmdLine string) *Result {
	h.t.Helper()
	r := h.Run(cmdLine)
	if r.err != nil && r.exitCode != 0 {
		// Check if it's an expected error response (JSON error envelope)
		if r.Type() == "error" {
			h.t.Fatalf("MustRun(%q) failed: code=%s error=%s", cmdLine, r.ErrorCode(), r.ErrorMessage())
		}
		h.t.Fatalf("MustRun(%q) failed: %v\nstdout: %s", cmdLine, r.err, string(r.stdout))
	}
	return r
}

// Result captures the output of a CLI command execution.
type Result struct {
	t        *testing.T
	stdout   []byte
	err      error
	exitCode int

	parsed map[string]interface{} // lazy-parsed JSON
}

// ExitCode returns the process exit code.
func (r *Result) ExitCode() int {
	return r.exitCode
}

// Err returns the raw error (nil if command succeeded).
func (r *Result) Err() error {
	return r.err
}

// Raw returns the raw stdout bytes.
func (r *Result) Raw() []byte {
	return r.stdout
}

// JSON returns the parsed JSON response.
func (r *Result) JSON() map[string]interface{} {
	r.t.Helper()
	if r.parsed == nil {
		r.parsed = make(map[string]interface{})
		if len(r.stdout) > 0 {
			if err := json.Unmarshal(r.stdout, &r.parsed); err != nil {
				r.t.Fatalf("Failed to parse JSON output: %v\nraw: %s", err, string(r.stdout))
			}
		}
	}
	return r.parsed
}

// Type returns the "type" field from the JSON response ("sync", "async", or "error").
func (r *Result) Type() string {
	r.t.Helper()
	j := r.JSON()
	if v, ok := j["type"].(string); ok {
		return v
	}
	return ""
}

// Data returns the "data" field from a sync response.
func (r *Result) Data() map[string]interface{} {
	r.t.Helper()
	j := r.JSON()
	if d, ok := j["data"].(map[string]interface{}); ok {
		return d
	}
	return nil
}

// Processes returns the "processes" array from an async response.
func (r *Result) Processes() []interface{} {
	r.t.Helper()
	j := r.JSON()
	if p, ok := j["processes"].([]interface{}); ok {
		return p
	}
	return nil
}

// ErrorCode returns the "code" field from an error response.
func (r *Result) ErrorCode() string {
	r.t.Helper()
	j := r.JSON()
	if c, ok := j["code"].(string); ok {
		return c
	}
	return ""
}

// ErrorMessage returns the "error" field from an error response.
func (r *Result) ErrorMessage() string {
	r.t.Helper()
	j := r.JSON()
	if e, ok := j["error"].(string); ok {
		return e
	}
	return ""
}

// AssertType asserts the response type matches expected.
func (r *Result) AssertType(expected string) {
	r.t.Helper()
	got := r.Type()
	if got != expected {
		r.t.Errorf("expected type %q, got %q\nraw: %s", expected, got, string(r.stdout))
	}
}

// AssertErrorCode asserts the error code matches expected.
func (r *Result) AssertErrorCode(expected string) {
	r.t.Helper()
	got := r.ErrorCode()
	if got != expected {
		r.t.Errorf("expected error code %q, got %q\nraw: %s", expected, got, string(r.stdout))
	}
}

// AssertExitCode asserts the exit code matches expected.
func (r *Result) AssertExitCode(expected int) {
	r.t.Helper()
	if r.exitCode != expected {
		r.t.Errorf("expected exit code %d, got %d\nraw: %s", expected, r.exitCode, string(r.stdout))
	}
}

// splitArgs splits a command string into arguments, respecting single quotes.
func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	for _, ch := range s {
		switch {
		case ch == '\'' && !inQuote:
			inQuote = true
		case ch == '\'' && inQuote:
			inQuote = false
		case ch == ' ' && !inQuote:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// exitCodeFromErr extracts the exit code from a ZaiaError, or returns 0/1.
func exitCodeFromErr(err error) int {
	if err == nil {
		return 0
	}
	if ec, ok := err.(interface{ ExitCode() int }); ok {
		return ec.ExitCode()
	}
	return 1
}
