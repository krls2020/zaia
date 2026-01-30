package integration

import (
	"testing"

	"github.com/zeropsio/zaia/internal/platform"
)

func TestFlow_EnvSetThenGet(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	// Set env vars
	r := h.MustRun("env set --service api DB_HOST=localhost DB_PORT=5432")
	r.AssertType("async")

	// Get env vars — should see the ones we just set
	r = h.MustRun("env get --service api")
	r.AssertType("sync")
	data := r.Data()
	vars := data["vars"].([]interface{})
	if len(vars) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(vars))
	}

	// Verify values
	found := make(map[string]string)
	for _, v := range vars {
		vm := v.(map[string]interface{})
		found[vm["key"].(string)] = vm["value"].(string)
	}
	if found["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %s", found["DB_HOST"])
	}
	if found["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT=5432, got %s", found["DB_PORT"])
	}
}

func TestFlow_EnvSetThenDeleteThenGet(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	// Set 2 vars
	h.MustRun("env set --service api KEY1=val1 KEY2=val2")

	// Delete one
	r := h.MustRun("env delete --service api KEY1")
	r.AssertType("async")

	// Get — should only have KEY2
	r = h.MustRun("env get --service api")
	data := r.Data()
	vars := data["vars"].([]interface{})
	if len(vars) != 1 {
		t.Fatalf("expected 1 env var after delete, got %d", len(vars))
	}
	vm := vars[0].(map[string]interface{})
	if vm["key"] != "KEY2" {
		t.Errorf("expected KEY2, got %v", vm["key"])
	}
}

func TestFlow_EnvGetProject(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)
	h.Mock().WithProjectEnv([]platform.EnvVar{
		{ID: "pe-1", Key: "APP_ENV", Content: "production"},
	})

	r := h.MustRun("env get --project")
	r.AssertType("sync")
	data := r.Data()
	if data["scope"] != "project" {
		t.Errorf("expected scope=project, got %v", data["scope"])
	}
	vars := data["vars"].([]interface{})
	if len(vars) != 1 {
		t.Fatalf("expected 1 project env var, got %d", len(vars))
	}
}

func TestFlow_EnvSetInvalidFormat(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("env set --service api NOEQUALS")
	r.AssertType("error")
	r.AssertErrorCode("INVALID_ENV_FORMAT")
	r.AssertExitCode(3)
}

func TestFlow_EnvSetEmptyValue(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	// Empty value is valid
	r := h.MustRun("env set --service api KEY=")
	r.AssertType("async")

	r = h.MustRun("env get --service api")
	data := r.Data()
	vars := data["vars"].([]interface{})
	if len(vars) != 1 {
		t.Fatalf("expected 1 env var, got %d", len(vars))
	}
	vm := vars[0].(map[string]interface{})
	if vm["value"] != "" {
		t.Errorf("expected empty value, got %q", vm["value"])
	}
}
