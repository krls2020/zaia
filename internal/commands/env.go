package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewEnv creates the env command with get/set/delete subcommands.
func NewEnv(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environment variables",
		RunE: func(cmd *cobra.Command, args []string) error {
			return output.Err(platform.ErrInvalidUsage,
				"No subcommand specified for 'env'",
				"Run: zaia env <get|set|delete>",
				map[string]interface{}{"availableSubcommands": []string{"get", "set", "delete"}})
		},
	}

	cmd.AddCommand(newEnvGet(storagePath, client))
	cmd.AddCommand(newEnvSet(storagePath, client))
	cmd.AddCommand(newEnvDelete(storagePath, client))

	return cmd
}

func newEnvGet(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get environment variables",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			isProject, _ := cmd.Flags().GetBool("project")
			hostname, _ := cmd.Flags().GetString("service")

			if isProject {
				envs, err := client.GetProjectEnv(ctx, creds.ProjectID)
				if err != nil {
					return output.Err(platform.ErrAPIError, err.Error(), "", nil)
				}
				vars := make([]map[string]interface{}, len(envs))
				for i, e := range envs {
					vars[i] = map[string]interface{}{"key": e.Key, "value": e.Content}
				}
				return output.Sync(map[string]interface{}{
					"scope": "project",
					"vars":  vars,
				})
			}

			if hostname == "" {
				return output.Err(platform.ErrServiceRequired,
					"--service or --project flag is required",
					"Run: zaia env get --service <hostname> or zaia env get --project", nil)
			}

			services, err := client.ListServices(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}
			svc, err := resolveServiceID(client, creds.ProjectID, hostname, services)
			if err != nil {
				return err
			}

			envs, err := client.GetServiceEnv(ctx, svc.ID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}
			vars := make([]map[string]interface{}, len(envs))
			for i, e := range envs {
				vars[i] = map[string]interface{}{"key": e.Key, "value": e.Content}
			}
			return output.Sync(map[string]interface{}{
				"scope":           "service",
				"serviceHostname": hostname,
				"vars":            vars,
			})
		},
	}

	cmd.Flags().String("service", "", "Service hostname")
	cmd.Flags().Bool("project", false, "Get project-level env vars")
	return cmd
}

func newEnvSet(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [KEY=val ...]",
		Short: "Set environment variables",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			// Parse KEY=value pairs
			pairs, err := parseEnvPairs(args)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			isProject, _ := cmd.Flags().GetBool("project")
			hostname, _ := cmd.Flags().GetString("service")

			if isProject {
				// For project env, create each var individually
				var lastProcess *platform.Process
				for _, p := range pairs {
					proc, err := client.CreateProjectEnv(ctx, creds.ProjectID, p.Key, p.Value, false)
					if err != nil {
						return output.Err(platform.ErrAPIError, err.Error(), "", nil)
					}
					lastProcess = proc
				}
				if lastProcess != nil {
					return output.Async([]output.ProcessOutput{
						output.MapProcessToOutput(lastProcess, ""),
					})
				}
				return output.Sync(map[string]interface{}{"message": "No variables to set"})
			}

			if hostname == "" {
				return output.Err(platform.ErrServiceRequired,
					"--service or --project flag is required",
					"Run: zaia env set --service <hostname> KEY=val", nil)
			}

			services, err := client.ListServices(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}
			svc, err := resolveServiceID(client, creds.ProjectID, hostname, services)
			if err != nil {
				return err
			}

			// Build .env format content
			var content strings.Builder
			for _, p := range pairs {
				fmt.Fprintf(&content, "%s=%s\n", p.Key, p.Value)
			}

			process, err := client.SetServiceEnvFile(ctx, svc.ID, content.String())
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			return output.Async([]output.ProcessOutput{
				output.MapProcessToOutput(process, hostname),
			})
		},
	}

	cmd.Flags().String("service", "", "Service hostname")
	cmd.Flags().Bool("project", false, "Set project-level env vars")
	return cmd
}

func newEnvDelete(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [KEY ...]",
		Short: "Delete environment variables",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			isProject, _ := cmd.Flags().GetBool("project")
			hostname, _ := cmd.Flags().GetString("service")

			if isProject {
				// Get project envs to find IDs by key
				envs, err := client.GetProjectEnv(ctx, creds.ProjectID)
				if err != nil {
					return output.Err(platform.ErrAPIError, err.Error(), "", nil)
				}

				var lastProcess *platform.Process
				for _, key := range args {
					envID := findEnvIDByKey(envs, key)
					if envID == "" {
						return output.Err(platform.ErrAPIError,
							fmt.Sprintf("Project env var '%s' not found", key), "", nil)
					}
					proc, err := client.DeleteProjectEnv(ctx, envID)
					if err != nil {
						return output.Err(platform.ErrAPIError, err.Error(), "", nil)
					}
					lastProcess = proc
				}
				if lastProcess != nil {
					return output.Async([]output.ProcessOutput{
						output.MapProcessToOutput(lastProcess, ""),
					})
				}
				return output.Sync(nil)
			}

			if hostname == "" {
				return output.Err(platform.ErrServiceRequired,
					"--service or --project flag is required",
					"Run: zaia env delete --service <hostname> KEY", nil)
			}

			services, err := client.ListServices(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}
			svc, err := resolveServiceID(client, creds.ProjectID, hostname, services)
			if err != nil {
				return err
			}

			// Get service envs to find userDataId by key
			envs, err := client.GetServiceEnv(ctx, svc.ID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			var lastProcess *platform.Process
			for _, key := range args {
				envID := findEnvIDByKey(envs, key)
				if envID == "" {
					return output.Err(platform.ErrAPIError,
						fmt.Sprintf("Service env var '%s' not found", key),
						fmt.Sprintf("Available vars: %s", listEnvKeys(envs)), nil)
				}
				proc, err := client.DeleteUserData(ctx, envID)
				if err != nil {
					return output.Err(platform.ErrAPIError, err.Error(), "", nil)
				}
				lastProcess = proc
			}
			if lastProcess != nil {
				return output.Async([]output.ProcessOutput{
					output.MapProcessToOutput(lastProcess, hostname),
				})
			}
			return output.Sync(nil)
		},
	}

	cmd.Flags().String("service", "", "Service hostname")
	cmd.Flags().Bool("project", false, "Delete project-level env vars")
	return cmd
}

type envPair struct {
	Key   string
	Value string
}

func parseEnvPairs(args []string) ([]envPair, error) {
	pairs := make([]envPair, 0, len(args))
	for _, arg := range args {
		idx := strings.IndexByte(arg, '=')
		if idx < 0 {
			return nil, output.Err(platform.ErrInvalidEnvFormat,
				fmt.Sprintf("Invalid format '%s', expected KEY=value", arg),
				"Format: KEY=value (split on first '=')", nil)
		}
		key := arg[:idx]
		value := arg[idx+1:]
		if key == "" {
			return nil, output.Err(platform.ErrInvalidEnvFormat,
				"Empty key in env var",
				"Format: KEY=value", nil)
		}
		pairs = append(pairs, envPair{Key: key, Value: value})
	}
	return pairs, nil
}

func findEnvIDByKey(envs []platform.EnvVar, key string) string {
	for _, e := range envs {
		if e.Key == key {
			return e.ID
		}
	}
	return ""
}

func listEnvKeys(envs []platform.EnvVar) string {
	if len(envs) == 0 {
		return "(none)"
	}
	keys := make([]string, len(envs))
	for i, e := range envs {
		keys[i] = e.Key
	}
	return strings.Join(keys, ", ")
}
