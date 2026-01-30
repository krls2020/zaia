package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadClaudeConfig_NonExistent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".claude.json")
	cfg, err := LoadClaudeConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("config should not be nil")
	}
	if len(cfg) != 0 {
		t.Errorf("config should be empty, got %v", cfg)
	}
}

func TestLoadClaudeConfig_Existing(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".claude.json")
	existing := `{"someKey": "someValue", "mcpServers": {"other": {"type": "stdio"}}}`
	_ = os.WriteFile(path, []byte(existing), 0644)

	cfg, err := LoadClaudeConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg["someKey"] != "someValue" {
		t.Errorf("someKey = %v, want someValue", cfg["someKey"])
	}
}

func TestLoadClaudeConfig_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".claude.json")
	_ = os.WriteFile(path, []byte("{invalid json"), 0644)

	_, err := LoadClaudeConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAddMCPServer(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]interface{}
		binPath  string
		wantType string
		wantCmd  string
	}{
		{
			name:     "empty_config",
			initial:  map[string]interface{}{},
			binPath:  "/home/user/.local/bin/zaia-mcp",
			wantType: "stdio",
			wantCmd:  "/home/user/.local/bin/zaia-mcp",
		},
		{
			name: "existing_mcpServers",
			initial: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"other-server": map[string]interface{}{"type": "stdio"},
				},
			},
			binPath:  "/usr/local/bin/zaia-mcp",
			wantType: "stdio",
			wantCmd:  "/usr/local/bin/zaia-mcp",
		},
		{
			name: "overwrite_existing_zaia_mcp",
			initial: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"zaia-mcp": map[string]interface{}{
						"type":    "stdio",
						"command": "/old/path/zaia-mcp",
					},
				},
			},
			binPath:  "/new/path/zaia-mcp",
			wantType: "stdio",
			wantCmd:  "/new/path/zaia-mcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.initial
			AddMCPServer(cfg, tt.binPath)

			servers, ok := cfg["mcpServers"].(map[string]interface{})
			if !ok {
				t.Fatal("mcpServers should be a map")
			}
			entry, ok := servers["zaia-mcp"].(map[string]interface{})
			if !ok {
				t.Fatal("zaia-mcp entry should be a map")
			}
			if entry["type"] != tt.wantType {
				t.Errorf("type = %v, want %v", entry["type"], tt.wantType)
			}
			if entry["command"] != tt.wantCmd {
				t.Errorf("command = %v, want %v", entry["command"], tt.wantCmd)
			}
		})
	}
}

func TestAddMCPServer_PreservesOtherKeys(t *testing.T) {
	cfg := map[string]interface{}{
		"theme":    "dark",
		"language": "en",
		"mcpServers": map[string]interface{}{
			"other-server": map[string]interface{}{"type": "stdio", "command": "/bin/other"},
		},
	}
	AddMCPServer(cfg, "/bin/zaia-mcp")

	if cfg["theme"] != "dark" {
		t.Error("theme key should be preserved")
	}
	if cfg["language"] != "en" {
		t.Error("language key should be preserved")
	}
	servers := cfg["mcpServers"].(map[string]interface{})
	if _, ok := servers["other-server"]; !ok {
		t.Error("other-server should be preserved")
	}
}

func TestSaveClaudeConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")

	cfg := map[string]interface{}{
		"theme": "dark",
		"mcpServers": map[string]interface{}{
			"zaia-mcp": map[string]interface{}{
				"type":    "stdio",
				"command": "/bin/zaia-mcp",
				"args":    []interface{}{},
				"env":     map[string]interface{}{},
			},
		},
	}

	err := SaveClaudeConfig(path, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	var loaded map[string]interface{}
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("parsing JSON: %v", err)
	}
	if loaded["theme"] != "dark" {
		t.Error("theme should be dark")
	}
}

func TestSaveClaudeConfig_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", ".claude.json")

	err := SaveClaudeConfig(path, map[string]interface{}{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	p := DefaultConfigPath()
	if !filepath.IsAbs(p) {
		t.Errorf("path should be absolute: %v", p)
	}
	if filepath.Base(p) != ".claude.json" {
		t.Errorf("filename = %v, want .claude.json", filepath.Base(p))
	}
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")

	// Write initial config with other keys
	initial := `{"globalSettings": {"theme": "dark"}, "mcpServers": {"existing": {"type": "stdio"}}}`
	_ = os.WriteFile(path, []byte(initial), 0644)

	// Load, modify, save
	cfg, err := LoadClaudeConfig(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	AddMCPServer(cfg, "/bin/zaia-mcp")
	err = SaveClaudeConfig(path, cfg)
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	// Load again and verify
	cfg2, err := LoadClaudeConfig(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	// Existing keys preserved
	if cfg2["globalSettings"] == nil {
		t.Error("globalSettings should be preserved")
	}
	// Existing MCP server preserved
	servers := cfg2["mcpServers"].(map[string]interface{})
	if servers["existing"] == nil {
		t.Error("existing server should be preserved")
	}
	// New zaia-mcp server added
	entry := servers["zaia-mcp"].(map[string]interface{})
	if entry["command"] != "/bin/zaia-mcp" {
		t.Errorf("command = %v, want /bin/zaia-mcp", entry["command"])
	}
}
