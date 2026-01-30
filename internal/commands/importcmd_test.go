package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestImportCmd_Success(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithImportResult(&platform.ImportResult{
			ProjectID:   "proj-1",
			ProjectName: "my-app",
			ServiceStacks: []platform.ImportedServiceStack{
				{
					ID:   "s1",
					Name: "api",
					Processes: []platform.Process{
						{ID: "proc-1", ActionName: "serviceStackCreate", Status: "PENDING"},
					},
				},
				{
					ID:   "s2",
					Name: "db",
					Processes: []platform.Process{
						{ID: "proc-2", ActionName: "serviceStackCreate", Status: "PENDING"},
					},
				},
			},
		})

	content := `services:
  - hostname: api
    type: nodejs@22
  - hostname: db
    type: postgresql@16
`
	cmd := NewImport(storagePath, mock)
	cmd.SetArgs([]string{"--content", content})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "async" {
		t.Errorf("type = %v, want async", resp["type"])
	}
	procs := resp["processes"].([]interface{})
	if len(procs) != 2 {
		t.Errorf("processes len = %d, want 2", len(procs))
	}
}

func TestImportCmd_DryRun(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()

	content := `services:
  - hostname: api
    type: nodejs@22
`
	cmd := NewImport(storagePath, mock)
	cmd.SetArgs([]string{"--content", content, "--dry-run"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	if data["dryRun"] != true {
		t.Errorf("dryRun = %v, want true", data["dryRun"])
	}
}

func TestImportCmd_HasProject_Error(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()

	content := `project:
  name: my-app
services:
  - hostname: api
    type: nodejs@22
`
	cmd := NewImport(storagePath, mock)
	cmd.SetArgs([]string{"--content", content})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for project: section")
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "IMPORT_HAS_PROJECT" {
		t.Errorf("code = %v, want IMPORT_HAS_PROJECT", resp["code"])
	}
}

func TestImportCmd_FromFile(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithImportResult(&platform.ImportResult{
			ProjectID: "proj-1",
			ServiceStacks: []platform.ImportedServiceStack{
				{
					ID:   "s1",
					Name: "api",
					Processes: []platform.Process{
						{ID: "proc-1", ActionName: "serviceStackCreate", Status: "PENDING"},
					},
				},
			},
		})

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "services.yml")
	os.WriteFile(yamlPath, []byte("services:\n  - hostname: api\n    type: nodejs@22\n"), 0644)

	cmd := NewImport(storagePath, mock)
	cmd.SetArgs([]string{"--file", yamlPath})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "async" {
		t.Errorf("type = %v, want async", resp["type"])
	}
}

func TestImportCmd_NoInput_Error(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()

	cmd := NewImport(storagePath, mock)
	cmd.SetArgs([]string{})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when neither --file nor --content")
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "INVALID_PARAMETER" {
		t.Errorf("code = %v, want INVALID_PARAMETER", resp["code"])
	}
}
