package integration

import (
	"fmt"
	"testing"
)

func TestFlow_SubdomainEnable(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.MustRun("subdomain --service api --action enable")
	r.AssertType("async")
	if len(r.Processes()) == 0 {
		t.Fatal("expected processes")
	}
}

func TestFlow_SubdomainDisable(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.MustRun("subdomain --service api --action disable")
	r.AssertType("async")
}

func TestFlow_SubdomainIdempotentEnable(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	// Set error to simulate "already enabled"
	h.Mock().WithError("EnableSubdomainAccess", fmt.Errorf("AlreadyEnabled"))

	r := h.Run("subdomain --service api --action enable")
	r.AssertType("sync")
	data := r.Data()
	if data["status"] != "already_enabled" {
		t.Errorf("expected already_enabled, got %v", data["status"])
	}
}

func TestFlow_SubdomainIdempotentDisable(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	h.Mock().WithError("DisableSubdomainAccess", fmt.Errorf("AlreadyDisabled"))

	r := h.Run("subdomain --service api --action disable")
	r.AssertType("sync")
	data := r.Data()
	if data["status"] != "already_disabled" {
		t.Errorf("expected already_disabled, got %v", data["status"])
	}
}

func TestFlow_SubdomainInvalidAction(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.Run("subdomain --service api --action toggle")
	r.AssertType("error")
	r.AssertErrorCode("INVALID_PARAMETER")
}
