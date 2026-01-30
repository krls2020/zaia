package setup

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	releaseBaseURL = "https://github.com/krls2020/zaia-mcp/releases/latest/download"
	binaryName     = "zaia-mcp"
)

// Platform holds OS and architecture info.
type Platform struct {
	OS   string
	Arch string
}

// DetectPlatform returns the current OS and architecture.
func DetectPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// DownloadURL builds the GitHub release download URL for the given platform.
func DownloadURL(p Platform) (string, error) {
	switch p.OS {
	case "darwin", "linux":
		// supported
	default:
		return "", fmt.Errorf("unsupported OS: %s (only darwin and linux are supported)", p.OS)
	}
	return fmt.Sprintf("%s/%s-%s-%s", releaseBaseURL, binaryName, p.OS, p.Arch), nil
}

// DefaultBinaryPath returns ~/.local/bin/zaia-mcp.
func DefaultBinaryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "bin", binaryName)
}

// Downloader downloads a URL and returns the bytes.
type Downloader interface {
	Download(url string) ([]byte, error)
}

// HTTPDownloader downloads via HTTP.
type HTTPDownloader struct{}

func (h *HTTPDownloader) Download(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec // URL is constructed internally
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return data, nil
}

// InstallBinary downloads from url and writes the binary to binPath with executable permissions.
// Creates parent directories if needed.
func InstallBinary(dl Downloader, url, binPath string) error {
	data, err := dl.Download(url)
	if err != nil {
		return err
	}

	dir := filepath.Dir(binPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(binPath, data, 0755); err != nil {
		return fmt.Errorf("writing binary: %w", err)
	}

	return nil
}
