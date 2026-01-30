package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/setup"
)

func TestSetup_Success(t *testing.T) {
	dir := t.TempDir()
	binPath := dir + "/zaia-mcp"
	configPath := dir + "/.claude.json"

	downloader := &mockSetupDownloader{data: []byte("binary-data")}
	cmd := NewSetupWithDeps(downloader, binPath, configPath)

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	if data["binaryPath"] != binPath {
		t.Errorf("binaryPath = %v, want %v", data["binaryPath"], binPath)
	}
	if data["configPath"] != configPath {
		t.Errorf("configPath = %v, want %v", data["configPath"], configPath)
	}
	if data["configUpdated"] != true {
		t.Errorf("configUpdated = %v, want true", data["configUpdated"])
	}
}

func TestSetup_DownloadFails(t *testing.T) {
	dir := t.TempDir()
	binPath := dir + "/zaia-mcp"
	configPath := dir + "/.claude.json"

	downloader := &mockSetupDownloader{err: "download failed"}
	cmd := NewSetupWithDeps(downloader, binPath, configPath)

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "SETUP_DOWNLOAD_FAILED" {
		t.Errorf("code = %v, want SETUP_DOWNLOAD_FAILED", resp["code"])
	}
}

func TestSetup_ConfigWriteFails(t *testing.T) {
	dir := t.TempDir()
	binPath := dir + "/zaia-mcp"
	// Use invalid path (directory) as config to trigger write error
	configPath := dir // dir exists as a directory, not a file — writing will fail

	downloader := &mockSetupDownloader{data: []byte("binary")}
	cmd := NewSetupWithDeps(downloader, binPath, configPath)

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for config write failure")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "SETUP_CONFIG_FAILED" {
		t.Errorf("code = %v, want SETUP_CONFIG_FAILED", resp["code"])
	}
}

func TestSetup_Idempotent(t *testing.T) {
	dir := t.TempDir()
	binPath := dir + "/zaia-mcp"
	configPath := dir + "/.claude.json"

	downloader := &mockSetupDownloader{data: []byte("binary-data")}

	// Run twice
	for i := 0; i < 2; i++ {
		cmd := NewSetupWithDeps(downloader, binPath, configPath)
		var stdout bytes.Buffer
		output.SetWriter(&stdout)

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("run %d: unexpected error: %v", i+1, err)
		}

		output.ResetWriter()
	}
}

func TestSetup_Help(t *testing.T) {
	cmd := NewSetup()
	if cmd.Use != "setup" {
		t.Errorf("Use = %q, want 'setup'", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
}

// mockSetupDownloader implements setup.Downloader for command tests.
type mockSetupDownloader struct {
	data []byte
	err  string
}

func (m *mockSetupDownloader) Download(url string) ([]byte, error) {
	if m.err != "" {
		return nil, &mockDownloadError{msg: m.err}
	}
	return m.data, nil
}

// Ensure interface compliance
var _ setup.Downloader = &mockSetupDownloader{}

type mockDownloadError struct{ msg string }

func (e *mockDownloadError) Error() string { return e.msg }
