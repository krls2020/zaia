package commands

import (
	"github.com/zeropsio/zaia/internal/auth"
	"github.com/zeropsio/zaia/internal/output"
	"github.com/zeropsio/zaia/internal/platform"
)

// resolveCredentials loads credentials from storage and returns an error response if not authenticated.
func resolveCredentials(storagePath string) (*auth.Credentials, error) {
	storage := auth.NewStorage(storagePath)
	mgr := auth.NewManager(storage, nil)
	creds, err := mgr.GetCredentials()
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			return nil, output.Err(authErr.Code, authErr.Message, authErr.Suggestion, nil)
		}
		return nil, output.Err(platform.ErrAuthRequired, "Not authenticated", "Run: zaia login <token>", nil)
	}
	return creds, nil
}

// findServiceByHostname finds a service by hostname in a list of services.
func findServiceByHostname(services []platform.ServiceStack, hostname string) *platform.ServiceStack {
	for i := range services {
		if services[i].Name == hostname {
			return &services[i]
		}
	}
	return nil
}

// listHostnames returns a comma-separated list of service hostnames.
func listHostnames(services []platform.ServiceStack) string {
	if len(services) == 0 {
		return "(none)"
	}
	result := ""
	for i, s := range services {
		if i > 0 {
			result += ", "
		}
		result += s.Name
	}
	return result
}

// resolveServiceID resolves a hostname to a service ID.
// Returns the service or outputs a SERVICE_NOT_FOUND error.
func resolveServiceID(client platform.Client, projectID, hostname string, services []platform.ServiceStack) (*platform.ServiceStack, error) {
	svc := findServiceByHostname(services, hostname)
	if svc == nil {
		return nil, output.Err(platform.ErrServiceNotFound,
			"Service '"+hostname+"' not found",
			"Available services: "+listHostnames(services),
			map[string]interface{}{
				"projectId":        projectID,
				"requestedService": hostname,
			})
	}
	return svc, nil
}
