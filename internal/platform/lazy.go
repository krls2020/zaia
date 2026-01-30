package platform

import (
	"context"
	"fmt"
	"sync"
)

// LazyClient implements Client by creating a real ZeropsClient on first use.
// It reads credentials from a resolver function and caches the client.
type LazyClient struct {
	mu       sync.Mutex
	client   Client
	resolver func() (token, apiHost string, err error)
}

// NewLazyClient creates a Client that defers ZeropsClient creation until first API call.
// The resolver function should return (token, apiHost, error) by reading stored credentials.
func NewLazyClient(resolver func() (token, apiHost string, err error)) *LazyClient {
	return &LazyClient{resolver: resolver}
}

// Compile-time check.
var _ Client = (*LazyClient)(nil)

func (l *LazyClient) init() (Client, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.client != nil {
		return l.client, nil
	}
	token, apiHost, err := l.resolver()
	if err != nil {
		return nil, err
	}
	c, err := NewZeropsClient(token, apiHost)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}
	l.client = c
	return l.client, nil
}

func (l *LazyClient) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.GetUserInfo(ctx)
}

func (l *LazyClient) ListProjects(ctx context.Context, clientID string) ([]Project, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.ListProjects(ctx, clientID)
}

func (l *LazyClient) GetProject(ctx context.Context, projectID string) (*Project, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.GetProject(ctx, projectID)
}

func (l *LazyClient) ListServices(ctx context.Context, projectID string) ([]ServiceStack, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.ListServices(ctx, projectID)
}

func (l *LazyClient) GetService(ctx context.Context, serviceID string) (*ServiceStack, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.GetService(ctx, serviceID)
}

func (l *LazyClient) StartService(ctx context.Context, serviceID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.StartService(ctx, serviceID)
}

func (l *LazyClient) StopService(ctx context.Context, serviceID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.StopService(ctx, serviceID)
}

func (l *LazyClient) RestartService(ctx context.Context, serviceID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.RestartService(ctx, serviceID)
}

func (l *LazyClient) SetAutoscaling(ctx context.Context, serviceID string, params AutoscalingParams) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.SetAutoscaling(ctx, serviceID, params)
}

func (l *LazyClient) GetServiceEnv(ctx context.Context, serviceID string) ([]EnvVar, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.GetServiceEnv(ctx, serviceID)
}

func (l *LazyClient) SetServiceEnvFile(ctx context.Context, serviceID string, content string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.SetServiceEnvFile(ctx, serviceID, content)
}

func (l *LazyClient) DeleteUserData(ctx context.Context, userDataID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.DeleteUserData(ctx, userDataID)
}

func (l *LazyClient) GetProjectEnv(ctx context.Context, projectID string) ([]EnvVar, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.GetProjectEnv(ctx, projectID)
}

func (l *LazyClient) CreateProjectEnv(ctx context.Context, projectID string, key, content string, sensitive bool) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.CreateProjectEnv(ctx, projectID, key, content, sensitive)
}

func (l *LazyClient) DeleteProjectEnv(ctx context.Context, envID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.DeleteProjectEnv(ctx, envID)
}

func (l *LazyClient) ImportServices(ctx context.Context, projectID string, yaml string) (*ImportResult, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.ImportServices(ctx, projectID, yaml)
}

func (l *LazyClient) DeleteService(ctx context.Context, serviceID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.DeleteService(ctx, serviceID)
}

func (l *LazyClient) GetProcess(ctx context.Context, processID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.GetProcess(ctx, processID)
}

func (l *LazyClient) CancelProcess(ctx context.Context, processID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.CancelProcess(ctx, processID)
}

func (l *LazyClient) EnableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.EnableSubdomainAccess(ctx, serviceID)
}

func (l *LazyClient) DisableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.DisableSubdomainAccess(ctx, serviceID)
}

func (l *LazyClient) GetProjectLog(ctx context.Context, projectID string) (*LogAccess, error) {
	c, err := l.init()
	if err != nil {
		return nil, err
	}
	return c.GetProjectLog(ctx, projectID)
}
