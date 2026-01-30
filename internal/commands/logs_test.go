package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestLogsCmd_Success(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}}).
		WithLogAccess(&platform.LogAccess{AccessToken: "tok", URL: "http://logs"})

	fetcher := platform.NewMockLogFetcher().WithEntries([]platform.LogEntry{
		{Timestamp: "2026-01-28T14:30:00Z", Severity: "error", Message: "Connection refused", Container: "api-1"},
		{Timestamp: "2026-01-28T14:31:00Z", Severity: "info", Message: "Retrying...", Container: "api-1"},
	})

	cmd := NewLogs(storagePath, mock, fetcher)
	cmd.SetArgs([]string{"--service", "api"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	entries := data["entries"].([]interface{})
	if len(entries) != 2 {
		t.Errorf("entries len = %d, want 2", len(entries))
	}
	entry0 := entries[0].(map[string]interface{})
	if entry0["severity"] != "error" {
		t.Errorf("entries[0].severity = %v, want error", entry0["severity"])
	}
}

func TestLogsCmd_InvalidSince(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()
	fetcher := platform.NewMockLogFetcher()

	cmd := NewLogs(storagePath, mock, fetcher)
	cmd.SetArgs([]string{"--service", "api", "--since", "abc"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --since")
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "INVALID_PARAMETER" {
		t.Errorf("code = %v, want INVALID_PARAMETER", resp["code"])
	}
}

func TestParseSince(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"30_minutes", "30m", false},
		{"1_hour", "1h", false},
		{"24_hours", "24h", false},
		{"7_days", "7d", false},
		{"iso8601", "2026-01-28T14:30:00Z", false},
		{"invalid", "abc", true},
		{"too_many_minutes", "1500m", true},
		{"too_many_hours", "200h", true},
		{"too_many_days", "31d", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSince(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSince(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
