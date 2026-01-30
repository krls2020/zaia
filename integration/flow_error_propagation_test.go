package integration

import (
	"testing"
)

// TestFlow_UnauthenticatedCommands verifies all commands requiring auth return AUTH_REQUIRED.
func TestFlow_UnauthenticatedCommands(t *testing.T) {
	commands := []struct {
		name    string
		cmdLine string
	}{
		{"discover", "discover"},
		{"logs", "logs --service api"},
		{"process", "process some-id"},
		{"cancel", "cancel some-id"},
		{"start", "start --service api"},
		{"stop", "stop --service api"},
		{"restart", "restart --service api"},
		{"scale", "scale --service api --min-cpu 1"},
		{"env get", "env get --service api"},
		{"env set", "env set --service api KEY=val"},
		{"env delete", "env delete --service api KEY"},
		{"import", "import --content 'services:\n  - hostname: x\n    type: nodejs@22'"},
		{"delete", "delete --service api --confirm"},
		{"subdomain", "subdomain enable --service api"},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHarness(t)
			FixtureUnauthenticated(h) // no stored credentials

			r := h.Run(tc.cmdLine)
			r.AssertType("error")
			r.AssertErrorCode("AUTH_REQUIRED")
			r.AssertExitCode(2)
		})
	}
}

// TestFlow_ServiceNotFound verifies commands requiring --service return SERVICE_NOT_FOUND.
func TestFlow_ServiceNotFound(t *testing.T) {
	commands := []struct {
		name    string
		cmdLine string
	}{
		{"start", "start --service nonexistent"},
		{"stop", "stop --service nonexistent"},
		{"restart", "restart --service nonexistent"},
		{"scale", "scale --service nonexistent --min-cpu 1"},
		{"env get", "env get --service nonexistent"},
		{"env set", "env set --service nonexistent KEY=val"},
		{"env delete", "env delete --service nonexistent KEY"},
		{"delete", "delete --service nonexistent --confirm"},
		{"subdomain", "subdomain enable --service nonexistent"},
		{"discover --service", "discover --service nonexistent"},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHarness(t)
			FixtureFullProject(h) // authenticated with 3 services: api, db, cache

			r := h.Run(tc.cmdLine)
			r.AssertType("error")
			r.AssertErrorCode("SERVICE_NOT_FOUND")
			r.AssertExitCode(4)
		})
	}
}
