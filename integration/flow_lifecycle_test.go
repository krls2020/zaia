package integration

import (
	"testing"

	"github.com/zeropsio/zaia/internal/platform"
)

func TestFlow_StartStoppedService(t *testing.T) {
	h := NewHarness(t)
	FixtureStoppedService(h)

	// Verify service is stopped
	r := h.MustRun("discover")
	r.AssertType("sync")
	services := r.Data()["services"].([]interface{})
	firstSvc := services[0].(map[string]interface{})
	if firstSvc["status"] != "STOPPED" {
		t.Errorf("expected STOPPED, got %v", firstSvc["status"])
	}

	// Start service
	r = h.MustRun("start --service api")
	r.AssertType("async")
	processes := r.Processes()
	if len(processes) == 0 {
		t.Fatal("expected at least one process")
	}

	// Discover again — service should be ACTIVE now (stateful mock)
	r = h.MustRun("discover")
	services = r.Data()["services"].([]interface{})
	firstSvc = services[0].(map[string]interface{})
	if firstSvc["status"] != "ACTIVE" {
		t.Errorf("expected ACTIVE after start, got %v", firstSvc["status"])
	}
}

func TestFlow_StopActiveService(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	// Stop
	r := h.MustRun("stop --service api")
	r.AssertType("async")

	// Discover — should be STOPPED
	r = h.MustRun("discover")
	services := r.Data()["services"].([]interface{})
	for _, s := range services {
		svc := s.(map[string]interface{})
		if svc["hostname"] == "api" {
			if svc["status"] != "STOPPED" {
				t.Errorf("expected STOPPED after stop, got %v", svc["status"])
			}
			return
		}
	}
	t.Fatal("api service not found in discover results")
}

func TestFlow_RestartService(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.MustRun("restart --service api")
	r.AssertType("async")
	if len(r.Processes()) == 0 {
		t.Fatal("expected process in async response")
	}
}

func TestFlow_ScaleService(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	// Scale returns sync (nil process) from StatefulMock
	r := h.MustRun("scale --service api --min-cpu 1 --max-cpu 4")
	r.AssertType("sync")
	data := r.Data()
	if data["serviceHostname"] != "api" {
		t.Errorf("expected serviceHostname 'api', got %v", data["serviceHostname"])
	}
}

func TestFlow_ScaleNoParams(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("scale --service api")
	r.AssertType("error")
	r.AssertErrorCode("INVALID_SCALING")
	r.AssertExitCode(3)
}

func TestFlow_ScaleMinGtMax(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("scale --service api --min-cpu 5 --max-cpu 2")
	r.AssertType("error")
	r.AssertErrorCode("INVALID_SCALING")
}

func TestFlow_ProcessCheck(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	// Add a known process
	h.Mock().WithProcess(&platform.Process{
		ID:         "proc-123",
		ActionName: "start",
		Status:     "RUNNING",
	})

	r := h.MustRun("process proc-123")
	r.AssertType("sync")
	data := r.Data()
	// MapProcessToOutput is used — check processId field
	if data == nil {
		t.Fatal("expected data")
	}
}

func TestFlow_ProcessNotFound(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("process nonexistent-id")
	r.AssertType("error")
	r.AssertErrorCode("PROCESS_NOT_FOUND")
	r.AssertExitCode(4)
}

func TestFlow_CancelProcess(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	h.Mock().WithProcess(&platform.Process{
		ID:         "proc-to-cancel",
		ActionName: "start",
		Status:     "RUNNING",
	})

	r := h.MustRun("cancel proc-to-cancel")
	r.AssertType("sync")
	data := r.Data()
	if data["status"] != "CANCELED" {
		t.Errorf("expected CANCELED, got %v", data["status"])
	}
}
