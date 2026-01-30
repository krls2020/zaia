package integration

import (
	"testing"
)

func TestEnvelope_SyncFormat(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.MustRun("discover")
	j := r.JSON()

	// Required fields for sync
	if j["type"] != "sync" {
		t.Errorf("expected type=sync, got %v", j["type"])
	}
	if j["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", j["status"])
	}
	if j["data"] == nil {
		t.Error("expected data field in sync response")
	}
}

func TestEnvelope_AsyncFormat(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.MustRun("start --service api")
	j := r.JSON()

	if j["type"] != "async" {
		t.Errorf("expected type=async, got %v", j["type"])
	}
	if j["status"] != "initiated" {
		t.Errorf("expected status=initiated, got %v", j["status"])
	}
	procs := j["processes"].([]interface{})
	if len(procs) == 0 {
		t.Fatal("expected processes array to be non-empty")
	}

	// Verify process output structure
	proc := procs[0].(map[string]interface{})
	requiredFields := []string{"processId", "actionName", "status"}
	for _, f := range requiredFields {
		if _, ok := proc[f]; !ok {
			t.Errorf("missing field %q in process output", f)
		}
	}
}

func TestEnvelope_ErrorFormat(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("discover --service nonexistent")
	j := r.JSON()

	if j["type"] != "error" {
		t.Errorf("expected type=error, got %v", j["type"])
	}
	if j["code"] == nil || j["code"] == "" {
		t.Error("expected non-empty code field in error response")
	}
	if j["error"] == nil || j["error"] == "" {
		t.Error("expected non-empty error field in error response")
	}
}

func TestEnvelope_ExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(h *Harness)
		cmdLine      string
		wantExitCode int
		wantCode     string
	}{
		{
			name:         "auth error → exit 2",
			setup:        func(h *Harness) { FixtureUnauthenticated(h) },
			cmdLine:      "discover",
			wantExitCode: 2,
			wantCode:     "AUTH_REQUIRED",
		},
		{
			name:         "validation error → exit 3",
			setup:        func(h *Harness) { FixtureFullProject(h) },
			cmdLine:      "env set --service api INVALID",
			wantExitCode: 3,
			wantCode:     "INVALID_ENV_FORMAT",
		},
		{
			name:         "not found → exit 4",
			setup:        func(h *Harness) { FixtureFullProject(h) },
			cmdLine:      "start --service ghost",
			wantExitCode: 4,
			wantCode:     "SERVICE_NOT_FOUND",
		},
		{
			name:         "process not found → exit 4",
			setup:        func(h *Harness) { FixtureFullProject(h) },
			cmdLine:      "process nonexistent",
			wantExitCode: 4,
			wantCode:     "PROCESS_NOT_FOUND",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHarness(t)
			tc.setup(h)

			r := h.Run(tc.cmdLine)
			r.AssertExitCode(tc.wantExitCode)
			r.AssertErrorCode(tc.wantCode)
		})
	}
}

func TestEnvelope_SuccessExitCode(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.MustRun("discover")
	r.AssertExitCode(0)
}
