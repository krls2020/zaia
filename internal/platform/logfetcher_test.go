package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogFetcher_FetchLogs_Success(t *testing.T) {
	response := logAPIResponse{
		Items: []logAPIItem{
			{ID: "1", Timestamp: "2025-01-01T00:00:02Z", Hostname: "api-0", Message: "second", SeverityLabel: "info"},
			{ID: "2", Timestamp: "2025-01-01T00:00:01Z", Hostname: "api-0", Message: "first", SeverityLabel: "error"},
		},
	}
	body, _ := json.Marshal(response)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("accessToken") != "test-token" {
			t.Error("missing accessToken query param")
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing Authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	fetcher := NewLogFetcher()
	entries, err := fetcher.FetchLogs(context.Background(), &LogAccess{
		URL:         srv.URL,
		AccessToken: "test-token",
	}, LogFetchParams{
		ServiceID: "svc-1",
		Limit:     100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Should be sorted chronologically (first before second).
	if entries[0].Message != "first" {
		t.Errorf("first entry = %s, want 'first'", entries[0].Message)
	}
	if entries[1].Message != "second" {
		t.Errorf("second entry = %s, want 'second'", entries[1].Message)
	}
	if entries[0].Severity != "error" {
		t.Errorf("severity = %s, want 'error'", entries[0].Severity)
	}
	if entries[0].Container != "api-0" {
		t.Errorf("container = %s, want 'api-0'", entries[0].Container)
	}
}

func TestLogFetcher_FetchLogs_QueryParams(t *testing.T) {
	var receivedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	fetcher := NewLogFetcher()
	since := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	_, err := fetcher.FetchLogs(context.Background(), &LogAccess{
		URL:         srv.URL,
		AccessToken: "tok",
	}, LogFetchParams{
		ServiceID: "svc-1",
		Severity:  "error",
		Since:     since,
		Limit:     50,
		Search:    "crash",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify all query params are present.
	for _, want := range []string{"serviceStackId=svc-1", "severity=error", "tail=50", "search=crash", "since=2025-01-15T10"} {
		if !containsString(receivedQuery, want) {
			t.Errorf("query missing %s, got: %s", want, receivedQuery)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestLogFetcher_FetchLogs_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	fetcher := NewLogFetcher()
	_, err := fetcher.FetchLogs(context.Background(), &LogAccess{
		URL:         srv.URL,
		AccessToken: "tok",
	}, LogFetchParams{Limit: 10})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestLogFetcher_FetchLogs_NilAccess(t *testing.T) {
	fetcher := NewLogFetcher()
	_, err := fetcher.FetchLogs(context.Background(), nil, LogFetchParams{})
	if err == nil {
		t.Fatal("expected error for nil access")
	}
}

func TestLogFetcher_FetchLogs_URLPrefix(t *testing.T) {
	// Test that "GET " prefix is stripped from URL.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	fetcher := NewLogFetcher()
	entries, err := fetcher.FetchLogs(context.Background(), &LogAccess{
		URL:         "GET " + srv.URL,
		AccessToken: "tok",
	}, LogFetchParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries == nil {
		t.Fatal("expected empty slice, got nil")
	}
}

func TestLogFetcher_FetchLogs_LimitApplied(t *testing.T) {
	// Return 5 items but request limit 2 — should truncate.
	items := make([]logAPIItem, 5)
	for i := range items {
		items[i] = logAPIItem{
			ID:            string(rune('1' + i)),
			Timestamp:     time.Date(2025, 1, 1, 0, 0, i, 0, time.UTC).Format(time.RFC3339),
			Message:       "msg",
			SeverityLabel: "info",
		}
	}
	response := logAPIResponse{Items: items}
	body, _ := json.Marshal(response)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	fetcher := NewLogFetcher()
	entries, err := fetcher.FetchLogs(context.Background(), &LogAccess{
		URL:         srv.URL,
		AccessToken: "tok",
	}, LogFetchParams{Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLogFetcher_SeverityAllNotSent(t *testing.T) {
	var receivedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	fetcher := NewLogFetcher()
	_, _ = fetcher.FetchLogs(context.Background(), &LogAccess{
		URL:         srv.URL,
		AccessToken: "tok",
	}, LogFetchParams{Severity: "all", Limit: 10})

	// "all" should not be sent as severity param
	if containsSubstring(receivedQuery, "severity=") {
		t.Errorf("severity=all should not be sent, got: %s", receivedQuery)
	}
}

func TestParseLogResponse(t *testing.T) {
	input := `{"items":[{"id":"1","timestamp":"2025-01-01T00:00:00Z","hostname":"web-0","message":"hello","severityLabel":"info"}]}`
	entries, err := parseLogResponse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "1" {
		t.Errorf("ID = %s, want 1", entries[0].ID)
	}
	if entries[0].Container != "web-0" {
		t.Errorf("Container = %s, want web-0", entries[0].Container)
	}
}

func TestParseLogResponse_InvalidJSON(t *testing.T) {
	_, err := parseLogResponse([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseLogResponse_EmptyItems(t *testing.T) {
	entries, err := parseLogResponse([]byte(`{"items":[]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
