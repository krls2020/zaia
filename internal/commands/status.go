package commands

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/auth"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewStatus creates the status command.
func NewStatus(storageDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			storage := auth.NewStorage(storageDir)
			mgr := auth.NewManager(storage, nil)

			data, err := mgr.GetStatus()
			if err != nil {
				var authErr *auth.AuthError
				if errors.As(err, &authErr) {
					return output.Err(authErr.Code, authErr.Message, authErr.Suggestion, nil)
				}
				return output.Err(platform.ErrAuthRequired, err.Error(), "Run: zaia login <token>", nil)
			}

			return output.Sync(map[string]interface{}{
				"authenticated": true,
				"user": map[string]interface{}{
					"name":  data.User.Name,
					"email": data.User.Email,
				},
				"project": map[string]interface{}{
					"id":   data.Project.ID,
					"name": data.Project.Name,
				},
				"region":  data.RegionData.Name,
				"apiHost": data.APIHost,
			})
		},
	}
}
