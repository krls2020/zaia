package platform

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Error codes for ZAIA CLI.
const (
	ErrAuthRequired           = "AUTH_REQUIRED"
	ErrAuthInvalidToken       = "AUTH_INVALID_TOKEN"
	ErrAuthTokenExpired       = "AUTH_TOKEN_EXPIRED"
	ErrAuthAPIError           = "AUTH_API_ERROR"
	ErrTokenNoProject         = "TOKEN_NO_PROJECT"
	ErrTokenMultiProject      = "TOKEN_MULTI_PROJECT"
	ErrServiceNotFound        = "SERVICE_NOT_FOUND"
	ErrServiceRequired        = "SERVICE_REQUIRED"
	ErrConfirmRequired        = "CONFIRM_REQUIRED"
	ErrFileNotFound           = "FILE_NOT_FOUND"
	ErrZeropsYmlNotFound      = "ZEROPS_YML_NOT_FOUND"
	ErrInvalidZeropsYml       = "INVALID_ZEROPS_YML"
	ErrInvalidImportYml       = "INVALID_IMPORT_YML"
	ErrImportHasProject       = "IMPORT_HAS_PROJECT"
	ErrInvalidScaling         = "INVALID_SCALING"
	ErrInvalidParameter       = "INVALID_PARAMETER"
	ErrInvalidEnvFormat       = "INVALID_ENV_FORMAT"
	ErrInvalidHostname        = "INVALID_HOSTNAME"
	ErrUnknownType            = "UNKNOWN_TYPE"
	ErrProcessNotFound        = "PROCESS_NOT_FOUND"
	ErrProcessAlreadyTerminal = "PROCESS_ALREADY_TERMINAL"
	ErrPermissionDenied       = "PERMISSION_DENIED"
	ErrAPIError               = "API_ERROR"
	ErrAPITimeout             = "API_TIMEOUT"
	ErrAPIRateLimited         = "API_RATE_LIMITED"
	ErrNetworkError           = "NETWORK_ERROR"
	ErrInvalidUsage           = "INVALID_USAGE"
	ErrSetupDownloadFailed    = "SETUP_DOWNLOAD_FAILED"
	ErrSetupInstallFailed     = "SETUP_INSTALL_FAILED"
	ErrSetupConfigFailed      = "SETUP_CONFIG_FAILED"
	ErrSetupUnsupportedOS     = "SETUP_UNSUPPORTED_OS"
)

// MapHTTPError maps an HTTP status code and error to a ZAIA error code.
func MapHTTPError(statusCode int, apiErr error) (code string, suggestion string) {
	switch statusCode {
	case 401:
		return ErrAuthTokenExpired, "Token is no longer valid. Create a new Personal Access Token and run: zaia login <new-token>"
	case 403:
		return ErrPermissionDenied, "Insufficient permissions for this operation"
	case 404:
		return ErrServiceNotFound, ""
	case 429:
		return ErrAPIRateLimited, "Rate limited. Try again later"
	default:
		if statusCode >= 500 {
			return ErrAPIError, "Zerops API error. Try again later"
		}
		return ErrAPIError, ""
	}
}

// MapNetworkError determines if an error is a network error and returns the appropriate code.
func MapNetworkError(err error) (code string, isNetwork bool) {
	if err == nil {
		return "", false
	}

	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return ErrNetworkError, true
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return ErrNetworkError, true
	}

	msg := err.Error()
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "network is unreachable") ||
		strings.Contains(msg, "i/o timeout") {
		return ErrNetworkError, true
	}

	if strings.Contains(msg, "context deadline exceeded") {
		return ErrAPITimeout, true
	}

	return "", false
}

// ExitCodeForError returns the exit code for a given error code.
func ExitCodeForError(code string) int {
	switch code {
	case ErrAuthRequired, ErrAuthInvalidToken, ErrAuthTokenExpired, ErrAuthAPIError,
		ErrTokenNoProject, ErrTokenMultiProject:
		return 2
	case ErrServiceRequired, ErrConfirmRequired, ErrFileNotFound, ErrZeropsYmlNotFound,
		ErrInvalidZeropsYml, ErrInvalidImportYml, ErrImportHasProject,
		ErrInvalidScaling, ErrInvalidParameter, ErrInvalidEnvFormat,
		ErrInvalidHostname, ErrUnknownType, ErrInvalidUsage:
		return 3
	case ErrServiceNotFound, ErrProcessNotFound, ErrProcessAlreadyTerminal:
		return 4
	case ErrPermissionDenied:
		return 5
	case ErrNetworkError:
		return 6
	case ErrSetupDownloadFailed, ErrSetupInstallFailed, ErrSetupConfigFailed, ErrSetupUnsupportedOS:
		return 7
	default:
		return 1
	}
}

// HTTPError represents an HTTP error from the Zerops API.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

// PlatformError carries a ZAIA error code, message, and suggestion.
// Used by ZeropsClient to map SDK errors to ZAIA error codes.
type PlatformError struct {
	Code       string
	Message    string
	Suggestion string
}

func (e *PlatformError) Error() string {
	return e.Message
}

// NewPlatformError creates a PlatformError with the given code, message, and suggestion.
func NewPlatformError(code, message, suggestion string) *PlatformError {
	return &PlatformError{
		Code:       code,
		Message:    message,
		Suggestion: suggestion,
	}
}
