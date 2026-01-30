package platform

import (
	"context"
	"fmt"
	"sync"
)

const statusCancelled = "CANCELLED"

// Compile-time interface check
var _ Client = (*Mock)(nil)

// Mock is a configurable mock for the Platform Client interface.
type Mock struct {
	mu sync.RWMutex

	userInfo     *UserInfo
	projects     []Project
	project      *Project
	services     []ServiceStack
	service      *ServiceStack
	processes    map[string]*Process
	envVars      map[string][]EnvVar // serviceID -> env vars
	projectEnv   []EnvVar
	logAccess    *LogAccess
	importResult *ImportResult

	// Error overrides: method name -> error
	errors map[string]error
}

// NewMock creates a new configurable mock.
func NewMock() *Mock {
	return &Mock{
		processes: make(map[string]*Process),
		envVars:   make(map[string][]EnvVar),
		errors:    make(map[string]error),
	}
}

// WithUserInfo sets the user info returned by GetUserInfo.
func (m *Mock) WithUserInfo(info *UserInfo) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userInfo = info
	return m
}

// WithProjects sets the projects returned by ListProjects.
func (m *Mock) WithProjects(projects []Project) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects = projects
	return m
}

// WithProject sets the project returned by GetProject.
func (m *Mock) WithProject(project *Project) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.project = project
	return m
}

// WithServices sets the services returned by ListServices.
func (m *Mock) WithServices(services []ServiceStack) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services = services
	return m
}

// WithService sets the service returned by GetService.
func (m *Mock) WithService(service *ServiceStack) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.service = service
	return m
}

// WithProcess adds a process to the mock.
func (m *Mock) WithProcess(process *Process) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processes[process.ID] = process
	return m
}

// WithServiceEnv sets env vars for a service.
func (m *Mock) WithServiceEnv(serviceID string, vars []EnvVar) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.envVars[serviceID] = vars
	return m
}

// WithProjectEnv sets project-level env vars.
func (m *Mock) WithProjectEnv(vars []EnvVar) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projectEnv = vars
	return m
}

// WithLogAccess sets the log access returned by GetProjectLog.
func (m *Mock) WithLogAccess(access *LogAccess) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logAccess = access
	return m
}

// WithImportResult sets the result returned by ImportServices.
func (m *Mock) WithImportResult(result *ImportResult) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.importResult = result
	return m
}

// WithError sets an error for a specific method.
func (m *Mock) WithError(method string, err error) *Mock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
	return m
}

func (m *Mock) getError(method string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errors[method]
}

