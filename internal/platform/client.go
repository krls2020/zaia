package platform

import (
	"context"
	"time"
)

// Client is the interface for Zerops API operations.
// Mocked in tests, real implementation wraps zerops-go SDK.
type Client interface {
	// Auth
	GetUserInfo(ctx context.Context) (*UserInfo, error)

	// Project discovery
	ListProjects(ctx context.Context, clientID string) ([]Project, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)

	// Service discovery
	ListServices(ctx context.Context, projectID string) ([]ServiceStack, error)
	GetService(ctx context.Context, serviceID string) (*ServiceStack, error)

	// Service management (async — return process)
	StartService(ctx context.Context, serviceID string) (*Process, error)
	StopService(ctx context.Context, serviceID string) (*Process, error)
	RestartService(ctx context.Context, serviceID string) (*Process, error)
	// SetAutoscaling returns *Process which MAY be nil (API: ResponseProcessNil).
	// When process == nil → treat as sync (scaling applied immediately).
	// When process != nil → treat as async (track via process ID).
	SetAutoscaling(ctx context.Context, serviceID string, params AutoscalingParams) (*Process, error)

	// Environment variables
	GetServiceEnv(ctx context.Context, serviceID string) ([]EnvVar, error)
	SetServiceEnvFile(ctx context.Context, serviceID string, content string) (*Process, error)
	DeleteUserData(ctx context.Context, userDataID string) (*Process, error)
	GetProjectEnv(ctx context.Context, projectID string) ([]EnvVar, error)
	CreateProjectEnv(ctx context.Context, projectID string, key, content string, sensitive bool) (*Process, error)
	DeleteProjectEnv(ctx context.Context, envID string) (*Process, error)

	// Import
	ImportServices(ctx context.Context, projectID string, yaml string) (*ImportResult, error)

	// Delete
	DeleteService(ctx context.Context, serviceID string) (*Process, error)

	// Process
	GetProcess(ctx context.Context, processID string) (*Process, error)
	CancelProcess(ctx context.Context, processID string) (*Process, error)

	// Subdomain
	EnableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error)
	DisableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error)

	// Logs (2-step: get access URL, then fetch from log backend)
	GetProjectLog(ctx context.Context, projectID string) (*LogAccess, error)

	// Activity
	SearchProcesses(ctx context.Context, projectID string, limit int) ([]ProcessEvent, error)
	SearchAppVersions(ctx context.Context, projectID string, limit int) ([]AppVersionEvent, error)
}

// LogFetcher fetches logs from the log backend (step 2).
// Separate interface because it's an HTTP call to a different service.
type LogFetcher interface {
	FetchLogs(ctx context.Context, logAccess *LogAccess, params LogFetchParams) ([]LogEntry, error)
}

// DefaultAPITimeout is the global timeout for each API call.
const DefaultAPITimeout = 30 * time.Second

// UserInfo contains user details from auth/info endpoint.
type UserInfo struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

// Project represents a Zerops project.
type Project struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ServiceStack represents a Zerops service.
type ServiceStack struct {
	ID                   string             `json:"id"`
	Name                 string             `json:"name"` // hostname
	ProjectID            string             `json:"projectId"`
	ServiceStackTypeInfo ServiceTypeInfo    `json:"serviceStackTypeInfo"`
	Status               string             `json:"status"`
	Mode                 string             `json:"mode"` // HA, NON_HA
	Ports                []Port             `json:"ports,omitempty"`
	CustomAutoscaling    *CustomAutoscaling `json:"customAutoscaling,omitempty"`
	Created              string             `json:"created"`
	LastUpdate           string             `json:"lastUpdate,omitempty"`
}

// ServiceTypeInfo contains service type details.
type ServiceTypeInfo struct {
	ServiceStackTypeVersionName string `json:"serviceStackTypeVersionName"` // e.g. "nodejs@22"
}

// Port represents a service port.
type Port struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Public   bool   `json:"public"`
}

