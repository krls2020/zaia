package integration

import (
	"testing"

	"github.com/zeropsio/zaia/internal/platform"
)

func TestFlow_LoginThenDiscoverThenLogoutThenDiscoverFails(t *testing.T) {
	h := NewHarness(t)
	FixtureUnauthenticated(h)

	// 1. Login
	r := h.MustRun("login --token test-token")
	r.AssertType("sync")
	data := r.Data()
	if data == nil {
		t.Fatal("expected data in login response")
	}
	project, _ := data["project"].(map[string]interface{})
	if project["name"] != "test-project" {
		t.Errorf("expected project name 'test-project', got %v", project["name"])
	}

	// 2. Discover (should work now)
	r = h.MustRun("discover")
	r.AssertType("sync")
	discoverData := r.Data()
	if discoverData == nil {
		t.Fatal("expected data in discover response")
	}

	// 3. Logout
	r = h.MustRun("logout")
	r.AssertType("sync")

	// 4. Discover (should fail — no auth)
	r = h.Run("discover")
	r.AssertType("error")
	r.AssertErrorCode("AUTH_REQUIRED")
	r.AssertExitCode(2)
}

func TestFlow_LoginMultiProjectToken(t *testing.T) {
	h := NewHarness(t)
	FixtureUnauthenticated(h)
	// Override with 2 projects
	h.Mock().WithProjects([]platform.Project{
		{ID: "proj-1", Name: "project-one", Status: "ACTIVE"},
		{ID: "proj-2", Name: "project-two", Status: "ACTIVE"},
	})

	r := h.Run("login --token test-token")
	r.AssertType("error")
	r.AssertErrorCode("TOKEN_MULTI_PROJECT")
	r.AssertExitCode(2)
}

func TestFlow_LoginNoProjectToken(t *testing.T) {
	h := NewHarness(t)
	FixtureUnauthenticated(h)
	h.Mock().WithProjects(nil) // 0 projects

	r := h.Run("login --token test-token")
	r.AssertType("error")
	r.AssertErrorCode("TOKEN_NO_PROJECT")
	r.AssertExitCode(2)
}

func TestFlow_StatusShowsProjectInfo(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	r := h.MustRun("status")
	r.AssertType("sync")
	data := r.Data()
	if data == nil {
		t.Fatal("expected data in status response")
	}
}

func TestFlow_StatusWhenUnauthenticated(t *testing.T) {
	h := NewHarness(t)
	FixtureUnauthenticated(h)

	r := h.Run("status")
	r.AssertType("error")
	r.AssertErrorCode("AUTH_REQUIRED")
}
