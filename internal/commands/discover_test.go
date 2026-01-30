package commands

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/zeropsio/zaia/internal/auth"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func setupAuthenticatedStorage(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "zaia.data")
	storage := auth.NewStorage(path)
	_ = storage.Save(auth.Data{
		Token:   "test-token",
		APIHost: "api.zerops.io",
		Project: auth.ProjectInfo{ID: "proj-1", Name: "my-app"},
		User:    auth.UserData{Name: "John", Email: "john@test.com"},
	})
	return path
}

func TestDiscoverCmd_ListServices(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithProject(&platform.Project{ID: "proj-1", Name: "my-app", Status: "ACTIVE"}).
		WithServices([]platform.ServiceStack{
			{ID: "s1", Name: "api", Status: "ACTIVE", ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "nodejs@22"}},
			{ID: "s2", Name: "db", Status: "ACTIVE", ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "postgresql@16"}},
		})

	cmd := NewDiscover(storagePath, mock)
	cmd.SetArgs([]string{})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, stdout.String())
	}

	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	services := data["services"].([]interface{})
	if len(services) != 2 {
		t.Errorf("services len = %d, want 2", len(services))
	}

	svc0 := services[0].(map[string]interface{})
	if svc0["hostname"] != "api" {
		t.Errorf("services[0].hostname = %v, want api", svc0["hostname"])
	}
}

func TestDiscoverCmd_EmptyProject(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithProject(&platform.Project{ID: "proj-1", Name: "my-app", Status: "ACTIVE"}).
		WithServices([]platform.ServiceStack{})

	cmd := NewDiscover(storagePath, mock)
	cmd.SetArgs([]string{})

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
	services := data["services"].([]interface{})
	if len(services) != 0 {
		t.Errorf("services len = %d, want 0", len(services))
	}
}

func TestDiscoverCmd_SingleService(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithProject(&platform.Project{ID: "proj-1", Name: "my-app", Status: "ACTIVE"}).
		WithServices([]platform.ServiceStack{
			{
				ID:                   "s1",
				Name:                 "api",
				Status:               "ACTIVE",
				ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "nodejs@22"},
				CustomAutoscaling: &platform.CustomAutoscaling{
					HorizontalMinCount: 1,
					HorizontalMaxCount: 3,
					CpuMode:            "SHARED",
					MinCpu:             1,
					MaxCpu:             5,
					MinRam:             0.5,
					MaxRam:             4,
				},
				Ports:   []platform.Port{{Port: 3000, Protocol: "http", Public: true}},
				Created: "2026-01-15T10:05:00Z",
			},
		})

	cmd := NewDiscover(storagePath, mock)
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
	data := resp["data"].(map[string]interface{})
	services := data["services"].([]interface{})
	if len(services) != 1 {
		t.Fatalf("services len = %d, want 1", len(services))
	}

	svc := services[0].(map[string]interface{})
	if svc["hostname"] != "api" {
		t.Errorf("hostname = %v, want api", svc["hostname"])
	}
	if svc["resources"] == nil {
		t.Error("expected resources in detailed view")
	}
	if svc["ports"] == nil {
		t.Error("expected ports in detailed view")
	}
}

func TestDiscoverCmd_ServiceNotFound(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)
	mock := platform.NewMock().
		WithProject(&platform.Project{ID: "proj-1", Name: "my-app", Status: "ACTIVE"}).
		WithServices([]platform.ServiceStack{
			{ID: "s1", Name: "api", Status: "ACTIVE"},
		})

	cmd := NewDiscover(storagePath, mock)
	cmd.SetArgs([]string{"--service", "nonexistent"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent service")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "SERVICE_NOT_FOUND" {
		t.Errorf("code = %v, want SERVICE_NOT_FOUND", resp["code"])
	}
}

func TestDiscoverCmd_NotAuthenticated(t *testing.T) {
	dir := t.TempDir()
	storagePath := filepath.Join(dir, "zaia.data")
	mock := platform.NewMock()

	cmd := NewDiscover(storagePath, mock)
	cmd.SetArgs([]string{})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when not authenticated")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["code"] != "AUTH_REQUIRED" {
		t.Errorf("code = %v, want AUTH_REQUIRED", resp["code"])
	}
}
