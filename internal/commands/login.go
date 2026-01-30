package commands

import (
	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/auth"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewLogin creates the login command.
// clientFactory creates a platform.Client from (token, apiHost).
func NewLogin(storageDir string, clientFactory func(token, apiHost string) platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login <token>",
		Short: "Authenticate with Zerops",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token := args[0]
			url, _ := cmd.Flags().GetString("url")
			region, _ := cmd.Flags().GetString("region")
			regionURL, _ := cmd.Flags().GetString("region-url")

			// Resolve API host for client creation
			apiHost := url
			if apiHost == "" {
				apiHost = "api.app-prg1.zerops.io"
			}

			client := clientFactory(token, apiHost)
			storage := auth.NewStorage(storageDir)
			mgr := auth.NewManager(storage, client)

			result, err := mgr.Login(cmd.Context(), token, auth.LoginOptions{
				URL:       url,
				Region:    region,
				RegionURL: regionURL,
			})
			if err != nil {
				if authErr, ok := err.(*auth.AuthError); ok {
					return output.Err(authErr.Code, authErr.Message, authErr.Suggestion, nil)
				}
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			return output.Sync(map[string]interface{}{
				"user": map[string]interface{}{
					"name":  result.User.Name,
					"email": result.User.Email,
				},
				"project": map[string]interface{}{
					"id":   result.Project.ID,
					"name": result.Project.Name,
				},
				"region": result.Region,
			})
		},
	}

	cmd.Flags().String("url", "", "API server URL (for staging/dev, skips region discovery)")
	cmd.Flags().String("region", "", "Region name")
	cmd.Flags().String("region-url", "", "Custom region metadata URL")

	return cmd
}
