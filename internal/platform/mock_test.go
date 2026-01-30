package platform

import (
	"fmt"
	"testing"
)

func TestMock_ImplementsClient(t *testing.T) {
	// Compile-time check is in mock.go, but let's also verify at runtime
	var _ Client = NewMock()
}

func TestMock_FluentAPI(t *testing.T) {
	mock := NewMock().
		WithUserInfo(&UserInfo{ID: "u1", FullName: "John", Email: "john@test.com"}).
		WithProjects([]Project{{ID: "p1", Name: "my-app"}}).
		WithServices([]ServiceStack{
			{ID: "s1", Name: "api", Status: "ACTIVE"},
			{ID: "s2", Name: "db", Status: "ACTIVE"},
		})

	ctx := t.Context()

	user, err := mock.GetUserInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if user.FullName != "John" {
		t.Errorf("fullName = %q, want John", user.FullName)
	}

	projects, err := mock.ListProjects(ctx, "client-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("projects len = %d, want 1", len(projects))
	}

	services, err := mock.ListServices(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 2 {
		t.Fatalf("services len = %d, want 2", len(services))
	}
}

func TestMock_ErrorOverride(t *testing.T) {
	mock := NewMock().
		WithUserInfo(&UserInfo{ID: "u1", FullName: "John", Email: "john@test.com"}).
		WithError("GetUserInfo", fmt.Errorf("auth failed"))

	ctx := t.Context()
	_, err := mock.GetUserInfo(ctx)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "auth failed" {
		t.Errorf("error = %q, want 'auth failed'", err.Error())
	}
}

func TestMock_ProcessLifecycle(t *testing.T) {
	mock := NewMock().
		WithProcess(&Process{
			ID:         "proc-1",
			ActionName: "restart",
			Status:     "RUNNING",
		})

	ctx := t.Context()

	// Get process
	p, err := mock.GetProcess(ctx, "proc-1")
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "RUNNING" {
		t.Errorf("status = %q, want RUNNING", p.Status)
	}

	// Cancel process
	p, err = mock.CancelProcess(ctx, "proc-1")
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != "CANCELLED" {
		t.Errorf("status = %q, want CANCELLED", p.Status)
	}
}

func TestMock_ProcessNotFound(t *testing.T) {
	mock := NewMock()
	ctx := t.Context()

	_, err := mock.GetProcess(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent process")
	}
}

func TestMock_ServiceManagement(t *testing.T) {
	mock := NewMock()
	ctx := t.Context()

	tests := []struct {
		name   string
		fn     func() (*Process, error)
		action string
	}{
		{"start", func() (*Process, error) { return mock.StartService(ctx, "svc-1") }, "start"},
		{"stop", func() (*Process, error) { return mock.StopService(ctx, "svc-1") }, "stop"},
		{"restart", func() (*Process, error) { return mock.RestartService(ctx, "svc-1") }, "restart"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := tt.fn()
			if err != nil {
				t.Fatal(err)
			}
			if p.ActionName != tt.action {
				t.Errorf("actionName = %q, want %q", p.ActionName, tt.action)
			}
			if p.Status != "PENDING" {
				t.Errorf("status = %q, want PENDING", p.Status)
			}
		})
	}
}

func TestMock_SetAutoscaling_ReturnsNil(t *testing.T) {
	mock := NewMock()
	ctx := t.Context()

	p, err := mock.SetAutoscaling(ctx, "svc-1", AutoscalingParams{})
	if err != nil {
		t.Fatal(err)
	}
	if p != nil {
		t.Error("expected nil process for SetAutoscaling (sync scaling)")
	}
}

func TestMock_EnvVars(t *testing.T) {
	mock := NewMock().
		WithServiceEnv("svc-1", []EnvVar{
			{ID: "e1", Key: "DB_HOST", Content: "db"},
			{ID: "e2", Key: "PORT", Content: "3000"},
		}).
		WithProjectEnv([]EnvVar{
			{ID: "pe1", Key: "SHARED", Content: "value"},
		})

	ctx := t.Context()

	vars, err := mock.GetServiceEnv(ctx, "svc-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(vars) != 2 {
		t.Errorf("service env len = %d, want 2", len(vars))
	}

	projVars, err := mock.GetProjectEnv(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(projVars) != 1 {
		t.Errorf("project env len = %d, want 1", len(projVars))
	}
}

func TestMock_GetServiceByID(t *testing.T) {
	mock := NewMock().
		WithServices([]ServiceStack{
			{ID: "svc-1", Name: "api"},
			{ID: "svc-2", Name: "db"},
		})

	ctx := t.Context()

	svc, err := mock.GetService(ctx, "svc-2")
	if err != nil {
		t.Fatal(err)
	}
	if svc.Name != "db" {
		t.Errorf("name = %q, want db", svc.Name)
	}

	_, err = mock.GetService(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent service")
	}
}
