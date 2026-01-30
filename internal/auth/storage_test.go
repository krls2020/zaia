package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorage_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zaia.data")
	s := NewStorage(path)

	data := Data{
		Token:   "test-token",
		APIHost: "api.zerops.io",
		Project: ProjectInfo{ID: "p1", Name: "my-app"},
		User:    UserData{Name: "John", Email: "john@test.com"},
	}

	if err := s.Save(data); err != nil {
		t.Fatal(err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Token != "test-token" {
		t.Errorf("token = %q, want test-token", loaded.Token)
	}
	if loaded.Project.ID != "p1" {
		t.Errorf("project.id = %q, want p1", loaded.Project.ID)
	}
	if loaded.User.Name != "John" {
		t.Errorf("user.name = %q, want John", loaded.User.Name)
	}
}

func TestStorage_LoadNonExistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.data")
	s := NewStorage(path)

	data, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if data.Token != "" {
		t.Errorf("token = %q, want empty", data.Token)
	}
}

func TestStorage_Clear(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zaia.data")
	s := NewStorage(path)

	// Save data
	if err := s.Save(Data{Token: "t1"}); err != nil {
		t.Fatal(err)
	}

	// Clear
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}

	// Load should return empty
	data, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if data.Token != "" {
		t.Errorf("token = %q after clear, want empty", data.Token)
	}
}

func TestStorage_ClearNonExistent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.data")
	s := NewStorage(path)

	// Should not error
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
}

func TestStorage_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zaia.data")
	s := NewStorage(path)

	if err := s.Save(Data{Token: "t1"}); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestStorage_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zaia.data")
	s := NewStorage(path)

	// Save initial
	if err := s.Save(Data{Token: "v1"}); err != nil {
		t.Fatal(err)
	}

	// Save new version
	if err := s.Save(Data{Token: "v2"}); err != nil {
		t.Fatal(err)
	}

	// No .new file should remain
	if _, err := os.Stat(path + ".new"); !os.IsNotExist(err) {
		t.Error(".new file should not exist after successful save")
	}

	// Verify latest data
	data, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if data.Token != "v2" {
		t.Errorf("token = %q, want v2", data.Token)
	}
}

func TestStorage_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "deep", "zaia.data")
	s := NewStorage(path)

	if err := s.Save(Data{Token: "t1"}); err != nil {
		t.Fatal(err)
	}

	data, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if data.Token != "t1" {
		t.Errorf("token = %q, want t1", data.Token)
	}
}
