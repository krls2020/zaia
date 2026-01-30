package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestProcessCmd_Running(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithProcess(&platform.Process{
			ID:         "proc-1",
			ActionName: "restart",
			Status:     "RUNNING",
			ServiceStacks: []platform.ServiceStackRef{{ID: "s1", Name: "api"}},
			Created:    "2026-01-29T10:00:00Z",
		})

	cmd := NewProcess(storagePath, mock)
	cmd.SetArgs([]string{"proc-1"})

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
	if data["status"] != "RUNNING" {
		t.Errorf("status = %v, want RUNNING", data["status"])
	}
	if data["processId"] != "proc-1" {
		t.Errorf("processId = %v, want proc-1", data["processId"])
	}
}

func TestProcessCmd_Finished(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	finished := "2026-01-29T10:00:12Z"
	mock := platform.NewMock().
		WithProcess(&platform.Process{
			ID:         "proc-1",
			ActionName: "restart",
			Status:     "DONE",
			Created:    "2026-01-29T10:00:00Z",
			Finished:   &finished,
		})

	cmd := NewProcess(storagePath, mock)
	cmd.SetArgs([]string{"proc-1"})

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
	if data["status"] != "FINISHED" {
		t.Errorf("status = %v, want FINISHED (mapped from DONE)", data["status"])
	}
}

func TestProcessCmd_NotFound(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()

	cmd := NewProcess(storagePath, mock)
	cmd.SetArgs([]string{"nonexistent"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "PROCESS_NOT_FOUND" {
		t.Errorf("code = %v, want PROCESS_NOT_FOUND", resp["code"])
	}
}

func TestCancelCmd_Success(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithProcess(&platform.Process{
			ID:         "proc-1",
			ActionName: "restart",
			Status:     "RUNNING",
		})

	cmd := NewCancel(storagePath, mock)
	cmd.SetArgs([]string{"proc-1"})

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
	if data["status"] != "CANCELED" {
		t.Errorf("status = %v, want CANCELED", data["status"])
	}
}

func TestCancelCmd_AlreadyFinished(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	finished := "2026-01-29T10:00:12Z"
	mock := platform.NewMock().
		WithProcess(&platform.Process{
			ID:         "proc-1",
			ActionName: "restart",
			Status:     "DONE",
			Finished:   &finished,
		})

	cmd := NewCancel(storagePath, mock)
	cmd.SetArgs([]string{"proc-1"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for already terminal process")
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "PROCESS_ALREADY_TERMINAL" {
		t.Errorf("code = %v, want PROCESS_ALREADY_TERMINAL", resp["code"])
	}
}

func TestCancelCmd_ProcessNotFound(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()

	cmd := NewCancel(storagePath, mock)
	cmd.SetArgs([]string{"nonexistent"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}

	var resp map[string]interface{}
	json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "PROCESS_NOT_FOUND" {
		t.Errorf("code = %v, want PROCESS_NOT_FOUND", resp["code"])
	}
}
