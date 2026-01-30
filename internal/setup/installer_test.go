package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	p := DetectPlatform()
	if p.OS == "" {
		t.Error("OS should not be empty")
	}
	if p.Arch == "" {
		t.Error("Arch should not be empty")
	}
	if p.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", p.OS, runtime.GOOS)
	}
}

func TestDownloadURL(t *testing.T) {
	tests := []struct {
		name    string
		os      string
		arch    string
		want    string
		wantErr bool
	}{
		{
			name: "darwin_amd64",
			os:   "darwin",
			arch: "amd64",
			want: "https://github.com/krls2020/zaia-mcp/releases/latest/download/zaia-mcp-darwin-amd64",
		},
		{
			name: "darwin_arm64",
			os:   "darwin",
			arch: "arm64",
			want: "https://github.com/krls2020/zaia-mcp/releases/latest/download/zaia-mcp-darwin-arm64",
		},
		{
			name: "linux_amd64",
			os:   "linux",
			arch: "amd64",
			want: "https://github.com/krls2020/zaia-mcp/releases/latest/download/zaia-mcp-linux-amd64",
		},
		{
			name: "linux_arm64",
			os:   "linux",
			arch: "arm64",
			want: "https://github.com/krls2020/zaia-mcp/releases/latest/download/zaia-mcp-linux-arm64",
		},
		{
			name:    "unsupported_windows",
			os:      "windows",
			arch:    "amd64",
			wantErr: true,
		},
		{
			name:    "unsupported_freebsd",
			os:      "freebsd",
			arch:    "amd64",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Platform{OS: tt.os, Arch: tt.arch}
			got, err := DownloadURL(p)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultBinaryPath(t *testing.T) {
	p := DefaultBinaryPath()
	if !strings.HasSuffix(p, filepath.Join(".local", "bin", "zaia-mcp")) {
		t.Errorf("path %q should end with .local/bin/zaia-mcp", p)
	}
}

// mockDownloader implements Downloader for testing.
type mockDownloader struct {
	data []byte
	err  error
}

func (m *mockDownloader) Download(url string) ([]byte, error) {
	return m.data, m.err
}

func TestInstallBinary_Success(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "subdir", "zaia-mcp")

	downloader := &mockDownloader{data: []byte("#!/bin/sh\necho hello")}

	err := InstallBinary(downloader, "https://example.com/zaia-mcp", binPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check binary exists
	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("binary not found: %v", err)
	}
	// Check executable permission
	if info.Mode()&0111 == 0 {
		t.Error("binary should be executable")
	}
	// Check content
	data, _ := os.ReadFile(binPath)
	if string(data) != "#!/bin/sh\necho hello" {
		t.Errorf("content = %q, want %q", string(data), "#!/bin/sh\necho hello")
	}
}

func TestInstallBinary_DownloadFails(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "zaia-mcp")

	downloader := &mockDownloader{err: fmt.Errorf("network error")}

	err := InstallBinary(downloader, "https://example.com/zaia-mcp", binPath)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("error = %q, should contain 'network error'", err.Error())
	}
}

func TestInstallBinary_Idempotent(t *testing.T) {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "zaia-mcp")

	downloader := &mockDownloader{data: []byte("v1")}

	// First install
	err := InstallBinary(downloader, "https://example.com/zaia-mcp", binPath)
	if err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	// Second install overwrites
	downloader.data = []byte("v2")
	err = InstallBinary(downloader, "https://example.com/zaia-mcp", binPath)
	if err != nil {
		t.Fatalf("second install failed: %v", err)
	}

	data, _ := os.ReadFile(binPath)
	if string(data) != "v2" {
		t.Errorf("content = %q, want %q", string(data), "v2")
	}
}

// httpDownloader test (just ensure interface compliance)
func TestHTTPDownloader_Implements(t *testing.T) {
	var _ Downloader = &HTTPDownloader{}
	_ = &HTTPDownloader{} // compile check
}
