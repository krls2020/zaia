package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zeropsio/zaia/internal/platform"
)

func TestSync_BasicData(t *testing.T) {
	var buf bytes.Buffer
	err := SyncTo(&buf, map[string]interface{}{
		"message": "hello",
	})
	if err != nil {
		t.Fatal(err)
	}

	var resp SyncResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp.Type != "sync" {
		t.Errorf("type = %q, want sync", resp.Type)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want ok", resp.Status)
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("data is not a map")
	}
	if data["message"] != "hello" {
		t.Errorf("data.message = %v, want hello", data["message"])
	}
}

func TestSync_NilData(t *testing.T) {
	var buf bytes.Buffer
	err := SyncTo(&buf, nil)
	if err != nil {
		t.Fatal(err)
	}

	var resp SyncResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp.Type != "sync" {
		t.Errorf("type = %q, want sync", resp.Type)
	}
	if resp.Data != nil {
		t.Errorf("data = %v, want nil", resp.Data)
	}
}

func TestSync_ComplexData(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"project": map[string]interface{}{
			"id":   "abc-123",
			"name": "my-app",
		},
		"services": []map[string]interface{}{
			{"hostname": "api", "type": "nodejs@22"},
			{"hostname": "db", "type": "postgresql@16"},
		},
	}
	err := SyncTo(&buf, data)
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync", resp["type"])
	}

	respData := resp["data"].(map[string]interface{})
	project := respData["project"].(map[string]interface{})
	if project["id"] != "abc-123" {
		t.Errorf("project.id = %v, want abc-123", project["id"])
	}
}

func TestAsync_SingleProcess(t *testing.T) {
	var buf bytes.Buffer
	err := AsyncTo(&buf, []ProcessOutput{
		{
			ProcessID:       "proc-uuid-1",
			ActionName:      "start",
			ServiceHostname: "api",
			ServiceID:       "svc-uuid-1",
			Status:          "PENDING",
			Created:         "2026-01-29T10:00:00Z",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resp AsyncResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if resp.Type != "async" {
		t.Errorf("type = %q, want async", resp.Type)
	}
	if resp.Status != "initiated" {
		t.Errorf("status = %q, want initiated", resp.Status)
	}
	if len(resp.Processes) != 1 {
		t.Fatalf("processes len = %d, want 1", len(resp.Processes))
	}
	if resp.Processes[0].ProcessID != "proc-uuid-1" {
		t.Errorf("processId = %q, want proc-uuid-1", resp.Processes[0].ProcessID)
	}
	if resp.Processes[0].ActionName != "start" {
		t.Errorf("actionName = %q, want start", resp.Processes[0].ActionName)
	}
}

func TestAsync_MultipleProcesses(t *testing.T) {
	var buf bytes.Buffer
	err := AsyncTo(&buf, []ProcessOutput{
		{ProcessID: "p1", ActionName: "import", ServiceHostname: "api", Status: "PENDING"},
		{ProcessID: "p2", ActionName: "import", ServiceHostname: "db", Status: "PENDING"},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resp AsyncResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Processes) != 2 {
		t.Errorf("processes len = %d, want 2", len(resp.Processes))
	}
}

func TestErr_BasicError(t *testing.T) {
	var buf bytes.Buffer
	err := ErrTo(&buf, "AUTH_REQUIRED", "Not authenticated", "Run: zaia login <token>", nil)

	// Check the returned error
	zaiaErr, ok := err.(*ZaiaError)
	if !ok {
		t.Fatal("returned error is not *ZaiaError")
	}
	if zaiaErr.Code != "AUTH_REQUIRED" {
		t.Errorf("error code = %q, want AUTH_REQUIRED", zaiaErr.Code)
	}
	if zaiaErr.Error() != "Not authenticated" {
		t.Errorf("error message = %q, want 'Not authenticated'", zaiaErr.Error())
	}

	// Check JSON output
	var resp ErrorResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Type != "error" {
		t.Errorf("type = %q, want error", resp.Type)
	}
	if resp.Code != "AUTH_REQUIRED" {
		t.Errorf("code = %q, want AUTH_REQUIRED", resp.Code)
	}
	if resp.Error != "Not authenticated" {
		t.Errorf("error = %q, want 'Not authenticated'", resp.Error)
	}
	if resp.Suggestion != "Run: zaia login <token>" {
		t.Errorf("suggestion = %q, want 'Run: zaia login <token>'", resp.Suggestion)
	}
}

func TestErr_WithContext(t *testing.T) {
	var buf bytes.Buffer
	ctx := map[string]interface{}{
		"projectId":        "abc-123",
		"requestedService": "apistage",
	}
	_ = ErrTo(&buf, "SERVICE_NOT_FOUND", "Service 'apistage' not found", "Available services: api, db", ctx)

	var resp ErrorResponse
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Context == nil {
		t.Fatal("context is nil")
	}
}

func TestErr_NoSuggestionOmitted(t *testing.T) {
	var buf bytes.Buffer
	_ = ErrTo(&buf, "API_ERROR", "Internal error", "", nil)

	var raw map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}

	if _, ok := raw["suggestion"]; ok {
		t.Error("suggestion should be omitted when empty")
	}
	if _, ok := raw["context"]; ok {
		t.Error("context should be omitted when nil")
	}
}

func TestMapProcessToOutput_StatusMapping(t *testing.T) {
	tests := []struct {
		name       string
		apiStatus  string
		wantStatus string
	}{
		{"pending", "PENDING", "PENDING"},
		{"running", "RUNNING", "RUNNING"},
		{"done_to_finished", "DONE", "FINISHED"},
		{"failed", "FAILED", "FAILED"},
		{"cancelled_to_canceled", "CANCELLED", "CANCELED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &platform.Process{
				ID:         "proc-1",
				ActionName: "test",
				Status:     tt.apiStatus,
			}
			out := MapProcessToOutput(p, "")
			if out.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", out.Status, tt.wantStatus)
			}
		})
	}
}

