package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestStartCmd_Success(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewStart(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "async" {
		t.Errorf("type = %v, want async", resp["type"])
	}
	procs := resp["processes"].([]interface{})
	if len(procs) != 1 {
		t.Fatalf("processes len = %d, want 1", len(procs))
	}
	p := procs[0].(map[string]interface{})
	if p["actionName"] != "start" {
		t.Errorf("actionName = %v, want start", p["actionName"])
	}
}

func TestStopCmd_Success(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewStop(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	p := resp["processes"].([]interface{})[0].(map[string]interface{})
	if p["actionName"] != "stop" {
		t.Errorf("actionName = %v, want stop", p["actionName"])
	}
}

func TestRestartCmd_Success(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewRestart(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	p := resp["processes"].([]interface{})[0].(map[string]interface{})
	if p["actionName"] != "restart" {
		t.Errorf("actionName = %v, want restart", p["actionName"])
	}
}

func TestScaleCmd_Success(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewScale(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api", "--min-cpu", "1", "--max-cpu", "4"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	// Mock returns nil process (sync)
	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync (nil process)", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	if data["message"] != "Scaling parameters updated" {
		t.Errorf("message = %v", data["message"])
	}
}

func TestScaleCmd_NoParams_Error(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewScale(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no scaling params")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "INVALID_SCALING" {
		t.Errorf("code = %v, want INVALID_SCALING", resp["code"])
	}
}

func TestScaleCmd_InvalidMinMax_Error(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewScale(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api", "--min-ram", "4", "--max-ram", "2"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for min > max")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "INVALID_SCALING" {
		t.Errorf("code = %v, want INVALID_SCALING", resp["code"])
	}
}

func TestLifecycleCmd_ServiceNotFound(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewRestart(storagePath, mock)
	cmd.SetArgs([]string{"--service", "nonexistent"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "SERVICE_NOT_FOUND" {
		t.Errorf("code = %v, want SERVICE_NOT_FOUND", resp["code"])
	}
}
