package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewEvents creates the events command.
func NewEvents(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Project activity timeline",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			serviceFilter, _ := cmd.Flags().GetString("service")
			limit, _ := cmd.Flags().GetInt("limit")

			ctx := cmd.Context()

			// Parallel fetch: processes, app versions, services
			type procResult struct {
				events []platform.ProcessEvent
				err    error
			}
			type avResult struct {
				events []platform.AppVersionEvent
				err    error
			}
			type svcResult struct {
				services []platform.ServiceStack
				err      error
			}

			prCh := make(chan procResult, 1)
			avCh := make(chan avResult, 1)
			svCh := make(chan svcResult, 1)

			go func() {
				events, err := client.SearchProcesses(ctx, creds.ProjectID, limit)
				prCh <- procResult{events, err}
			}()
			go func() {
				events, err := client.SearchAppVersions(ctx, creds.ProjectID, limit)
				avCh <- avResult{events, err}
			}()
			go func() {
				services, err := client.ListServices(ctx, creds.ProjectID)
				svCh <- svcResult{services, err}
			}()

			pr := <-prCh
			av := <-avCh
			sv := <-svCh

			if pr.err != nil {
				return output.Err(platform.ErrAPIError, pr.err.Error(), "", nil)
			}
			if av.err != nil {
				return output.Err(platform.ErrAPIError, av.err.Error(), "", nil)
			}
			if sv.err != nil {
				return output.Err(platform.ErrAPIError, sv.err.Error(), "", nil)
			}

			// Build serviceID→info map
			svcMap := buildServiceMap(sv.services)

			// Build unified timeline
			timeline := buildTimeline(pr.events, av.events, svcMap)

			// Filter by service
			if serviceFilter != "" {
				timeline = filterByService(timeline, serviceFilter)
			}

			// Sort by timestamp desc
			sort.Slice(timeline, func(i, j int) bool {
				return timeline[i].Timestamp > timeline[j].Timestamp
			})

			// Trim to limit
			if len(timeline) > limit {
				timeline = timeline[:limit]
			}

			return output.Sync(map[string]interface{}{
				"projectId": creds.ProjectID,
				"events":    timeline,
				"summary": map[string]interface{}{
					"total":     len(timeline),
					"processes": len(pr.events),
					"deploys":   len(av.events),
				},
			})
		},
	}

	cmd.Flags().String("service", "", "Filter by service hostname")
	cmd.Flags().Int("limit", 50, "Max events")

	return cmd
}

type serviceInfo struct {
	hostname    string
	serviceType string
}

func buildServiceMap(services []platform.ServiceStack) map[string]serviceInfo {
	m := make(map[string]serviceInfo, len(services))
	for _, s := range services {
		m[s.ID] = serviceInfo{
			hostname:    s.Name,
			serviceType: s.ServiceStackTypeInfo.ServiceStackTypeVersionName,
		}
	}
	return m
}

// TimelineEvent is a unified activity event for JSON output.
type TimelineEvent struct {
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type"`
	Action      string `json:"action"`
	Status      string `json:"status"`
	Service     string `json:"service"`
	ServiceType string `json:"serviceType,omitempty"`
	Detail      string `json:"detail,omitempty"`
	Duration    string `json:"duration,omitempty"`
	User        string `json:"user,omitempty"`
	ProcessID   string `json:"processId,omitempty"`
}

func buildTimeline(
	processes []platform.ProcessEvent,
	appVersions []platform.AppVersionEvent,
	svcMap map[string]serviceInfo,
) []TimelineEvent {
	events := make([]TimelineEvent, 0, len(processes)+len(appVersions))

	for _, p := range processes {
		hostname := ""
		svcType := ""
		if len(p.ServiceStacks) > 0 {
			ref := p.ServiceStacks[0]
			if info, ok := svcMap[ref.ID]; ok {
				hostname = info.hostname
				svcType = info.serviceType
			} else {
				hostname = ref.Name
			}
		}

		user := "system"
		if p.CreatedByUser != nil && p.CreatedByUser.Email != "" {
			user = p.CreatedByUser.Email
		}

		events = append(events, TimelineEvent{
			Timestamp:   p.Created,
			Type:        "process",
			Action:      mapActionName(p.ActionName),
			Status:      p.Status,
			Service:     hostname,
			ServiceType: svcType,
			Detail:      buildProcessDetail(p),
			Duration:    calcDuration(p.Started, p.Finished),
			User:        user,
			ProcessID:   p.ID,
		})
	}

	for _, av := range appVersions {
		hostname := ""
		svcType := ""
		if info, ok := svcMap[av.ServiceStackID]; ok {
			hostname = info.hostname
			svcType = info.serviceType
		}

		action := "deploy"
		detail := fmt.Sprintf("Deploy v%d from %s", av.Sequence, av.Source)
		if av.Build != nil && av.Build.PipelineStart != nil {
			action = "build"
			detail = fmt.Sprintf("Build v%d from %s", av.Sequence, av.Source)
		}

		duration := ""
		if av.Build != nil {
			duration = calcDuration(av.Build.PipelineStart, av.Build.PipelineFinish)
		}

		events = append(events, TimelineEvent{
			Timestamp:   av.Created,
			Type:        action,
			Action:      action,
			Status:      av.Status,
			Service:     hostname,
			ServiceType: svcType,
			Detail:      detail,
			Duration:    duration,
		})
	}

	return events
}

func filterByService(events []TimelineEvent, hostname string) []TimelineEvent {
	filtered := make([]TimelineEvent, 0)
	for _, e := range events {
		if e.Service == hostname {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

var actionNameMap = map[string]string{
	"serviceStackStart":                  "start",
	"serviceStackStop":                   "stop",
	"serviceStackRestart":                "restart",
	"serviceStackAutoscaling":            "scale",
	"serviceStackImport":                 "import",
	"serviceStackDelete":                 "delete",
	"serviceStackUserDataFile":           "env-update",
	"serviceStackEnableSubdomainAccess":  "subdomain-enable",
	"serviceStackDisableSubdomainAccess": "subdomain-disable",
}

func mapActionName(name string) string {
	if mapped, ok := actionNameMap[name]; ok {
		return mapped
	}
	return name
}

func buildProcessDetail(p platform.ProcessEvent) string {
	action := mapActionName(p.ActionName)
	hostname := ""
	if len(p.ServiceStacks) > 0 {
		hostname = p.ServiceStacks[0].Name
	}
	if hostname != "" {
		return fmt.Sprintf("%s %s", strings.Title(action), hostname) //nolint:staticcheck
	}
	return strings.Title(action) //nolint:staticcheck
}

func calcDuration(started, finished *string) string {
	if started == nil || finished == nil {
		return ""
	}
	s, err := time.Parse(time.RFC3339, *started)
	if err != nil {
		return ""
	}
	f, err := time.Parse(time.RFC3339, *finished)
	if err != nil {
		return ""
	}
	d := f.Sub(s)
	if d < 0 {
		return ""
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
