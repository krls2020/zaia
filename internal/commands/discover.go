package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewDiscover creates the discover command.
func NewDiscover(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover project and services",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			project, err := client.GetProject(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(),
					"Project may have been deleted. Run: zaia login <token>", nil)
			}

			services, err := client.ListServices(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			serviceFlag, _ := cmd.Flags().GetString("service")
			includeEnvs, _ := cmd.Flags().GetBool("include-envs")

			projectData := map[string]interface{}{
				"id":     project.ID,
				"name":   project.Name,
				"status": project.Status,
			}

			if serviceFlag != "" {
				svc := findServiceByHostname(services, serviceFlag)
				if svc == nil {
					return output.Err(platform.ErrServiceNotFound,
						fmt.Sprintf("Service '%s' not found", serviceFlag),
						"Available services: "+listHostnames(services),
						map[string]interface{}{"projectId": creds.ProjectID})
				}

				detail := buildServiceDetail(svc)
				if includeEnvs {
					envs, err := client.GetServiceEnv(ctx, svc.ID)
					if err == nil {
						envData := make([]map[string]interface{}, len(envs))
						for i, e := range envs {
							envData[i] = map[string]interface{}{
								"key":   e.Key,
								"value": e.Content,
							}
						}
						detail["envs"] = envData
					}
				}

				return output.Sync(map[string]interface{}{
					"project":  projectData,
					"services": []interface{}{detail},
				})
			}

			// List all services
			svcList := make([]interface{}, len(services))
			for i, s := range services {
				svcData := map[string]interface{}{
					"hostname":  s.Name,
					"serviceId": s.ID,
					"type":      s.ServiceStackTypeInfo.ServiceStackTypeVersionName,
					"status":    s.Status,
				}
				if includeEnvs {
					envs, err := client.GetServiceEnv(ctx, s.ID)
					if err == nil {
						envData := make([]map[string]interface{}, len(envs))
						for j, e := range envs {
							envData[j] = map[string]interface{}{
								"key":   e.Key,
								"value": e.Content,
							}
						}
						svcData["envs"] = envData
					}
				}
				svcList[i] = svcData
			}

			return output.Sync(map[string]interface{}{
				"project":  projectData,
				"services": svcList,
			})
		},
	}

	cmd.Flags().String("service", "", "Get detailed info for specific service (hostname)")
	cmd.Flags().Bool("include-envs", false, "Include environment variables")

	return cmd
}

func buildServiceDetail(svc *platform.ServiceStack) map[string]interface{} {
	detail := map[string]interface{}{
		"hostname":  svc.Name,
		"serviceId": svc.ID,
		"type":      svc.ServiceStackTypeInfo.ServiceStackTypeVersionName,
		"status":    svc.Status,
		"created":   svc.Created,
	}

	if svc.CustomAutoscaling != nil {
		detail["containers"] = map[string]interface{}{
			"min": svc.CustomAutoscaling.HorizontalMinCount,
			"max": svc.CustomAutoscaling.HorizontalMaxCount,
		}
		detail["resources"] = map[string]interface{}{
			"cpuMode": svc.CustomAutoscaling.CpuMode,
			"cpu":     map[string]interface{}{"min": svc.CustomAutoscaling.MinCpu, "max": svc.CustomAutoscaling.MaxCpu},
			"ram":     map[string]interface{}{"min": svc.CustomAutoscaling.MinRam, "max": svc.CustomAutoscaling.MaxRam},
			"disk":    map[string]interface{}{"min": svc.CustomAutoscaling.MinDisk, "max": svc.CustomAutoscaling.MaxDisk},
		}
	}

	if len(svc.Ports) > 0 {
		ports := make([]map[string]interface{}, len(svc.Ports))
		for i, p := range svc.Ports {
			ports[i] = map[string]interface{}{
				"port":     p.Port,
				"protocol": p.Protocol,
				"public":   p.Public,
			}
		}
		detail["ports"] = ports
	}

	return detail
}
