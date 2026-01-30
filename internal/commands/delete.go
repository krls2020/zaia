package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewDelete creates the delete command.
func NewDelete(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a service",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			hostname, _ := cmd.Flags().GetString("service")
			confirm, _ := cmd.Flags().GetBool("confirm")

			if hostname == "" {
				return output.Err(platform.ErrServiceRequired,
					"--service flag is required",
					"Run: zaia delete --service <hostname> --confirm", nil)
			}

			if !confirm {
				return output.Err(platform.ErrConfirmRequired,
					"Destructive operation requires --confirm flag",
					fmt.Sprintf("Run: zaia delete --service %s --confirm", hostname),
					map[string]interface{}{
						"wouldDelete": map[string]string{
							"type":     "service",
							"hostname": hostname,
						},
					})
			}

			ctx := cmd.Context()
			services, err := client.ListServices(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			svc, err := resolveServiceID(client, creds.ProjectID, hostname, services)
			if err != nil {
				return err
			}

			process, err := client.DeleteService(ctx, svc.ID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			return output.Async([]output.ProcessOutput{
				output.MapProcessToOutput(process, hostname),
			})
		},
	}

	cmd.Flags().String("service", "", "Service hostname (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")

	return cmd
}
