package auth

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/zeropsio/zaia/internal/platform"
)

func newTestManager(t *testing.T) (*Manager, *platform.Mock) {
	t.Helper()
	dir := t.TempDir()
	storage := NewStorage(filepath.Join(dir, "zaia.data"))
	mock := platform.NewMock()
	mgr := NewManager(storage, mock)
	return mgr, mock
}

func TestLogin_SingleProject_Success(t *testing.T) {
	mgr, mock := newTestManager(t)
	mock.WithUserInfo(&platform.UserInfo{
		ID:       "u1",
		FullName: "John Doe",
		Email:    "john@example.com",
	}).WithProjects([]platform.Project{
		{ID: "abc-123", Name: "my-app", Status: "ACTIVE"},
	})

	result, err := mgr.Login(context.Background(), "valid-token", LoginOptions{
		URL: "api.zerops.io",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.User.Name != "John Doe" {
		t.Errorf("user.name = %q, want 'John Doe'", result.User.Name)
	}
	if result.Project.ID != "abc-123" {
		t.Errorf("project.id = %q, want abc-123", result.Project.ID)
	}
	if result.Project.Name != "my-app" {
		t.Errorf("project.name = %q, want my-app", result.Project.Name)
	}

	// Verify credentials are stored
	creds, err := mgr.GetCredentials()
	if err != nil {
		t.Fatal(err)
	}
	if creds.Token != "valid-token" {
		t.Errorf("stored token = %q, want valid-token", creds.Token)
	}
	if creds.ProjectID != "abc-123" {
		t.Errorf("stored projectID = %q, want abc-123", creds.ProjectID)
	}
}

func TestLogin_NoProject_ReturnsTokenNoProjectError(t *testing.T) {
	mgr, mock := newTestManager(t)
	mock.WithUserInfo(&platform.UserInfo{
		ID:       "u1",
		FullName: "John",
		Email:    "john@test.com",
	}).WithProjects([]platform.Project{})

	_, err := mgr.Login(context.Background(), "token", LoginOptions{URL: "api.zerops.io"})
	if err == nil {
		t.Fatal("expected error")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("error type = %T, want *AuthError", err)
	}
	if authErr.Code != platform.ErrTokenNoProject {
		t.Errorf("code = %q, want %q", authErr.Code, platform.ErrTokenNoProject)
	}
}

func TestLogin_MultiProject_ReturnsTokenMultiProjectError(t *testing.T) {
	mgr, mock := newTestManager(t)
	mock.WithUserInfo(&platform.UserInfo{
		ID:       "u1",
		FullName: "John",
		Email:    "john@test.com",
	}).WithProjects([]platform.Project{
		{ID: "p1", Name: "app1"},
		{ID: "p2", Name: "app2"},
		{ID: "p3", Name: "app3"},
	})

	_, err := mgr.Login(context.Background(), "token", LoginOptions{URL: "api.zerops.io"})
	if err == nil {
		t.Fatal("expected error")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("error type = %T, want *AuthError", err)
	}
	if authErr.Code != platform.ErrTokenMultiProject {
		t.Errorf("code = %q, want %q", authErr.Code, platform.ErrTokenMultiProject)
	}
	// Should mention count and project names
	if authErr.Message == "" {
		t.Error("message should not be empty")
	}
}

func TestLogin_InvalidToken_ReturnsAuthInvalidTokenError(t *testing.T) {
	mgr, mock := newTestManager(t)
	mock.WithError("GetUserInfo", &platform.HTTPError{StatusCode: 401, Body: "unauthorized"})

	_, err := mgr.Login(context.Background(), "invalid", LoginOptions{URL: "api.zerops.io"})
	if err == nil {
		t.Fatal("expected error")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("error type = %T, want *AuthError", err)
	}
	if authErr.Code != platform.ErrAuthInvalidToken {
		t.Errorf("code = %q, want %q", authErr.Code, platform.ErrAuthInvalidToken)
	}
}

func TestLogin_NetworkError_ReturnsAuthAPIError(t *testing.T) {
	mgr, mock := newTestManager(t)
	mock.WithUserInfo(&platform.UserInfo{ID: "u1", FullName: "John", Email: "john@test.com"}).
		WithError("ListProjects", &platform.HTTPError{StatusCode: 500, Body: "server error"})

	_, err := mgr.Login(context.Background(), "token", LoginOptions{URL: "api.zerops.io"})
	if err == nil {
		t.Fatal("expected error")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("error type = %T, want *AuthError", err)
	}
	if authErr.Code != platform.ErrAuthAPIError {
		t.Errorf("code = %q, want %q", authErr.Code, platform.ErrAuthAPIError)
	}
}

func TestLogout_ClearsStorage(t *testing.T) {
	mgr, mock := newTestManager(t)
	mock.WithUserInfo(&platform.UserInfo{ID: "u1", FullName: "John", Email: "john@test.com"}).
		WithProjects([]platform.Project{{ID: "p1", Name: "app"}})

	// Login first
	_, err := mgr.Login(context.Background(), "token", LoginOptions{URL: "api.zerops.io"})
	if err != nil {
		t.Fatal(err)
	}

	// Verify logged in
	_, err = mgr.GetCredentials()
	if err != nil {
		t.Fatal("should be logged in")
	}

	// Logout
	if err := mgr.Logout(); err != nil {
		t.Fatal(err)
	}

	// Verify logged out
	_, err = mgr.GetCredentials()
	if err == nil {
		t.Error("should not be authenticated after logout")
	}
}

func TestGetCredentials_NotAuthenticated(t *testing.T) {
	mgr, _ := newTestManager(t)

	_, err := mgr.GetCredentials()
	if err == nil {
		t.Fatal("expected error")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("error type = %T, want *AuthError", err)
	}
	if authErr.Code != platform.ErrAuthRequired {
		t.Errorf("code = %q, want %q", authErr.Code, platform.ErrAuthRequired)
	}
}

func TestGetStatus_Authenticated(t *testing.T) {
	mgr, mock := newTestManager(t)
	mock.WithUserInfo(&platform.UserInfo{ID: "u1", FullName: "John", Email: "john@test.com"}).
		WithProjects([]platform.Project{{ID: "p1", Name: "app"}})

	_, err := mgr.Login(context.Background(), "token", LoginOptions{URL: "api.zerops.io"})
	if err != nil {
		t.Fatal(err)
	}

	data, err := mgr.GetStatus()
	if err != nil {
		t.Fatal(err)
	}

	if data.User.Name != "John" {
		t.Errorf("user.name = %q, want John", data.User.Name)
	}
	if data.Project.ID != "p1" {
		t.Errorf("project.id = %q, want p1", data.Project.ID)
	}
}

func TestGetStatus_NotAuthenticated(t *testing.T) {
	mgr, _ := newTestManager(t)

	_, err := mgr.GetStatus()
	if err == nil {
		t.Fatal("expected error")
	}

	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("error type = %T, want *AuthError", err)
	}
	if authErr.Code != platform.ErrAuthRequired {
		t.Errorf("code = %q, want %q", authErr.Code, platform.ErrAuthRequired)
	}
}

func TestLogin_DefaultRegion(t *testing.T) {
	mgr, mock := newTestManager(t)
	mock.WithUserInfo(&platform.UserInfo{ID: "u1", FullName: "John", Email: "john@test.com"}).
		WithProjects([]platform.Project{{ID: "p1", Name: "app"}})

	// No URL specified — should use default
	result, err := mgr.Login(context.Background(), "token", LoginOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if result.Region != "prg1" {
		t.Errorf("region = %q, want prg1", result.Region)
	}
}
