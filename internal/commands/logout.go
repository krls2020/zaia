package commands

import (
	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/auth"
	"github.com/zeropsio/zaia/internal/output"
)

// NewLogout creates the logout command.
func NewLogout(storageDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			storage := auth.NewStorage(storageDir)
			mgr := auth.NewManager(storage, nil)

			if err := mgr.Logout(); err != nil {
				return output.Err("API_ERROR", err.Error(), "", nil)
			}

			return output.Sync(map[string]interface{}{
				"message": "Logged out successfully",
			})
		},
	}
}
