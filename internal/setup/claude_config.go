package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultConfigPath returns ~/.claude.json.
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude.json")
}

// LoadClaudeConfig reads and parses ~/.claude.json.
// Returns an empty map if the file doesn't exist.
func LoadClaudeConfig(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// AddMCPServer adds or overwrites the zaia-mcp entry in mcpServers.
// It sets PATH in env to include the binary's directory so the MCP server
// can find zaia and zcli at runtime (IDEs don't inherit /etc/paths).
func AddMCPServer(cfg map[string]interface{}, binaryPath string) {
	servers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
	}

	binDir := filepath.Dir(binaryPath)
	path := binDir + ":/usr/local/bin:/usr/bin:/bin"

	servers["zaia-mcp"] = map[string]interface{}{
		"type":    "stdio",
		"command": binaryPath,
		"args":    []interface{}{},
		"env": map[string]interface{}{
			"PATH": path,
		},
	}

	cfg["mcpServers"] = servers
}

// SaveClaudeConfig writes the config to path as indented JSON.
// Creates parent directories if needed.
func SaveClaudeConfig(path string, cfg map[string]interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Append newline for POSIX compliance
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
