package integration

import (
	"testing"

	"github.com/zeropsio/zaia/internal/platform"
)

func TestFlow_LogsFetchEntries(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	h.Mock().WithLogAccess(&platform.LogAccess{
		URL:         "https://log.zerops.io/logs",
		AccessToken: "log-token-123",
	})

	mockFetcher := platform.NewMockLogFetcher().WithEntries([]platform.LogEntry{
		{ID: "1", Timestamp: "2025-01-01T00:00:01Z", Severity: "info", Message: "app started", Container: "api-0"},
		{ID: "2", Timestamp: "2025-01-01T00:00:02Z", Severity: "error", Message: "connection refused", Container: "api-0"},
		{ID: "3", Timestamp: "2025-01-01T00:00:03Z", Severity: "info", Message: "retrying", Container: "api-1"},
	})
	h.SetLogFetcher(mockFetcher)

	r := h.MustRun("logs --service api")
	r.AssertType("sync")

	data := r.Data()
	entries, ok := data["entries"].([]interface{})
	if !ok {
		t.Fatalf("expected entries array, got %T", data["entries"])
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	first := entries[0].(map[string]interface{})
	if first["message"] != "app started" {
		t.Errorf("first message = %v, want 'app started'", first["message"])
	}
	if first["severity"] != "info" {
		t.Errorf("first severity = %v, want 'info'", first["severity"])
	}
	if first["container"] != "api-0" {
		t.Errorf("first container = %v, want 'api-0'", first["container"])
	}

	second := entries[1].(map[string]interface{})
	if second["severity"] != "error" {
		t.Errorf("second severity = %v, want 'error'", second["severity"])
	}
}

func TestFlow_LogsRequiresService(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("logs")
	r.AssertType("error")
	r.AssertErrorCode("SERVICE_REQUIRED")
}

func TestFlow_LogsServiceNotFound(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("logs --service nonexistent")
	r.AssertType("error")
	r.AssertErrorCode("SERVICE_NOT_FOUND")
}

func TestFlow_LogsUnauthenticated(t *testing.T) {
	h := NewHarness(t)
	FixtureUnauthenticated(h)

	r := h.Run("logs --service api")
	r.AssertType("error")
	r.AssertErrorCode("AUTH_REQUIRED")
	r.AssertExitCode(2)
}

func TestFlow_LogsEmptyResult(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	h.Mock().WithLogAccess(&platform.LogAccess{
		URL:         "https://log.zerops.io/logs",
		AccessToken: "log-token-123",
	})

	mockFetcher := platform.NewMockLogFetcher().WithEntries([]platform.LogEntry{})
	h.SetLogFetcher(mockFetcher)

	r := h.MustRun("logs --service api")
	r.AssertType("sync")

	data := r.Data()
	entries, ok := data["entries"].([]interface{})
	if !ok {
		t.Fatalf("expected entries array, got %T", data["entries"])
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestFlow_LogsFetchError(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	h.Mock().WithLogAccess(&platform.LogAccess{
		URL:         "https://log.zerops.io/logs",
		AccessToken: "log-token-123",
	})

	mockFetcher := platform.NewMockLogFetcher().WithError(
		platform.NewPlatformError(platform.ErrAPIError, "log backend unreachable", ""),
	)
	h.SetLogFetcher(mockFetcher)

	r := h.Run("logs --service api")
	r.AssertType("error")
	r.AssertErrorCode("API_ERROR")
}
