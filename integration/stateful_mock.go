package integration

import (
	"context"
	"fmt"
	"sync"

	"github.com/zeropsio/zaia/internal/platform"
)

const statusStopped = "STOPPED"

// Compile-time interface check.
var _ platform.Client = (*StatefulMock)(nil)

// StatefulMock wraps platform.Mock but tracks mutations so that write operations
// (start, stop, env set, delete, import) affect subsequent read operations.
type StatefulMock struct {
	mu sync.RWMutex

	userInfo   *platform.UserInfo
	projects   []platform.Project
	project    *platform.Project
	services   []platform.ServiceStack
	envVars    map[string][]platform.EnvVar // serviceID -> vars
	projectEnv []platform.EnvVar
	processes  map[string]*platform.Process
	logAccess  *platform.LogAccess

	// Error overrides
	errors map[string]error

	// Auto-increment counters for unique IDs
	processCounter int
	envCounter     int
	serviceCounter int
}

// NewStatefulMock creates a new StatefulMock.
func NewStatefulMock() *StatefulMock {
	return &StatefulMock{
		envVars:   make(map[string][]platform.EnvVar),
		processes: make(map[string]*platform.Process),
		errors:    make(map[string]error),
	}
}

// --- Configuration methods (Builder pattern) ---

func (m *StatefulMock) WithUserInfo(info *platform.UserInfo) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userInfo = info
	return m
}

func (m *StatefulMock) WithProjects(projects []platform.Project) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects = projects
	return m
}

func (m *StatefulMock) WithProject(project *platform.Project) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.project = project
	return m
}

func (m *StatefulMock) WithServices(services []platform.ServiceStack) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services = services
	return m
}

func (m *StatefulMock) WithServiceEnv(serviceID string, vars []platform.EnvVar) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.envVars[serviceID] = vars
	return m
}

func (m *StatefulMock) WithProjectEnv(vars []platform.EnvVar) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projectEnv = vars
	return m
}

func (m *StatefulMock) WithProcess(p *platform.Process) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processes[p.ID] = p
	return m
}

func (m *StatefulMock) WithLogAccess(access *platform.LogAccess) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logAccess = access
	return m
}

func (m *StatefulMock) WithError(method string, err error) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
	return m
}

func (m *StatefulMock) ClearError(method string) *StatefulMock {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.errors, method)
	return m
}

// --- Helper methods ---

func (m *StatefulMock) getError(method string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errors[method]
}

func (m *StatefulMock) nextProcessID(action string) string {
	m.processCounter++
	return fmt.Sprintf("proc-%s-%d", action, m.processCounter)
}

func (m *StatefulMock) nextEnvID() string {
	m.envCounter++
	return fmt.Sprintf("env-%d", m.envCounter)
}

func (m *StatefulMock) nextServiceID() string {
	m.serviceCounter++
	return fmt.Sprintf("svc-imported-%d", m.serviceCounter)
}

func (m *StatefulMock) makeProcess(action, serviceID string) *platform.Process {
	id := m.nextProcessID(action)
	p := &platform.Process{
		ID:         id,
		ActionName: action,
		Status:     "DONE", // auto-complete for stateful testing
	}
	if serviceID != "" {
		p.ServiceStacks = []platform.ServiceStackRef{{ID: serviceID}}
	}
	m.processes[id] = p
	return p
}

func (m *StatefulMock) findServiceByID(id string) *platform.ServiceStack {
	for i := range m.services {
		if m.services[i].ID == id {
			return &m.services[i]
		}
	}
	return nil
}

// Services returns the current list of services (for test assertions).
func (m *StatefulMock) Services() []platform.ServiceStack {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]platform.ServiceStack, len(m.services))
	copy(result, m.services)
	return result
}

// --- platform.Client implementation ---

