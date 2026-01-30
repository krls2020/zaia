package integration

import (
	"github.com/zeropsio/zaia/internal/auth"
	"github.com/zeropsio/zaia/internal/platform"
)

// FixtureUnauthenticated sets up a harness with no stored credentials.
// All commands requiring auth will fail with AUTH_REQUIRED.
func FixtureUnauthenticated(h *Harness) {
	// No-op: default state has no auth data
	h.Mock().
		WithUserInfo(&platform.UserInfo{ID: "user-1", FullName: "Test User", Email: "test@example.com"}).
		WithProjects([]platform.Project{{ID: "proj-1", Name: "test-project", Status: "ACTIVE"}}).
		WithProject(&platform.Project{ID: "proj-1", Name: "test-project", Status: "ACTIVE"})
}

// FixtureEmptyProject sets up an authenticated harness with 0 services.
func FixtureEmptyProject(h *Harness) {
	writeAuth(h)
	h.Mock().
		WithUserInfo(&platform.UserInfo{ID: "user-1", FullName: "Test User", Email: "test@example.com"}).
		WithProjects([]platform.Project{{ID: "proj-1", Name: "test-project", Status: "ACTIVE"}}).
		WithProject(&platform.Project{ID: "proj-1", Name: "test-project", Status: "ACTIVE"}).
		WithServices(nil)
}

// FixtureFullProject sets up an authenticated harness with 3 services:
//   - api (nodejs@22, ACTIVE)
//   - db (postgresql@16, ACTIVE)
//   - cache (valkey@8, ACTIVE)
func FixtureFullProject(h *Harness) {
	writeAuth(h)
	h.Mock().
		WithUserInfo(&platform.UserInfo{ID: "user-1", FullName: "Test User", Email: "test@example.com"}).
		WithProjects([]platform.Project{{ID: "proj-1", Name: "test-project", Status: "ACTIVE"}}).
		WithProject(&platform.Project{ID: "proj-1", Name: "test-project", Status: "ACTIVE"}).
		WithServices([]platform.ServiceStack{
			{
				ID:                   "svc-api",
				Name:                 "api",
				ProjectID:            "proj-1",
				ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "nodejs@22"},
				Status:               "ACTIVE",
				Created:              "2024-01-01T00:00:00Z",
			},
			{
				ID:                   "svc-db",
				Name:                 "db",
				ProjectID:            "proj-1",
				ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "postgresql@16"},
				Status:               "ACTIVE",
				Created:              "2024-01-01T00:00:00Z",
			},
			{
				ID:                   "svc-cache",
				Name:                 "cache",
				ProjectID:            "proj-1",
				ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "valkey@8"},
				Status:               "ACTIVE",
				Created:              "2024-01-01T00:00:00Z",
			},
		})
}

// FixtureStoppedService sets up an authenticated harness with 1 stopped service.
func FixtureStoppedService(h *Harness) {
	writeAuth(h)
	h.Mock().
		WithUserInfo(&platform.UserInfo{ID: "user-1", FullName: "Test User", Email: "test@example.com"}).
		WithProjects([]platform.Project{{ID: "proj-1", Name: "test-project", Status: "ACTIVE"}}).
		WithProject(&platform.Project{ID: "proj-1", Name: "test-project", Status: "ACTIVE"}).
		WithServices([]platform.ServiceStack{
			{
				ID:                   "svc-api",
				Name:                 "api",
				ProjectID:            "proj-1",
				ServiceStackTypeInfo: platform.ServiceTypeInfo{ServiceStackTypeVersionName: "nodejs@22"},
				Status:               statusStopped,
				Created:              "2024-01-01T00:00:00Z",
			},
		})
}

// writeAuth creates a valid zaia.data file in the harness's storage path.
func writeAuth(h *Harness) {
	storage := auth.NewStorage(h.storagePath)
	err := storage.Save(auth.Data{
		Token:   "test-token",
		APIHost: "https://api.zerops.io",
		RegionData: auth.RegionItem{
			Name:    "prg1",
			Address: "https://api.zerops.io",
		},
		Project: auth.ProjectInfo{
			ID:   "proj-1",
			Name: "test-project",
		},
		User: auth.UserData{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	if err != nil {
		h.t.Fatalf("Failed to write auth fixture: %v", err)
	}
}
