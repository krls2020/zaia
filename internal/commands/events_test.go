package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

func TestEvents_BasicTimeline(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)

	started := "2026-01-15T10:00:00Z"
	finished := "2026-01-15T10:02:15Z"
	pipelineStart := "2026-01-15T09:50:00Z"
	pipelineFinish := "2026-01-15T09:55:30Z"

	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{
			{ID: "s1", Name: "api", ProjectID: "proj-1", ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "nodejs@22"}},
		}).
		WithProcessEvents([]platform.ProcessEvent{
			{
				ID:            "p1",
				ProjectID:     "proj-1",
				ServiceStacks: []platform.ServiceStackRef{{ID: "s1", Name: "api"}},
				ActionName:    "serviceStackRestart",
				Status:        "FINISHED",
				Created:       "2026-01-15T10:00:00Z",
				Started:       &started,
				Finished:      &finished,
				CreatedByUser: &platform.UserRef{FullName: "John", Email: "john@test.com"},
			},
		}).
		WithAppVersionEvents([]platform.AppVersionEvent{
			{
				ID:             "av1",
				ProjectID:      "proj-1",
				ServiceStackID: "s1",
				Source:         "CLI",
				Status:         "ACTIVE",
				Sequence:       3,
				Build:          &platform.BuildInfo{PipelineStart: &pipelineStart, PipelineFinish: &pipelineFinish},
				Created:        "2026-01-15T09:50:00Z",
				LastUpdate:     "2026-01-15T09:55:30Z",
			},
		})

	cmd := NewEvents(storagePath, mock)
	cmd.SetArgs([]string{})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	if err := cmd.Execute(); err != nil {
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
	events := data["events"].([]interface{})
	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}

	// Events should be sorted desc by timestamp — process (10:00) first, then build (09:50)
	ev0 := events[0].(map[string]interface{})
	if ev0["action"] != "restart" {
		t.Errorf("events[0].action = %v, want restart", ev0["action"])
	}
	if ev0["service"] != "api" {
		t.Errorf("events[0].service = %v, want api", ev0["service"])
	}
	if ev0["duration"] != "2m15s" {
		t.Errorf("events[0].duration = %v, want 2m15s", ev0["duration"])
	}
	if ev0["user"] != "john@test.com" {
		t.Errorf("events[0].user = %v, want john@test.com", ev0["user"])
	}

	ev1 := events[1].(map[string]interface{})
	if ev1["action"] != "build" {
		t.Errorf("events[1].action = %v, want build", ev1["action"])
	}
	if ev1["duration"] != "5m30s" {
		t.Errorf("events[1].duration = %v, want 5m30s", ev1["duration"])
	}

	summary := data["summary"].(map[string]interface{})
	if summary["total"] != float64(2) {
		t.Errorf("summary.total = %v, want 2", summary["total"])
	}
}

func TestEvents_FilterByService(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)

	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{
			{ID: "s1", Name: "api", ProjectID: "proj-1", ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "nodejs@22"}},
			{ID: "s2", Name: "db", ProjectID: "proj-1", ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "postgresql@16"}},
		}).
		WithProcessEvents([]platform.ProcessEvent{
			{ID: "p1", ProjectID: "proj-1", ServiceStacks: []platform.ServiceStackRef{{ID: "s1", Name: "api"}}, ActionName: "serviceStackRestart", Status: "FINISHED", Created: "2026-01-15T10:00:00Z"},
			{ID: "p2", ProjectID: "proj-1", ServiceStacks: []platform.ServiceStackRef{{ID: "s2", Name: "db"}}, ActionName: "serviceStackStart", Status: "FINISHED", Created: "2026-01-15T09:00:00Z"},
		}).
		WithAppVersionEvents(nil)

	cmd := NewEvents(storagePath, mock)
	cmd.SetArgs([]string{"--service", "api"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	events := data["events"].([]interface{})
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1 (filtered to api)", len(events))
	}
	ev := events[0].(map[string]interface{})
	if ev["service"] != "api" {
		t.Errorf("service = %v, want api", ev["service"])
	}
}

func TestEvents_EmptyProject(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)

	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{}).
		WithProcessEvents(nil).
		WithAppVersionEvents(nil)

	cmd := NewEvents(storagePath, mock)
	cmd.SetArgs([]string{})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	events := data["events"].([]interface{})
	if len(events) != 0 {
		t.Errorf("events len = %d, want 0", len(events))
	}
}

func TestEvents_NotAuthenticated(t *testing.T) {
	dir := t.TempDir()
	storagePath := filepath.Join(dir, "zaia.data")
	mock := platform.NewMock()

	cmd := NewEvents(storagePath, mock)
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

func TestEvents_LimitFlag(t *testing.T) {
	storagePath := setupAuthenticatedStorage(t)

	processes := make([]platform.ProcessEvent, 10)
	for i := 0; i < 10; i++ {
		processes[i] = platform.ProcessEvent{
			ID:            fmt.Sprintf("p%d", i),
			ProjectID:     "proj-1",
			ServiceStacks: []platform.ServiceStackRef{{ID: "s1", Name: "api"}},
			ActionName:    "serviceStackRestart",
			Status:        "FINISHED",
			Created:       fmt.Sprintf("2026-01-15T10:%02d:00Z", i),
		}
	}

	mock := platform.NewMock().
		WithServices([]platform.ServiceStack{
			{ID: "s1", Name: "api", ProjectID: "proj-1"},
		}).
		WithProcessEvents(processes).
		WithAppVersionEvents(nil)

	cmd := NewEvents(storagePath, mock)
	cmd.SetArgs([]string{"--limit", "3"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	events := data["events"].([]interface{})
	if len(events) != 3 {
		t.Errorf("events len = %d, want 3", len(events))
	}
}

func TestEvents_DurationCalculation(t *testing.T) {
	tests := []struct {
		name     string
		started  *string
		finished *string
		want     string
	}{
		{"nil started", nil, strPtr("2026-01-15T10:00:00Z"), ""},
		{"nil finished", strPtr("2026-01-15T10:00:00Z"), nil, ""},
		{"30 seconds", strPtr("2026-01-15T10:00:00Z"), strPtr("2026-01-15T10:00:30Z"), "30s"},
		{"2m15s", strPtr("2026-01-15T10:00:00Z"), strPtr("2026-01-15T10:02:15Z"), "2m15s"},
		{"1h5m", strPtr("2026-01-15T10:00:00Z"), strPtr("2026-01-15T11:05:00Z"), "1h5m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcDuration(tt.started, tt.finished)
			if got != tt.want {
				t.Errorf("calcDuration() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEvents_ActionNameMapping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"serviceStackStart", "start"},
		{"serviceStackStop", "stop"},
		{"serviceStackRestart", "restart"},
		{"serviceStackAutoscaling", "scale"},
		{"serviceStackImport", "import"},
		{"serviceStackDelete", "delete"},
		{"serviceStackUserDataFile", "env-update"},
		{"serviceStackEnableSubdomainAccess", "subdomain-enable"},
		{"serviceStackDisableSubdomainAccess", "subdomain-disable"},
		{"unknownAction", "unknownAction"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapActionName(tt.input)
			if got != tt.want {
				t.Errorf("mapActionName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