// CustomAutoscaling contains scaling configuration.
type CustomAutoscaling struct {
	HorizontalMinCount int32   `json:"horizontalMinCount"`
	HorizontalMaxCount int32   `json:"horizontalMaxCount"`
	CpuMode            string  `json:"cpuMode"` // SHARED, DEDICATED
	StartCpuCoreCount  int32   `json:"startCpuCoreCount"`
	MinCpu             int32   `json:"minCpu"`
	MaxCpu             int32   `json:"maxCpu"`
	MinRam             float64 `json:"minRam"`
	MaxRam             float64 `json:"maxRam"`
	MinDisk            float64 `json:"minDisk"`
	MaxDisk            float64 `json:"maxDisk"`
}

// AutoscalingParams maps CLI flags to API request.
type AutoscalingParams struct {
	HorizontalMinCount  *int32
	HorizontalMaxCount  *int32
	VerticalCpuMode     *string
	VerticalStartCpu    *int32
	VerticalMinCpu      *int32
	VerticalMaxCpu      *int32
	VerticalMinRam      *float64
	VerticalMaxRam      *float64
	VerticalMinDisk     *float64
	VerticalMaxDisk     *float64
	VerticalSwapEnabled *bool
}

// Process represents an async operation tracked by Zerops.
type Process struct {
	ID            string            `json:"id"`
	ActionName    string            `json:"actionName"`
	Status        string            `json:"status"` // PENDING, RUNNING, DONE, FAILED, CANCELLED
	ServiceStacks []ServiceStackRef `json:"serviceStacks,omitempty"`
	Created       string            `json:"created"`
	Started       *string           `json:"started,omitempty"`
	Finished      *string           `json:"finished,omitempty"`
	FailReason    *string           `json:"failReason,omitempty"`
}

// ServiceStackRef is a lightweight service reference in a process.
type ServiceStackRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EnvVar represents an environment variable.
type EnvVar struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Content string `json:"content"`
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	ProjectID     string                 `json:"projectId"`
	ProjectName   string                 `json:"projectName"`
	ServiceStacks []ImportedServiceStack `json:"serviceStacks"`
}

// ImportedServiceStack represents one imported service.
type ImportedServiceStack struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Processes []Process `json:"processes,omitempty"`
	Error     *APIError `json:"error,omitempty"`
}

// APIError represents an error from the Zerops API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

// LogAccess contains temporary credentials for log backend access.
type LogAccess struct {
	AccessToken string `json:"accessToken"`
	Expiration  string `json:"expiration"`
	URL         string `json:"url"`
	URLPlain    string `json:"urlPlain"`
}

// LogFetchParams contains parameters for fetching logs from the backend.
type LogFetchParams struct {
	ServiceID string
	Severity  string // error, warning, info, debug, all
	Since     time.Time
	Limit     int
	Search    string
}

// LogEntry represents a single log entry.
type LogEntry struct {
	ID        string `json:"id,omitempty"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Container string `json:"container,omitempty"`
}

// ProcessEvent represents a process from the search API (activity timeline).
type ProcessEvent struct {
	ID              string            `json:"id"`
	ProjectID       string            `json:"projectId"`
	ServiceStacks   []ServiceStackRef `json:"serviceStacks,omitempty"`
	ActionName      string            `json:"actionName"`
	Status          string            `json:"status"`
	Created         string            `json:"created"`
	Started         *string           `json:"started,omitempty"`
	Finished        *string           `json:"finished,omitempty"`
	CreatedByUser   *UserRef          `json:"createdByUser,omitempty"`
	CreatedBySystem bool              `json:"createdBySystem"`
}

// AppVersionEvent represents a build/deploy event from the search API.
type AppVersionEvent struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"projectId"`
	ServiceStackID string     `json:"serviceStackId"`
	Source         string     `json:"source"`
	Status         string     `json:"status"`
	Sequence       int        `json:"sequence"`
	Build          *BuildInfo `json:"build,omitempty"`
	Created        string     `json:"created"`
	LastUpdate     string     `json:"lastUpdate"`
}

// BuildInfo contains build pipeline timing.
type BuildInfo struct {
	PipelineStart  *string `json:"pipelineStart,omitempty"`
	PipelineFinish *string `json:"pipelineFinish,omitempty"`
	PipelineFailed *string `json:"pipelineFailed,omitempty"`
}

// UserRef is a lightweight user reference.
type UserRef struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}