func (m *StatefulMock) GetUserInfo(_ context.Context) (*platform.UserInfo, error) {
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

func (m *StatefulMock) ListProjects(_ context.Context, _ string) ([]platform.Project, error) {
	if err := m.getError("ListProjects"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.projects, nil
}

func (m *StatefulMock) GetProject(_ context.Context, projectID string) (*platform.Project, error) {
	if err := m.getError("GetProject"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.project != nil {
		return m.project, nil
	}
	for i := range m.projects {
		if m.projects[i].ID == projectID {
			return &m.projects[i], nil
		}
	}
	return nil, fmt.Errorf("mock: project %s not found", projectID)
}

func (m *StatefulMock) ListServices(_ context.Context, _ string) ([]platform.ServiceStack, error) {
	if err := m.getError("ListServices"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]platform.ServiceStack, len(m.services))
	copy(result, m.services)
	return result, nil
}

func (m *StatefulMock) GetService(_ context.Context, serviceID string) (*platform.ServiceStack, error) {
	if err := m.getError("GetService"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := range m.services {
		if m.services[i].ID == serviceID {
			return &m.services[i], nil
		}
	}
	return nil, fmt.Errorf("mock: service %s not found", serviceID)
}

func (m *StatefulMock) StartService(_ context.Context, serviceID string) (*platform.Process, error) {
	if err := m.getError("StartService"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	// Mutate: set service status to ACTIVE
	if svc := m.findServiceByID(serviceID); svc != nil {
		svc.Status = "ACTIVE"
	}
	return m.makeProcess("start", serviceID), nil
}

func (m *StatefulMock) StopService(_ context.Context, serviceID string) (*platform.Process, error) {
	if err := m.getError("StopService"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if svc := m.findServiceByID(serviceID); svc != nil {
		svc.Status = statusStopped
	}
	return m.makeProcess("stop", serviceID), nil
}

func (m *StatefulMock) RestartService(_ context.Context, serviceID string) (*platform.Process, error) {
	if err := m.getError("RestartService"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.makeProcess("restart", serviceID), nil
}

func (m *StatefulMock) SetAutoscaling(_ context.Context, serviceID string, _ platform.AutoscalingParams) (*platform.Process, error) {
	if err := m.getError("SetAutoscaling"); err != nil {
		return nil, err
	}
	// Sync operation (applied immediately) — no process to track, matching real Mock behavior.
	return nil, nil //nolint:nilnil // intentional: nil process means sync (no async process)
}

func (m *StatefulMock) GetServiceEnv(_ context.Context, serviceID string) ([]platform.EnvVar, error) {
	if err := m.getError("GetServiceEnv"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.envVars[serviceID], nil
}

func (m *StatefulMock) SetServiceEnvFile(_ context.Context, serviceID string, content string) (*platform.Process, error) {
	if err := m.getError("SetServiceEnvFile"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse KEY=value lines and merge into env vars
	existing := m.envVars[serviceID]
	existingMap := make(map[string]int) // key -> index
	for i, e := range existing {
		existingMap[e.Key] = i
	}

	lines := splitLines(content)
	for _, line := range lines {
		if line == "" {
			continue
		}
		key, value := splitEnvLine(line)
		if key == "" {
			continue
		}
		if idx, ok := existingMap[key]; ok {
			existing[idx].Content = value
		} else {
			existing = append(existing, platform.EnvVar{
				ID:      m.nextEnvID(),
				Key:     key,
				Content: value,
			})
		}
	}
	m.envVars[serviceID] = existing

	return m.makeProcess("envSet", serviceID), nil
}

func (m *StatefulMock) DeleteUserData(_ context.Context, userDataID string) (*platform.Process, error) {
	if err := m.getError("DeleteUserData"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove env var by ID from all services
	for svcID, vars := range m.envVars {
		for i, v := range vars {
			if v.ID == userDataID {
				m.envVars[svcID] = append(vars[:i], vars[i+1:]...)
				break
			}
		}
	}

	return m.makeProcess("envDelete", ""), nil
}

func (m *StatefulMock) GetProjectEnv(_ context.Context, _ string) ([]platform.EnvVar, error) {
	if err := m.getError("GetProjectEnv"); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.projectEnv, nil
}

func (m *StatefulMock) CreateProjectEnv(_ context.Context, _ string, key, content string, _ bool) (*platform.Process, error) {
	if err := m.getError("CreateProjectEnv"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projectEnv = append(m.projectEnv, platform.EnvVar{
		ID:      m.nextEnvID(),
		Key:     key,
		Content: content,
	})
	return m.makeProcess("envSet", ""), nil
}

func (m *StatefulMock) DeleteProjectEnv(_ context.Context, envID string) (*platform.Process, error) {
	if err := m.getError("DeleteProjectEnv"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, v := range m.projectEnv {
		if v.ID == envID {
			m.projectEnv = append(m.projectEnv[:i], m.projectEnv[i+1:]...)
			break
		}
	}
	return m.makeProcess("envDelete", ""), nil
}

func (m *StatefulMock) ImportServices(_ context.Context, _ string, _ string) (*platform.ImportResult, error) {
	if err := m.getError("ImportServices"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a new service for each import
	newID := m.nextServiceID()
	newSvc := platform.ServiceStack{
		ID:     newID,
		Name:   "imported-service",
		Status: "ACTIVE",
	}
	m.services = append(m.services, newSvc)

	proc := m.makeProcess("import", newID)
	return &platform.ImportResult{
		ServiceStacks: []platform.ImportedServiceStack{
			{
				ID:        newID,
				Name:      "imported-service",
				Processes: []platform.Process{*proc},
			},
		},
	}, nil
}

func (m *StatefulMock) DeleteService(_ context.Context, serviceID string) (*platform.Process, error) {
	if err := m.getError("DeleteService"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove service from list
	for i, s := range m.services {
		if s.ID == serviceID {
			m.services = append(m.services[:i], m.services[i+1:]...)
			break
		}
	}

	return m.makeProcess("delete", serviceID), nil
}

func (m *StatefulMock) GetProcess(_ context.Context, processID string) (*platform.Process, error) {
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

func (m *StatefulMock) CancelProcess(_ context.Context, processID string) (*platform.Process, error) {
	if err := m.getError("CancelProcess"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.processes[processID]
	if !ok {
		return nil, fmt.Errorf("mock: process %s not found", processID)
	}
	p.Status = "CANCELLED"
	return p, nil
}

func (m *StatefulMock) EnableSubdomainAccess(_ context.Context, serviceID string) (*platform.Process, error) {
	if err := m.getError("EnableSubdomainAccess"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.makeProcess("enableSubdomain", serviceID), nil
}

func (m *StatefulMock) DisableSubdomainAccess(_ context.Context, serviceID string) (*platform.Process, error) {
	if err := m.getError("DisableSubdomainAccess"); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.makeProcess("disableSubdomain", serviceID), nil
}

func (m *StatefulMock) GetProjectLog(_ context.Context, _ string) (*platform.LogAccess, error) {
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

// --- helpers ---

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitEnvLine(line string) (string, string) {
	for i := 0; i < len(line); i++ {
		if line[i] == '=' {
			return line[:i], line[i+1:]
		}
	}
	return line, ""
}
