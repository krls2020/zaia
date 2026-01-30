package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// newLifecycleCmd creates a lifecycle command (start/stop/restart).
func newLifecycleCmd(storagePath string, client platform.Client, action string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   action,
		Short: fmt.Sprintf("%s a service", action),
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
					fmt.Sprintf("Run: zaia %s --service <hostname>", action), nil)
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
			switch action {
			case "start":
				process, err = client.StartService(ctx, svc.ID)
			case "stop":
				process, err = client.StopService(ctx, svc.ID)
			case "restart":
				process, err = client.RestartService(ctx, svc.ID)
			}
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			return output.Async([]output.ProcessOutput{
				output.MapProcessToOutput(process, hostname),
			})
		},
	}

	cmd.Flags().String("service", "", "Service hostname (required)")
	return cmd
}

// NewStart creates the start command.
func NewStart(storagePath string, client platform.Client) *cobra.Command {
	return newLifecycleCmd(storagePath, client, "start")
}

// NewStop creates the stop command.
func NewStop(storagePath string, client platform.Client) *cobra.Command {
	return newLifecycleCmd(storagePath, client, "stop")
}

// NewRestart creates the restart command.
func NewRestart(storagePath string, client platform.Client) *cobra.Command {
	return newLifecycleCmd(storagePath, client, "restart")
}

// NewScale creates the scale command.
func NewScale(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale",
		Short: "Scale a service",
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
					"Run: zaia scale --service <hostname> [scaling flags]", nil)
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

			params, err := parseScalingParams(cmd)
			if err != nil {
				return err
			}

			process, err := client.SetAutoscaling(ctx, svc.ID, params)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			// API may return nil process (sync — applied immediately)
			if process == nil {
				return output.Sync(map[string]interface{}{
					"message":         "Scaling parameters updated",
					"serviceHostname": hostname,
					"serviceId":       svc.ID,
				})
			}

			return output.Async([]output.ProcessOutput{
				output.MapProcessToOutput(process, hostname),
			})
		},
	}

	cmd.Flags().String("service", "", "Service hostname (required)")
	cmd.Flags().String("cpu-mode", "", "CPU mode: SHARED or DEDICATED")
	cmd.Flags().Int32("min-cpu", 0, "Min CPU cores (1-8)")
	cmd.Flags().Int32("max-cpu", 0, "Max CPU cores (1-8)")
	cmd.Flags().Float64("min-ram", 0, "Min RAM in GB (0.125-48)")
	cmd.Flags().Float64("max-ram", 0, "Max RAM in GB (0.125-48)")
	cmd.Flags().Float64("min-disk", 0, "Min disk in GB (0.5-250)")
	cmd.Flags().Float64("max-disk", 0, "Max disk in GB (0.5-250)")
	cmd.Flags().Int32("min-replicas", 0, "Min containers (1-10)")
	cmd.Flags().Int32("max-replicas", 0, "Max containers (1-10)")

	return cmd
}

func parseScalingParams(cmd *cobra.Command) (platform.AutoscalingParams, error) {
	var params platform.AutoscalingParams
	hasAny := false

	if cmd.Flags().Changed("cpu-mode") {
		v, _ := cmd.Flags().GetString("cpu-mode")
		if v != "SHARED" && v != "DEDICATED" {
			return params, output.Err(platform.ErrInvalidScaling,
				"Invalid --cpu-mode: must be SHARED or DEDICATED",
				"Use: --cpu-mode SHARED or --cpu-mode DEDICATED", nil)
		}
		params.VerticalCpuMode = &v
		hasAny = true
	}

	if cmd.Flags().Changed("min-cpu") {
		v, _ := cmd.Flags().GetInt32("min-cpu")
		params.VerticalMinCpu = &v
		hasAny = true
	}
	if cmd.Flags().Changed("max-cpu") {
		v, _ := cmd.Flags().GetInt32("max-cpu")
		params.VerticalMaxCpu = &v
		hasAny = true
	}
	if cmd.Flags().Changed("min-ram") {
		v, _ := cmd.Flags().GetFloat64("min-ram")
		params.VerticalMinRam = &v
		hasAny = true
	}
	if cmd.Flags().Changed("max-ram") {
		v, _ := cmd.Flags().GetFloat64("max-ram")
		params.VerticalMaxRam = &v
		hasAny = true
	}
	if cmd.Flags().Changed("min-disk") {
		v, _ := cmd.Flags().GetFloat64("min-disk")
		params.VerticalMinDisk = &v
		hasAny = true
	}
	if cmd.Flags().Changed("max-disk") {
		v, _ := cmd.Flags().GetFloat64("max-disk")
		params.VerticalMaxDisk = &v
		hasAny = true
	}
	if cmd.Flags().Changed("min-replicas") {
		v, _ := cmd.Flags().GetInt32("min-replicas")
		params.HorizontalMinCount = &v
		hasAny = true
	}
	if cmd.Flags().Changed("max-replicas") {
		v, _ := cmd.Flags().GetInt32("max-replicas")
		params.HorizontalMaxCount = &v
		hasAny = true
	}

	if !hasAny {
		return params, output.Err(platform.ErrInvalidScaling,
			"At least one scaling parameter required",
			"Available: --cpu-mode, --min-cpu, --max-cpu, --min-ram, --max-ram, --min-disk, --max-disk, --min-replicas, --max-replicas", nil)
	}

	// Validate min <= max pairs
	if params.VerticalMinCpu != nil && params.VerticalMaxCpu != nil && *params.VerticalMinCpu > *params.VerticalMaxCpu {
		return params, output.Err(platform.ErrInvalidScaling,
			"Invalid scaling parameters: minCpu must be <= maxCpu",
			"Set --min-cpu <= --max-cpu", nil)
	}
	if params.VerticalMinRam != nil && params.VerticalMaxRam != nil && *params.VerticalMinRam > *params.VerticalMaxRam {
		return params, output.Err(platform.ErrInvalidScaling,
			"Invalid scaling parameters: minRam must be <= maxRam",
			"Set --min-ram <= --max-ram", nil)
	}
	if params.VerticalMinDisk != nil && params.VerticalMaxDisk != nil && *params.VerticalMinDisk > *params.VerticalMaxDisk {
		return params, output.Err(platform.ErrInvalidScaling,
			"Invalid scaling parameters: minDisk must be <= maxDisk",
			"Set --min-disk <= --max-disk", nil)
	}
	if params.HorizontalMinCount != nil && params.HorizontalMaxCount != nil && *params.HorizontalMinCount > *params.HorizontalMaxCount {
		return params, output.Err(platform.ErrInvalidScaling,
			"Invalid scaling parameters: minReplicas must be <= maxReplicas",
			"Set --min-replicas <= --max-replicas", nil)
	}

	return params, nil
}
