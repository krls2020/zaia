package integration

import (
	"testing"
)

func TestFlow_ImportServices(t *testing.T) {
	h := NewHarness(t)
	FixtureEmptyProject(h)

	yaml := "services:\n  - hostname: worker\n    type: nodejs@22"
	r := h.MustRun("import --content '" + yaml + "'")
	r.AssertType("async")
	if len(r.Processes()) == 0 {
		t.Fatal("expected processes in import response")
	}
}

func TestFlow_ImportWithProjectSection(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	yaml := "project:\n  name: bad\nservices:\n  - hostname: api\n    type: nodejs@22"
	r := h.Run("import --content '" + yaml + "'")
	r.AssertType("error")
	r.AssertErrorCode("IMPORT_HAS_PROJECT")
	r.AssertExitCode(3)
}

func TestFlow_ImportDryRun(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	yaml := "services:\n  - hostname: worker\n    type: nodejs@22"
	r := h.MustRun("import --content '" + yaml + "' --dry-run")
	r.AssertType("sync")
	data := r.Data()
	if data["dryRun"] != true {
		t.Errorf("expected dryRun=true, got %v", data["dryRun"])
	}
	if data["valid"] != true {
		t.Errorf("expected valid=true, got %v", data["valid"])
	}
}

func TestFlow_DeleteServiceWithConfirm(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	// 3 services initially
	r := h.MustRun("discover")
	services := r.Data()["services"].([]interface{})
	initialCount := len(services)

	// Delete api
	r = h.MustRun("delete --service api --confirm")
	r.AssertType("async")

	// Discover — should have one fewer service
	r = h.MustRun("discover")
	services = r.Data()["services"].([]interface{})
	if len(services) != initialCount-1 {
		t.Errorf("expected %d services after delete, got %d", initialCount-1, len(services))
	}

	// Verify "api" is gone
	for _, s := range services {
		svc := s.(map[string]interface{})
		if svc["hostname"] == "api" {
			t.Error("api service should have been deleted")
		}
	}
}

func TestFlow_DeleteWithoutConfirm(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("delete --service api")
	r.AssertType("error")
	r.AssertErrorCode("CONFIRM_REQUIRED")
	r.AssertExitCode(3)
}

func TestFlow_DeleteNonexistentService(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("delete --service nonexistent --confirm")
	r.AssertType("error")
	r.AssertErrorCode("SERVICE_NOT_FOUND")
	r.AssertExitCode(4)
}

func TestFlow_ImportNoInput(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("import")
	r.AssertType("error")
	r.AssertErrorCode("INVALID_PARAMETER")
}
