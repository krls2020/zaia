package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
	"gopkg.in/yaml.v3"
)

// NewImport creates the import command.
func NewImport(storagePath string, client platform.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import services from YAML",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := resolveCredentials(storagePath)
			if err != nil {
				return err
			}

			file, _ := cmd.Flags().GetString("file")
			contentStr, _ := cmd.Flags().GetString("content")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			var content string
			if contentStr != "" {
				content = contentStr
			} else if file != "" {
				data, err := os.ReadFile(file)
				if err != nil {
					return output.Err(platform.ErrFileNotFound,
						"Cannot read file: "+file, "", nil)
				}
				content = string(data)
			} else {
				return output.Err(platform.ErrInvalidParameter,
					"--file or --content is required",
					"Run: zaia import --file services.yml or zaia import --content '<yaml>'", nil)
			}

			// Check for project: section
			if hasProjectSection(content) {
				return output.Err(platform.ErrImportHasProject,
					"import.yml must not contain 'project:' section in project-scoped context",
					"Remove the 'project:' section. ZAIA imports services into the current project context. Only 'services:' array is expected.",
					map[string]interface{}{"projectId": creds.ProjectID})
			}

			if dryRun {
				// For dry-run, just parse and return preview
				return importDryRun(content)
			}

			ctx := cmd.Context()
			result, err := client.ImportServices(ctx, creds.ProjectID, content)
			if err != nil {
				return output.Err(platform.ErrAPIError, err.Error(), "", nil)
			}

			// Extract processes from nested serviceStacks
			var processes []output.ProcessOutput
			for _, ss := range result.ServiceStacks {
				for _, p := range ss.Processes {
					proc := output.MapProcessToOutput(&p, ss.Name)
					proc.ActionName = "import"
					proc.ServiceID = ss.ID
					processes = append(processes, proc)
				}
			}

			return output.Async(processes)
		},
	}

	cmd.Flags().String("file", "", "Path to YAML file")
	cmd.Flags().String("content", "", "Inline YAML content")
	cmd.Flags().Bool("dry-run", false, "Validate and preview without executing")

	return cmd
}

func hasProjectSection(content string) bool {
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		return false
	}
	_, hasProject := parsed["project"]
	return hasProject
}

func importDryRun(content string) error {
	var parsed map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		return output.Err(platform.ErrInvalidImportYml,
			"Invalid YAML syntax: "+err.Error(), "", nil)
	}

	servicesRaw, ok := parsed["services"]
	if !ok {
		return output.Err(platform.ErrInvalidImportYml,
			"Missing 'services' section",
			"import.yml must contain a 'services:' array", nil)
	}

	servicesList, ok := servicesRaw.([]interface{})
	if !ok {
		return output.Err(platform.ErrInvalidImportYml,
			"'services' must be an array",
			"Format: services:\n  - hostname: api\n    type: nodejs@22", nil)
	}

	preview := make([]map[string]interface{}, 0, len(servicesList))
	for _, s := range servicesList {
		sMap, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		entry := map[string]interface{}{
			"action": "create",
		}
		if h, ok := sMap["hostname"]; ok {
			entry["hostname"] = h
		}
		if t, ok := sMap["type"]; ok {
			entry["type"] = t
		}
		// Also check for "name" as hostname alias
		if h, ok := sMap["name"]; ok && entry["hostname"] == nil {
			entry["hostname"] = h
		}
		preview = append(preview, entry)
	}

	return output.Sync(map[string]interface{}{
		"dryRun":   true,
		"valid":    true,
		"services": preview,
		"warnings": []string{},
	})
}
