package platform

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/zeropsio/zerops-go/apiError"
)

func TestNewZeropsClient(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		apiHost string
	}{
		{"bare host", "tok", "api.zerops.io"},
		{"https host", "tok", "https://api.zerops.io"},
		{"with trailing slash", "tok", "https://api.zerops.io/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewZeropsClient(tt.token, tt.apiHost)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c == nil {
				t.Fatal("client is nil")
			}
		})
	}
}

func TestMapSDKError_Nil(t *testing.T) {
	err := mapSDKError(nil, "service")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestMapSDKError_NetworkErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{
			"net.OpError",
			&net.OpError{Op: "dial", Err: errors.New("connection refused")},
			ErrNetworkError,
		},
		{
			"DNS error",
			&net.DNSError{Err: "no such host", Name: "api.zerops.io"},
			ErrNetworkError,
		},
		{
			"context deadline",
			context.DeadlineExceeded,
			ErrAPITimeout,
		},
		{
			"context canceled",
			context.Canceled,
			ErrAPIError,
		},
		{
			"connection refused string",
			errors.New("connection refused"),
			ErrNetworkError,
		},
		{
			"unknown error",
			errors.New("something unexpected"),
			ErrAPIError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapSDKError(tt.err, "service")
			if result == nil {
				t.Fatal("expected error, got nil")
			}
			var pe *PlatformError
			if !errors.As(result, &pe) {
				t.Fatalf("expected PlatformError, got %T", result)
			}
			if pe.Code != tt.wantCode {
				t.Errorf("code = %s, want %s", pe.Code, tt.wantCode)
			}
		})
	}
}

func TestMapAPIError_HTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
		entity     string
		wantCode   string
	}{
		{"401 unauthorized", 401, "", "service", ErrAuthTokenExpired},
		{"403 forbidden", 403, "", "service", ErrPermissionDenied},
		{"404 service", 404, "", "service", ErrServiceNotFound},
		{"404 process", 404, "", "process", ErrProcessNotFound},
		{"429 rate limited", 429, "", "service", ErrAPIRateLimited},
		{"500 server error", 500, "", "service", ErrAPIError},
		{"503 unavailable", 503, "", "service", ErrAPIError},
		{"subdomain already enabled", 200, "SubdomainAccessAlreadyEnabled", "service", "SUBDOMAIN_ALREADY_ENABLED"},
		{"subdomain already disabled", 200, "serviceStackSubdomainAccessAlreadyDisabled", "service", "SUBDOMAIN_ALREADY_DISABLED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := apiError.Error{
				HttpStatusCode: tt.statusCode,
				ErrorCode:      tt.errorCode,
				Message:        "test error",
			}
			result := mapAPIError(apiErr, tt.entity)
			var pe *PlatformError
			if !errors.As(result, &pe) {
				t.Fatalf("expected PlatformError, got %T", result)
			}
			if pe.Code != tt.wantCode {
				t.Errorf("code = %s, want %s", pe.Code, tt.wantCode)
			}
		})
	}
}

func TestMapSDKError_APIError(t *testing.T) {
	apiErr := apiError.Error{
		HttpStatusCode: 401,
		ErrorCode:      "unauthorized",
		Message:        "token expired",
	}
	result := mapSDKError(apiErr, "service")
	var pe *PlatformError
	if !errors.As(result, &pe) {
		t.Fatalf("expected PlatformError, got %T", result)
	}
	if pe.Code != ErrAuthTokenExpired {
		t.Errorf("code = %s, want %s", pe.Code, ErrAuthTokenExpired)
	}
}
