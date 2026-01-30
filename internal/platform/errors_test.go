package platform

import (
	"fmt"
	"net"
	"testing"
)

func TestMapHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantCode   string
	}{
		{"401_expired", 401, ErrAuthTokenExpired},
		{"403_denied", 403, ErrPermissionDenied},
		{"404_not_found", 404, ErrServiceNotFound},
		{"429_rate_limited", 429, ErrAPIRateLimited},
		{"500_server_error", 500, ErrAPIError},
		{"503_unavailable", 503, ErrAPIError},
		{"400_client_error", 400, ErrAPIError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := MapHTTPError(tt.statusCode, nil)
			if code != tt.wantCode {
				t.Errorf("MapHTTPError(%d) = %q, want %q", tt.statusCode, code, tt.wantCode)
			}
		})
	}
}

func TestMapNetworkError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  string
		wantIsNet bool
	}{
		{"nil_error", nil, "", false},
		{"regular_error", fmt.Errorf("something"), "", false},
		{"connection_refused", fmt.Errorf("connection refused"), ErrNetworkError, true},
		{"no_such_host", fmt.Errorf("no such host"), ErrNetworkError, true},
		{"deadline_exceeded", fmt.Errorf("context deadline exceeded"), ErrAPITimeout, true},
		{"net_op_error", &net.OpError{Op: "dial", Err: fmt.Errorf("test")}, ErrNetworkError, true},
		{"dns_error", &net.DNSError{Name: "example.com"}, ErrNetworkError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, isNet := MapNetworkError(tt.err)
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if isNet != tt.wantIsNet {
				t.Errorf("isNetwork = %v, want %v", isNet, tt.wantIsNet)
			}
		})
	}
}

func TestExitCodeForError(t *testing.T) {
	tests := []struct {
		code     string
		wantExit int
	}{
		{ErrAuthRequired, 2},
		{ErrAuthInvalidToken, 2},
		{ErrAuthTokenExpired, 2},
		{ErrTokenNoProject, 2},
		{ErrTokenMultiProject, 2},
		{ErrServiceRequired, 3},
		{ErrConfirmRequired, 3},
		{ErrFileNotFound, 3},
		{ErrInvalidZeropsYml, 3},
		{ErrInvalidImportYml, 3},
		{ErrImportHasProject, 3},
		{ErrInvalidScaling, 3},
		{ErrInvalidParameter, 3},
		{ErrInvalidEnvFormat, 3},
		{ErrInvalidHostname, 3},
		{ErrUnknownType, 3},
		{ErrInvalidUsage, 3},
		{ErrServiceNotFound, 4},
		{ErrProcessNotFound, 4},
		{ErrProcessAlreadyTerminal, 4},
		{ErrPermissionDenied, 5},
		{ErrNetworkError, 6},
		{ErrAPIError, 1},
		{ErrAPITimeout, 1},
		{ErrAPIRateLimited, 1},
		{"UNKNOWN", 1},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := ExitCodeForError(tt.code); got != tt.wantExit {
				t.Errorf("ExitCodeForError(%q) = %d, want %d", tt.code, got, tt.wantExit)
			}
		})
	}
}
