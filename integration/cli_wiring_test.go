package integration

import (
	"testing"

	"github.com/zeropsio/zaia/internal/commands"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestWiring_AllCommandsRegistered(t *testing.T) {
	root := commands.NewRootForTest(commands.RootDeps{
		StoragePath: t.TempDir(),
		Client:      platform.NewMock(),
		LogFetcher:  platform.NewMockLogFetcher(),
	})

	expected := []string{
		"login", "logout", "status", "version",
		"discover", "process", "cancel", "logs",
		"validate", "search",
		"start", "stop", "restart", "scale",
		"env", "import", "delete", "subdomain",
		"events", "setup",
	}

	cmds := root.Commands()
	registered := make(map[string]bool)
	for _, c := range cmds {
		registered[c.Name()] = true
	}

	for _, name := range expected {
		if !registered[name] {
			t.Errorf("expected command %q to be registered, but it was not", name)
		}
	}

	if len(cmds) != len(expected) {
		t.Errorf("expected %d commands, got %d", len(expected), len(cmds))
		for _, c := range cmds {
			if !contains(expected, c.Name()) {
				t.Logf("unexpected command: %s", c.Name())
			}
		}
	}
}

func TestWiring_HelpWorks(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	root := commands.NewRootForTest(commands.RootDeps{
		StoragePath: h.StoragePath(),
		Client:      h.Mock(),
		LogFetcher:  platform.NewMockLogFetcher(),
	})
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("help should not fail: %v", err)
	}
}

func TestWiring_UnknownCommand(t *testing.T) {
	h := NewHarness(t)
	FixtureFullProject(h)

	root := commands.NewRootForTest(commands.RootDeps{
		StoragePath: h.StoragePath(),
		Client:      h.Mock(),
		LogFetcher:  platform.NewMockLogFetcher(),
	})
	root.SetArgs([]string{"nonexistent"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
