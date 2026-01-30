package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestDeleteCmd_WithConfirm(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewDelete(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api", "--confirm"})

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
	p := procs[0].(map[string]interface{})
	if p["actionName"] != "delete" {
		t.Errorf("actionName = %v, want delete", p["actionName"])
	}
}

func TestDeleteCmd_WithoutConfirm(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock()

	cmd := NewDelete(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error without --confirm")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "CONFIRM_REQUIRED" {
		t.Errorf("code = %v, want CONFIRM_REQUIRED", resp["code"])
	}
	ctx := resp["context"].(map[string]interface{})
	wouldDelete := ctx["wouldDelete"].(map[string]interface{})
	if wouldDelete["hostname"] != "api" {
		t.Errorf("wouldDelete.hostname = %v, want api", wouldDelete["hostname"])
	}
}