func TestMapProcessToOutput_ServiceInfo(t *testing.T) {
	p := &platform.Process{
		ID:         "proc-1",
		ActionName: "restart",
		Status:     "PENDING",
		ServiceStacks: []platform.ServiceStackRef{
			{ID: "svc-1", Name: "api"},
		},
		Created: "2026-01-29T10:00:00Z",
	}

	out := MapProcessToOutput(p, "")
	if out.ServiceHostname != "api" {
		t.Errorf("serviceHostname = %q, want api", out.ServiceHostname)
	}
	if out.ServiceID != "svc-1" {
		t.Errorf("serviceId = %q, want svc-1", out.ServiceID)
	}
}

func TestMapProcessToOutput_HostnameOverride(t *testing.T) {
	p := &platform.Process{
		ID:         "proc-1",
		ActionName: "restart",
		Status:     "PENDING",
		ServiceStacks: []platform.ServiceStackRef{
			{ID: "svc-1", Name: "api"},
		},
	}

	out := MapProcessToOutput(p, "custom-hostname")
	if out.ServiceHostname != "custom-hostname" {
		t.Errorf("serviceHostname = %q, want custom-hostname", out.ServiceHostname)
	}
}

func TestMapProcessToOutput_WithFailure(t *testing.T) {
	reason := "Health check failed"
	finished := "2026-01-29T10:00:12Z"
	p := &platform.Process{
		ID:         "proc-1",
		ActionName: "restart",
		Status:     "FAILED",
		Created:    "2026-01-29T10:00:00Z",
		Finished:   &finished,
		FailReason: &reason,
	}

	out := MapProcessToOutput(p, "api")
	if out.Status != "FAILED" {
		t.Errorf("status = %q, want FAILED", out.Status)
	}
	if out.FailureReason == nil || *out.FailureReason != "Health check failed" {
		t.Errorf("failureReason = %v, want 'Health check failed'", out.FailureReason)
	}
	if out.Finished == nil || *out.Finished != "2026-01-29T10:00:12Z" {
		t.Errorf("finished = %v, want '2026-01-29T10:00:12Z'", out.Finished)
	}
}

func TestZaiaError_ExitCodes(t *testing.T) {
	tests := []struct {
		code     string
		wantExit int
	}{
		{"AUTH_REQUIRED", 2},
		{"AUTH_INVALID_TOKEN", 2},
		{"AUTH_TOKEN_EXPIRED", 2},
		{"TOKEN_NO_PROJECT", 2},
		{"TOKEN_MULTI_PROJECT", 2},
		{"SERVICE_REQUIRED", 3},
		{"CONFIRM_REQUIRED", 3},
		{"INVALID_ZEROPS_YML", 3},
		{"INVALID_IMPORT_YML", 3},
		{"INVALID_SCALING", 3},
		{"INVALID_PARAMETER", 3},
		{"INVALID_ENV_FORMAT", 3},
		{"INVALID_HOSTNAME", 3},
		{"SERVICE_NOT_FOUND", 4},
		{"PROCESS_NOT_FOUND", 4},
		{"PERMISSION_DENIED", 5},
		{"NETWORK_ERROR", 6},
		{"API_ERROR", 1},
		{"API_TIMEOUT", 1},
		{"UNKNOWN_CODE", 1},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			e := &ZaiaError{Code: tt.code, Message: "test"}
			if got := e.ExitCode(); got != tt.wantExit {
				t.Errorf("ExitCode() = %d, want %d", got, tt.wantExit)
			}
		})
	}
}
