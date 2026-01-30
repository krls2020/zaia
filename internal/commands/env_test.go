package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestEnvGet_Service(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}}).
		WithServiceEnv("s1", []platform.EnvVar{
			{ID: "e1", Key: "DB_HOST", Content: "db"},
			{ID: "e2", Key: "PORT", Content: "3000"},
		})

	cmd := NewEnv(storagePath, mock)
	cmd.SetArgs([]string{"get", "--service", "api"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["scope"] != "service" {
		t.Errorf("scope = %v, want service", data["scope"])
	}
	vars := data["vars"].([]interface{})
	if len(vars) != 2 {
		t.Errorf("vars len = %d, want 2", len(vars))
	}
}

func TestEnvGet_Project(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithProjectEnv([]platform.EnvVar{
			{ID: "pe1", Key: "SHARED", Content: "value"},
		})

	cmd := NewEnv(storagePath, mock)
	cmd.SetArgs([]string{"get", "--project"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["scope"] != "project" {
		t.Errorf("scope = %v, want project", data["scope"])
	}
}

func TestEnvSet_Service(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewEnv(storagePath, mock)
	cmd.SetArgs([]string{"set", "--service", "api", "DB_HOST=db", "PORT=3000"})

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

func TestEnvSet_EmptyValue(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewEnv(storagePath, mock)
	cmd.SetArgs([]string{"set", "--service", "api", "KEY="})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal("KEY= with empty value should be valid")
	}
}

func TestEnvSet_ValueWithEquals(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewEnv(storagePath, mock)
	cmd.SetArgs([]string{"set", "--service", "api", "KEY=val=ue"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal("KEY=val=ue should split on first = only")
	}
}

func TestEnvSet_InvalidFormat(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()

	cmd := NewEnv(storagePath, mock)
	cmd.SetArgs([]string{"set", "--service", "api", "NOEQUALS"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format")
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "INVALID_ENV_FORMAT" {
		t.Errorf("code = %v, want INVALID_ENV_FORMAT", resp["code"])
	}
}

func TestEnvDelete_Service(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}}).
		WithServiceEnv("s1", []platform.EnvVar{
			{ID: "e1", Key: "DB_HOST", Content: "db"},
		})

	cmd := NewEnv(storagePath, mock)
	cmd.SetArgs([]string{"delete", "--service", "api", "DB_HOST"})

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

func TestEnvGet_NoScope_Error(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()

	cmd := NewEnv(storagePath, mock)
	cmd.SetArgs([]string{"get"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when neither --service nor --project is set")
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "SERVICE_REQUIRED" {
		t.Errorf("code = %v, want SERVICE_REQUIRED", resp["code"])
	}
}
