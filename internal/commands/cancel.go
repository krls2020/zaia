package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewCancel creates the cancel command for canceling async processes.
func NewCancel(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <process-id>",
		Short: "Cancel an async process",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			processID := args[0]
			ctx := cmd.Context()

			// First check current status
			process, err := client.GetProcess(ctx, processID)
			if err != nil {
				return output.Err(platform.ErrProcessNotFound,
					fmt.Sprintf("Process '%s' not found", processID), "", nil)
			}

			// Check if already in terminal state
			status := output.MapProcessToOutput(process, "").Status
			if status == "FINISHED" || status == "FAILED" || status == "CANCELED" {
				return output.Err(platform.ErrProcessAlreadyTerminal,
					fmt.Sprintf("Process is already in terminal state: %s", status),
					fmt.Sprintf("Process %s has already completed", processID), nil)
			}

			// Cancel
			_, err = client.CancelProcess(ctx, processID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			return output.Sync(map[string]interface{}{
				"processId": processID,
				"status":    "CANCELED",
				"message":   "Process canceled successfully",
			})
		},
	}
	return cmd
}