func (m *Mock) GetUserInfo(_ context.Context) (*UserInfo, error) {
	if err := m.getError("GetUserInfo"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.userInfo == nil {
		return nil, fmt.Errorf("mock: no user info configured")
	}
	return m.userInfo, nil
}

func (m *Mock) ListProjects(_ context.Context, _ string) ([]Project, error) {
	if err := m.getError("ListProjects"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.projects, nil
}

func (m *Mock) GetProject(_ context.Context, _ string) (*Project, error) {
	if err := m.getError("GetProject"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.project == nil {
		return nil, fmt.Errorf("mock: no project configured")
	}
	return m.project, nil
}

func (m *Mock) ListServices(_ context.Context, _ string) ([]ServiceStack, error) {
	if err := m.getError("ListServices"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.services, nil
}

func (m *Mock) GetService(_ context.Context, serviceID string) (*ServiceStack, error) {
	if err := m.getError("GetService"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.service != nil {
		return m.service, nil
	}
	// Look up in services list
	for i := range m.services {
		if m.services[i].ID == serviceID {
			return &m.services[i], nil
		}
	}
	return nil, fmt.Errorf("mock: service %s not found", serviceID)
}

func (m *Mock) StartService(_ context.Context, serviceID string) (*Process, error) {
	if err := m.getError("StartService"); err != nil {
		return nil, err
	}
	return &Process{
		ID:            "proc-start-" + serviceID,
		ActionName:    "start",
		Status:        "PENDING",
		ServiceStacks: []ServiceStackRef{{ID: serviceID}},
	}, nil
}

func (m *Mock) StopService(_ context.Context, serviceID string) (*Process, error) {
	if err := m.getError("StopService"); err != nil {
		return nil, err
	}
	return &Process{
		ID:            "proc-stop-" + serviceID,
		ActionName:    "stop",
		Status:        "PENDING",
		ServiceStacks: []ServiceStackRef{{ID: serviceID}},
	}, nil
}

func (m *Mock) RestartService(_ context.Context, serviceID string) (*Process, error) {
	if err := m.getError("RestartService"); err != nil {
		return nil, err
	}
	return &Process{
		ID:            "proc-restart-" + serviceID,
		ActionName:    "restart",
		Status:        "PENDING",
		ServiceStacks: []ServiceStackRef{{ID: serviceID}},
	}, nil
}

func (m *Mock) SetAutoscaling(_ context.Context, serviceID string, _ AutoscalingParams) (*Process, error) {
	if err := m.getError("SetAutoscaling"); err != nil {
		return nil, err
	}
	// Sync operation (applied immediately) — no process to track.
	return nil, nil //nolint:nilnil // intentional: nil process means sync (no async process)
}

func (m *Mock) GetServiceEnv(_ context.Context, serviceID string) ([]EnvVar, error) {
	if err := m.getError("GetServiceEnv"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.envVars[serviceID], nil
}

func (m *Mock) SetServiceEnvFile(_ context.Context, serviceID string, _ string) (*Process, error) {
	if err := m.getError("SetServiceEnvFile"); err != nil {
		return nil, err
	}
	return &Process{
		ID:            "proc-envset-" + serviceID,
		ActionName:    "envSet",
		Status:        "PENDING",
		ServiceStacks: []ServiceStackRef{{ID: serviceID}},
	}, nil
}

func (m *Mock) DeleteUserData(_ context.Context, userDataID string) (*Process, error) {
	if err := m.getError("DeleteUserData"); err != nil {
		return nil, err
	}
	return &Process{
		ID:         "proc-envdel-" + userDataID,
		ActionName: "envDelete",
		Status:     "PENDING",
	}, nil
}

func (m *Mock) GetProjectEnv(_ context.Context, _ string) ([]EnvVar, error) {
	if err := m.getError("GetProjectEnv"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.projectEnv, nil
}

func (m *Mock) CreateProjectEnv(_ context.Context, _ string, _ string, _ string, _ bool) (*Process, error) {
	if err := m.getError("CreateProjectEnv"); err != nil {
		return nil, err
	}
	return &Process{
		ID:         "proc-projenvset",
		ActionName: "envSet",
		Status:     "PENDING",
	}, nil
}

func (m *Mock) DeleteProjectEnv(_ context.Context, envID string) (*Process, error) {
	if err := m.getError("DeleteProjectEnv"); err != nil {
		return nil, err
	}
	return &Process{
		ID:         "proc-projenvdel-" + envID,
		ActionName: "envDelete",
		Status:     "PENDING",
	}, nil
}

func (m *Mock) ImportServices(_ context.Context, _ string, _ string) (*ImportResult, error) {
	if err := m.getError("ImportServices"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.importResult == nil {
		return nil, fmt.Errorf("mock: no import result configured")
	}
	return m.importResult, nil
}

func (m *Mock) DeleteService(_ context.Context, serviceID string) (*Process, error) {
	if err := m.getError("DeleteService"); err != nil {
		return nil, err
	}
	return &Process{
		ID:            "proc-delete-" + serviceID,
		ActionName:    "delete",
		Status:        "PENDING",
		ServiceStacks: []ServiceStackRef{{ID: serviceID}},
	}, nil
}

func (m *Mock) GetProcess(_ context.Context, processID string) (*Process, error) {
	if err := m.getError("GetProcess"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.processes[processID]
	if !ok {
		return nil, fmt.Errorf("mock: process %s not found", processID)
	}
	return p, nil
}

func (m *Mock) CancelProcess(_ context.Context, processID string) (*Process, error) {
	if err := m.getError("CancelProcess"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.processes[processID]
	if !ok {
		return nil, fmt.Errorf("mock: process %s not found", processID)
	}
	p.Status = statusCancelled
	return p, nil
}

func (m *Mock) EnableSubdomainAccess(_ context.Context, serviceID string) (*Process, error) {
	if err := m.getError("EnableSubdomainAccess"); err != nil {
		return nil, err
	}
	return &Process{
		ID:            "proc-subdomain-enable-" + serviceID,
		ActionName:    "enableSubdomain",
		Status:        "PENDING",
		ServiceStacks: []ServiceStackRef{{ID: serviceID}},
	}, nil
}

func (m *Mock) DisableSubdomainAccess(_ context.Context, serviceID string) (*Process, error) {
	if err := m.getError("DisableSubdomainAccess"); err != nil {
		return nil, err
	}
	return &Process{
		ID:            "proc-subdomain-disable-" + serviceID,
		ActionName:    "disableSubdomain",
		Status:        "PENDING",
		ServiceStacks: []ServiceStackRef{{ID: serviceID}},
	}, nil
}

func (m *Mock) GetProjectLog(_ context.Context, _ string) (*LogAccess, error) {
	if err := m.getError("GetProjectLog"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.logAccess == nil {
		return nil, fmt.Errorf("mock: no log access configured")
	}
	return m.logAccess, nil
}

// Compile-time interface check for LogFetcher
var _ LogFetcher = (*MockLogFetcher)(nil)

// MockLogFetcher is a configurable mock for LogFetcher.
type MockLogFetcher struct {
	entries []LogEntry
	err     error
}

// NewMockLogFetcher creates a new MockLogFetcher.
func NewMockLogFetcher() *MockLogFetcher {
	return &MockLogFetcher{}
}

// WithEntries sets the log entries to return.
func (f *MockLogFetcher) WithEntries(entries []LogEntry) *MockLogFetcher {
	f.entries = entries
	return f
}

// WithError sets the error to return.
func (f *MockLogFetcher) WithError(err error) *MockLogFetcher {
	f.err = err
	return f
}

func (f *MockLogFetcher) FetchLogs(_ context.Context, _ *LogAccess, _ LogFetchParams) ([]LogEntry, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.entries, nil
}
