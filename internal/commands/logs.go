package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// NewLogs creates the logs command.
func NewLogs(storagePath string, client platform.Client, fetcher platform.LogFetcher) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Fetch service logs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			hostname, _ := cmd.Flags().GetString("service")
			severity, _ := cmd.Flags().GetString("severity")
			since, _ := cmd.Flags().GetString("since")
			limit, _ := cmd.Flags().GetInt("limit")
			search, _ := cmd.Flags().GetString("search")

			if hostname == "" {
				return output.Err(platform.ErrServiceRequired,
					"--service flag is required",
					"Run: zaia logs --service <hostname>", nil)
			}

			sinceTime, err := parseSince(since)
			if err != nil {
				return output.Err(platform.ErrInvalidParameter,
					fmt.Sprintf("Invalid --since format: %s", since),
					"Use: 30m, 1h, 24h, 7d, or ISO 8601 timestamp", nil)
			}

			ctx := cmd.Context()

			// Resolve service
			services, err := client.ListServices(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}
			svc, err := resolveServiceID(client, creds.ProjectID, hostname, services)
			if err != nil {
				return err
			}

			// Step 1: Get log access
			logAccess, err := client.GetProjectLog(ctx, creds.ProjectID)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			// Step 2: Fetch from log backend
			entries, err := fetcher.FetchLogs(ctx, logAccess, platform.LogFetchParams{
				ServiceID: svc.ID,
				Severity:  severity,
				Since:     sinceTime,
				Limit:     limit,
				Search:    search,
			})
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			logEntries := make([]map[string]interface{}, len(entries))
			for i, e := range entries {
				entry := map[string]interface{}{
					"timestamp": e.Timestamp,
					"severity":  e.Severity,
					"message":   e.Message,
				}
				if e.Container != "" {
					entry["container"] = e.Container
				}
				logEntries[i] = entry
			}

			return output.Sync(map[string]interface{}{
				"entries": logEntries,
				"hasMore": len(entries) >= limit,
			})
		},
	}

	cmd.Flags().String("service", "", "Service hostname (required)")
	cmd.Flags().String("severity", "all", "Filter: error, warning, info, debug, all")
	cmd.Flags().String("since", "1h", "Duration (30m, 1h, 24h, 7d) or ISO 8601 timestamp")
	cmd.Flags().Int("limit", 100, "Max entries")
	cmd.Flags().String("search", "", "Text search")
	cmd.Flags().String("build", "", "Build ID for build logs")

	return cmd
}

var durationRegex = regexp.MustCompile(`^(\d+)(m|h|d)$`)

func parseSince(s string) (time.Time, error) {
	if s == "" {
		return time.Now().Add(-1 * time.Hour), nil
	}

	// Try duration format (30m, 1h, 24h, 7d)
	matches := durationRegex.FindStringSubmatch(s)
	if len(matches) == 3 {
		n, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration number: %s", s)
		}
		switch matches[2] {
		case "m":
			if n < 1 || n > 1440 {
				return time.Time{}, fmt.Errorf("minutes must be 1-1440")
			}
			return time.Now().Add(-time.Duration(n) * time.Minute), nil
		case "h":
			if n < 1 || n > 168 {
				return time.Time{}, fmt.Errorf("hours must be 1-168")
			}
			return time.Now().Add(-time.Duration(n) * time.Hour), nil
		case "d":
			if n < 1 || n > 30 {
				return time.Time{}, fmt.Errorf("days must be 1-30")
			}
			return time.Now().Add(-time.Duration(n) * 24 * time.Hour), nil
		}
	}

	// Try ISO 8601
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid format: %s", s)
}
