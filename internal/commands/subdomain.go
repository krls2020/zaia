package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NewSubdomain creates the subdomain command with enable/disable subcommands.
func NewSubdomain(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subdomain",
		Short: "Enable or disable Zerops subdomain access",
		RunE: func(cmd *cobra.Command, args []string) error {
			return output.Err(platform.ErrInvalidUsage,
				"No subcommand specified for 'subdomain'",
				"Run: zaia subdomain <enable|disable>",
				map[string]interface{}{"availableSubcommands": []string{"enable", "disable"}})
		},
	}

	cmd.AddCommand(newSubdomainAction(storagePath, client, "enable"))
	cmd.AddCommand(newSubdomainAction(storagePath, client, "disable"))

	return cmd
}

func newSubdomainAction(storagePath string, client platform.Client, action string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   action,
		Short: fmt.Sprintf("%s Zerops subdomain access", cases.Title(language.English).String(action)),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			hostname, _ := cmd.Flags().GetString("service")

			if hostname == "" {
				return output.Err(platform.ErrServiceRequired,
					"--service flag is required",
					fmt.Sprintf("Run: zaia subdomain %s --service <hostname>", action), nil)
			}

			ctx := cmd.Context()
			services, err := client.ListServices(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			svc, err := resolveServiceID(creds.ProjectID, hostname, services)
			if err != nil {
				return err
			}

			var process *platform.Process
			if action == "enable" {
				process, err = client.EnableSubdomainAccess(ctx, svc.ID)
			} else {
				process, err = client.DisableSubdomainAccess(ctx, svc.ID)
			}

			if err != nil {
				// Check for idempotent cases
				errMsg := err.Error()
				if strings.Contains(errMsg, "AlreadyEnabled") || strings.Contains(errMsg, "already enabled") {
					return output.Sync(map[string]interface{}{
						"serviceHostname": hostname,
						"serviceId":       svc.ID,
						"action":          action,
						"status":          "already_enabled",
					})
				}
				if strings.Contains(errMsg, "AlreadyDisabled") || strings.Contains(errMsg, "already disabled") {
					return output.Sync(map[string]interface{}{
						"serviceHostname": hostname,
						"serviceId":       svc.ID,
						"action":          action,
						"status":          "already_disabled",
					})
				}
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			actionName := fmt.Sprintf("%sSubdomain", action)
			proc := output.MapProcessToOutput(process, hostname)
			proc.ActionName = actionName

			return output.Async([]output.ProcessOutput{proc})
		},
	}

	cmd.Flags().String("service", "", "Service hostname (required)")

	return cmd
}
