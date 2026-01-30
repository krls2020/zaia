package commands

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
	"gopkg.in/yaml.v3"
)

// NewValidate creates the validate command.
func NewValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate YAML configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			file, _ := cmd.Flags().GetString("file")
			contentStr, _ := cmd.Flags().GetString("content")
			fileType, _ := cmd.Flags().GetString("type")

			var content []byte
			var source string

			if contentStr != "" {
				content = []byte(contentStr)
				source = "inline"
			} else {
				if file == "" {
					file = "zerops.yml"
				}
				var err error
				content, err = os.ReadFile(file)
				if err != nil {
					if os.IsNotExist(err) && file == "zerops.yml" {
						return output.Err(platform.ErrZeropsYmlNotFound,
							"zerops.yml not found in current directory",
							"Create zerops.yml or use --file to specify path", nil)
					}
					return output.Err(platform.ErrFileNotFound,
						"Cannot read file: "+file, "", nil)
				}
				source = file
			}

			if fileType == "" {
				fileType = detectYamlType(source, content)
			}

			switch fileType {
			case "zerops.yml":
				return validateZeropsYml(content, source)
			case "import.yml":
				return validateImportYml(content, source)
			default:
				return output.Err(platform.ErrUnknownType,
					"Unknown file type: "+fileType,
					"Specify --type zerops.yml or --type import.yml", nil)
			}
		},
	}

	cmd.Flags().String("file", "", "File to validate (default: zerops.yml)")
	cmd.Flags().String("content", "", "Inline YAML content to validate")
	cmd.Flags().String("type", "", "File type: zerops.yml or import.yml")

	return cmd
}

func detectYamlType(source string, content []byte) string {
	// Detect by filename
	if strings.Contains(source, "import") {
		return "import.yml"
	}

	// Detect by content structure
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return "zerops.yml" // default
	}

	if _, ok := parsed["services"]; ok {
		return "import.yml"
	}
	if _, ok := parsed["zerops"]; ok {
		return "zerops.yml"
	}

	return "zerops.yml" // default
}

func validateZeropsYml(content []byte, source string) error {
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return output.Err(platform.ErrInvalidZeropsYml,
			"zerops.yml validation failed",
			"",
			map[string]interface{}{
				"file": source,
				"errors": []map[string]string{
					{"path": "", "error": "Invalid YAML syntax: " + err.Error(), "fix": "Check YAML formatting"},
				},
			})
	}

	// Basic structural validation
	var errors []map[string]string

	zeropsRaw, ok := parsed["zerops"]
	if !ok {
		errors = append(errors, map[string]string{
			"path":  "",
			"error": "Missing 'zerops' key",
			"fix":   "Add: zerops:\n  - run:\n      base: nodejs@22",
		})
	} else {
		zeropsArr, ok := zeropsRaw.([]interface{})
		if !ok {
			errors = append(errors, map[string]string{
				"path":  "zerops",
				"error": "'zerops' must be an array",
				"fix":   "Format: zerops:\n  - run:\n      base: nodejs@22",
			})
		} else if len(zeropsArr) == 0 {
			errors = append(errors, map[string]string{
				"path":  "zerops",
				"error": "'zerops' array is empty",
				"fix":   "Add at least one service configuration",
			})
		}
	}

	if len(errors) > 0 {
		return output.Err(platform.ErrInvalidZeropsYml,
			"zerops.yml validation failed",
			"",
			map[string]interface{}{
				"file":   source,
				"errors": errors,
			})
	}

	return output.Sync(map[string]interface{}{
		"valid":    true,
		"file":     source,
		"type":     "zerops.yml",
		"warnings": []string{},
		"info":     []string{},
	})
}

func validateImportYml(content []byte, source string) error {
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		return output.Err(platform.ErrInvalidImportYml,
			"import.yml validation failed",
			"",
			map[string]interface{}{
				"file": source,
				"errors": []map[string]string{
					{"path": "", "error": "Invalid YAML syntax: " + err.Error(), "fix": "Check YAML formatting"},
				},
			})
	}

	var errors []map[string]string

	// Check for project: section (not allowed in project-scoped context)
	if _, ok := parsed["project"]; ok {
		return output.Err(platform.ErrImportHasProject,
			"import.yml must not contain 'project:' section in project-scoped context",
			"Remove the 'project:' section. ZAIA imports services into the current project context.",
			map[string]interface{}{"file": source})
	}

	// Check services array
	if _, ok := parsed["services"]; !ok {
		errors = append(errors, map[string]string{
			"path":  "",
			"error": "Missing 'services' key",
			"fix":   "Add: services:\n  - hostname: api\n    type: nodejs@22",
		})
	}

	if len(errors) > 0 {
		return output.Err(platform.ErrInvalidImportYml,
			"import.yml validation failed",
			"",
			map[string]interface{}{
				"file":   source,
				"errors": errors,
			})
	}

	return output.Sync(map[string]interface{}{
		"valid":    true,
		"file":     source,
		"type":     "import.yml",
		"warnings": []string{},
		"info":     []string{},
	})
}
