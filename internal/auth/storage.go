package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Data is ZAIA's persistent storage format (zaia.data).
type Data struct {
	Token      string      `json:"token"`
	APIHost    string      `json:"apiHost"`
	RegionData RegionItem  `json:"regionData"`
	Project    ProjectInfo `json:"project"`
	User       UserData    `json:"user"`
}

// RegionItem contains server/region information.
type RegionItem struct {
	Name      string  `json:"name"`
	IsDefault bool    `json:"isDefault"`
	Address   string  `json:"address"`
	GUIAddr   *string `json:"guiAddress,omitempty"`
}

// ProjectInfo contains active project context.
type ProjectInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UserData contains user details stored at login time.
type UserData struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Storage handles reading and writing zaia.data.
type Storage struct {
	filePath string
}

// NewStorage creates a new storage handler.
// If filePath is empty, it uses the default location.
func NewStorage(filePath string) *Storage {
	if filePath == "" {
		filePath = defaultFilePath()
	}
	return &Storage{filePath: filePath}
}

// Load reads the stored data. Returns empty Data if file doesn't exist.
func (s *Storage) Load() (Data, error) {
	content, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return Data{}, nil
		}
		return Data{}, fmt.Errorf("failed to read %s: %w", s.filePath, err)
	}

	var data Data
	if err := json.Unmarshal(content, &data); err != nil {
		return Data{}, fmt.Errorf("failed to parse %s: %w", s.filePath, err)
	}
	return data, nil
}

// Save writes data atomically (write to .new, then rename).
func (s *Storage) Save(data Data) error {
	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	newPath := s.filePath + ".new"
	f, err := os.OpenFile(newPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if err := json.NewEncoder(f).Encode(data); err != nil {
		f.Close()
		os.Remove(newPath)
		return fmt.Errorf("failed to write data: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(newPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(newPath, s.filePath); err != nil {
		os.Remove(newPath)
		return fmt.Errorf("failed to rename: %w", err)
	}
	return nil
}

// Clear removes the data file.
func (s *Storage) Clear() error {
	err := os.Remove(s.filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", s.filePath, err)
	}
	return nil
}

// FilePath returns the storage file path.
func (s *Storage) FilePath() string {
	return s.filePath
}

func defaultFilePath() string {
	// Check env override
	if p := os.Getenv("ZAIA_DATA_FILE_PATH"); p != "" {
		return p
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "zerops.zaia.data")
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "zerops", "zaia.data")
	default: // linux
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "zerops", "zaia.data")
		}
		return filepath.Join(home, ".config", "zerops", "zaia.data")
	}
}
