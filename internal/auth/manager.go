package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeropsio/zaia/internal/platform"
)

// Credentials contains the active authentication context.
type Credentials struct {
	Token       string
	APIHost     string
	ProjectID   string
	ProjectName string
	Region      string
}

// LoginOptions contains optional login flags.
type LoginOptions struct {
	URL       string // Direct API URL (skips region discovery)
	Region    string // Region name
	RegionURL string // Custom region metadata URL
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	User    UserData    `json:"user"`
	Project ProjectInfo `json:"project"`
	Region  string      `json:"region"`
}

// Manager handles authentication operations.
type Manager struct {
	storage *Storage
	client  platform.Client
}

// NewManager creates a new auth manager.
func NewManager(storage *Storage, client platform.Client) *Manager {
	return &Manager{
		storage: storage,
		client:  client,
	}
}

// Login authenticates and auto-discovers the project from the token.
func (m *Manager) Login(ctx context.Context, token string, opts LoginOptions) (*LoginResult, error) {
	// 1. Resolve API host
	apiHost := opts.URL
	region := opts.Region
	if apiHost == "" {
		// For now, use default. Region discovery will be implemented later.
		apiHost = "api.app-prg1.zerops.io"
		if region == "" {
			region = "prg1"
		}
	}

	// 2. Validate token (GetUserInfo)
	user, err := m.client.GetUserInfo(ctx)
	if err != nil {
		return nil, &AuthError{
			Code:       platform.ErrAuthInvalidToken,
			Message:    "Authentication failed: invalid token",
			Suggestion: "Check your Personal Access Token in Zerops GUI",
		}
	}

	// 3. Discover project (must be exactly 1)
	projects, err := m.client.ListProjects(ctx, user.ID)
	if err != nil {
		return nil, &AuthError{
			Code:       platform.ErrAuthAPIError,
			Message:    "Failed to list projects",
			Suggestion: "Check network connectivity and token permissions",
		}
	}

	if len(projects) == 0 {
		return nil, &AuthError{
			Code:       platform.ErrTokenNoProject,
			Message:    "Token has no project access",
			Suggestion: "Check token permissions in Zerops GUI",
		}
	}

	if len(projects) > 1 {
		names := make([]string, len(projects))
		for i, p := range projects {
			names[i] = p.Name
		}
		return nil, &AuthError{
			Code:    platform.ErrTokenMultiProject,
			Message: fmt.Sprintf("Token has access to %d projects. ZAIA requires a single-project-scoped token.", len(projects)),
			Suggestion: fmt.Sprintf("Create a project-scoped token in Zerops GUI. Projects: %s",
				strings.Join(names, ", ")),
		}
	}

	project := projects[0]

	// 4. Store credentials
	data := Data{
		Token:   token,
		APIHost: apiHost,
		RegionData: RegionItem{
			Name:    region,
			Address: apiHost,
		},
		Project: ProjectInfo{
			ID:   project.ID,
			Name: project.Name,
		},
		User: UserData{
			Name:  user.FullName,
			Email: user.Email,
		},
	}
	if err := m.storage.Save(data); err != nil {
		return nil, fmt.Errorf("failed to store credentials: %w", err)
	}

	return &LoginResult{
		User: UserData{
			Name:  user.FullName,
			Email: user.Email,
		},
		Project: ProjectInfo{
			ID:   project.ID,
			Name: project.Name,
		},
		Region: region,
	}, nil
}

// GetCredentials returns the current authentication context from storage.
// Does NOT validate the token against the API (lazy validation).
func (m *Manager) GetCredentials() (*Credentials, error) {
	data, err := m.storage.Load()
	if err != nil {
		return nil, &AuthError{
			Code:       platform.ErrAuthRequired,
			Message:    "Not authenticated",
			Suggestion: "Run: zaia login <token>",
		}
	}

	if data.Token == "" {
		return nil, &AuthError{
			Code:       platform.ErrAuthRequired,
			Message:    "Not authenticated",
			Suggestion: "Run: zaia login <token>",
		}
	}

	if data.Project.ID == "" {
		return nil, &AuthError{
			Code:       platform.ErrAuthRequired,
			Message:    "Authenticated but no project discovered",
			Suggestion: "Run: zaia login <token>",
		}
	}

	return &Credentials{
		Token:       data.Token,
		APIHost:     data.APIHost,
		ProjectID:   data.Project.ID,
		ProjectName: data.Project.Name,
		Region:      data.RegionData.Name,
	}, nil
}

// GetStatus returns the full stored data for the status command.
func (m *Manager) GetStatus() (*Data, error) {
	data, err := m.storage.Load()
	if err != nil {
		return nil, &AuthError{
			Code:       platform.ErrAuthRequired,
			Message:    "Not authenticated",
			Suggestion: "Run: zaia login <token>",
		}
	}
	if data.Token == "" {
		return nil, &AuthError{
			Code:       platform.ErrAuthRequired,
			Message:    "Not authenticated",
			Suggestion: "Run: zaia login <token>",
		}
	}
	return &data, nil
}

// Logout clears stored credentials.
func (m *Manager) Logout() error {
	return m.storage.Clear()
}

// AuthError represents an authentication error.
type AuthError struct {
	Code       string `json:"code"`
	Message    string `json:"error"`
	Suggestion string `json:"suggestion"`
}

func (e *AuthError) Error() string {
	return e.Message
}
