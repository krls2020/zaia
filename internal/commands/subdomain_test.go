package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestSubdomainCmd_Enable(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewSubdomain(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api", "--action", "enable"})

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
}

func TestSubdomainCmd_Disable(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}})

	cmd := NewSubdomain(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api", "--action", "disable"})

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
}

func TestSubdomainCmd_AlreadyEnabled(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{{ID: "s1", Name: "api", Status: "ACTIVE"}}).
		WithError("EnableSubdomainAccess", fmt.Errorf("AlreadyEnabled"))

	cmd := NewSubdomain(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api", "--action", "enable"})

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
		t.Errorf("type = %v, want sync (idempotent)", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	if data["status"] != "already_enabled" {
		t.Errorf("status = %v, want already_enabled", data["status"])
	}
}
