package commands

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/zeropsio/zaia/internal/platform"
)

func TestLoginCmd_SingleProject_Success(t *testing.T) {
	dir := t.TempDir()
	storagePath := filepath.Join(dir, "zaia.data")

	mock := platform.NewMock().
		WithUserInfo(&platform.UserInfo{ID: "u1", FullName: "John Doe", Email: "john@test.com"}).
		WithProjects([]platform.Project{{ID: "abc-123", Name: "my-app"}})

	cmd := NewLogin(storagePath, func(_, _ string) platform.Client { return mock })

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--token", "test-token", "--url", "api.zerops.io"})

	// Capture output
	var stdout bytes.Buffer
	origWriter := getWriter()
	setWriter(&stdout)
	defer setWriter(origWriter)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, stdout.String())
	}

	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	project := data["project"].(map[string]interface{})
	if project["id"] != "abc-123" {
		t.Errorf("project.id = %v, want abc-123", project["id"])
	}
}

func TestLoginCmd_NoToken_ReturnsUsageError(t *testing.T) {
	cmd := NewLogin("", func(_, _ string) platform.Client { return platform.NewMock() })
	cmd.SetArgs([]string{}) // no token

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestLoginCmd_MultiProject_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	storagePath := filepath.Join(dir, "zaia.data")

	mock := platform.NewMock().
		WithUserInfo(&platform.UserInfo{ID: "u1", FullName: "John", Email: "john@test.com"}).
		WithProjects([]platform.Project{
			{ID: "p1", Name: "app1"},
			{ID: "p2", Name: "app2"},
		})

	cmd := NewLogin(storagePath, func(_, _ string) platform.Client { return mock })
	cmd.SetArgs([]string{"--token", "token", "--url", "api.zerops.io"})

	var stdout bytes.Buffer
	origWriter := getWriter()
	setWriter(&stdout)
	defer setWriter(origWriter)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for multi-project")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, stdout.String())
	}
	if resp["code"] != "TOKEN_MULTI_PROJECT" {
		t.Errorf("code = %v, want TOKEN_MULTI_PROJECT", resp["code"])
	}
}
