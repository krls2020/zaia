package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
)

func TestValidateCmd_ValidZeropsYml(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `zerops:
  - run:
      base: nodejs@22
      ports:
        - port: 3000
          httpSupport: true
`
	yamlPath := filepath.Join(dir, "zerops.yml")
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	cmd := NewValidate()
	cmd.SetArgs([]string{"--file", yamlPath})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	if data["valid"] != true {
		t.Errorf("valid = %v, want true", data["valid"])
	}
}

func TestValidateCmd_InvalidZeropsYml(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `somethingWrong: yes`
	yamlPath := filepath.Join(dir, "zerops.yml")
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	cmd := NewValidate()
	cmd.SetArgs([]string{"--file", yamlPath, "--type", "zerops.yml"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid zerops.yml")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "INVALID_ZEROPS_YML" {
		t.Errorf("code = %v, want INVALID_ZEROPS_YML", resp["code"])
	}
}

func TestValidateCmd_ValidImportYml(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `services:
  - hostname: api
    type: nodejs@22
  - hostname: db
    type: postgresql@16
`
	yamlPath := filepath.Join(dir, "import.yml")
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	cmd := NewValidate()
	cmd.SetArgs([]string{"--file", yamlPath})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["valid"] != true {
		t.Errorf("valid = %v, want true", data["valid"])
	}
	if data["type"] != "import.yml" {
		t.Errorf("type = %v, want import.yml", data["type"])
	}
}

func TestValidateCmd_ImportWithProject(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `project:
  name: my-app
services:
  - hostname: api
    type: nodejs@22
`
	yamlPath := filepath.Join(dir, "import.yml")
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	cmd := NewValidate()
	cmd.SetArgs([]string{"--file", yamlPath})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for import with project:")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "IMPORT_HAS_PROJECT" {
		t.Errorf("code = %v, want IMPORT_HAS_PROJECT", resp["code"])
	}
}

func TestValidateCmd_InlineContent(t *testing.T) {
	content := `zerops:
  - run:
      base: nodejs@22
`
	cmd := NewValidate()
	cmd.SetArgs([]string{"--content", content, "--type", "zerops.yml"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["valid"] != true {
		t.Errorf("valid = %v, want true", data["valid"])
	}
}

func TestValidateCmd_FileNotFound(t *testing.T) {
	cmd := NewValidate()
	cmd.SetArgs([]string{"--file", "/nonexistent/zerops.yml"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "FILE_NOT_FOUND" {
		t.Errorf("code = %v, want FILE_NOT_FOUND", resp["code"])
	}
}

func TestValidateCmd_InvalidYamlSyntax(t *testing.T) {
	content := `{invalid yaml[`
	cmd := NewValidate()
	cmd.SetArgs([]string{"--content", content, "--type", "zerops.yml"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid YAML syntax")
	}
}
