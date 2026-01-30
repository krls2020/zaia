package commands

import (
	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
	"github.com/zeropsio/zaia/internal/setup"
)

// NewSetup creates the setup command with default dependencies.
func NewSetup() *cobra.Command {
	return NewSetupWithDeps(
		&setup.HTTPDownloader{},
		setup.DefaultBinaryPath(),
		setup.DefaultConfigPath(),
	)
}

// NewSetupWithDeps creates the setup command with injected dependencies for testing.
func NewSetupWithDeps(dl setup.Downloader, binPath, configPath string) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Install zaia-mcp and configure Claude Code MCP server",
		Long:  "Downloads the zaia-mcp binary and registers it as an MCP server in Claude Code's ~/.claude.json.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(dl, binPath, configPath)
		},
	}
}

func runSetup(dl setup.Downloader, binPath, configPath string) error {
	// 1. Detect platform and build download URL
	p := setup.DetectPlatform()
	url, err := setup.DownloadURL(p)
	if err != nil {
		return output.Err(platform.ErrSetupUnsupportedOS, err.Error(),
			"Only darwin and linux are supported", nil)
	}

	// 2. Download and install binary
	if err := setup.InstallBinary(dl, url, binPath); err != nil {
		return output.Err(platform.ErrSetupDownloadFailed, err.Error(),
			"Check your internet connection and try again", nil)
	}

	// 3. Load Claude config
	cfg, err := setup.LoadClaudeConfig(configPath)
	if err != nil {
		return output.Err(platform.ErrSetupConfigFailed, err.Error(),
			"Check ~/.claude.json for syntax errors", nil)
	}

	// 4. Add MCP server entry
	setup.AddMCPServer(cfg, binPath)

	// 5. Save config
	if err := setup.SaveClaudeConfig(configPath, cfg); err != nil {
		return output.Err(platform.ErrSetupConfigFailed, err.Error(),
			"Check file permissions for ~/.claude.json", nil)
	}

	return output.Sync(map[string]interface{}{
		"binaryPath":    binPath,
		"configPath":    configPath,
		"configUpdated": true,
	})
}
