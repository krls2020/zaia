package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewProcess creates the process command for checking async process status.
func NewProcess(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process <id>",
		Short: "Check process status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			processID := args[0]
			ctx := cmd.Context()

			process, err := client.GetProcess(ctx, processID)
			if err != nil {
				return output.Err(platform.ErrProcessNotFound,
					fmt.Sprintf("Process '%s' not found", processID), "", nil)
			}

			out := output.MapProcessToOutput(process, "")
			return output.Sync(out)
		},
	}
	return cmd
}
