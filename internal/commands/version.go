package commands

import (
	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
)

// NewVersion creates the version command.
func NewVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show CLI version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output.Sync(map[string]interface{}{
				"version": version,
				"commit":  commit,
				"built":   built,
			})
		},
	}
}
