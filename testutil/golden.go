package testutil

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// AssertGolden compares got against a golden file.
// If -update flag is set, it regenerates the golden file.
func AssertGolden(t *testing.T, name string, got []byte) {
	t.Helper()

	goldenDir := filepath.Join("testdata", "golden")
	goldenPath := filepath.Join(goldenDir, name)

	if *update {
		if err := os.MkdirAll(goldenDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, got, 0644); err != nil {
			t.Fatal(err)
		}
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden file %s not found. Run with -update to create it.\nGot: %s", goldenPath, got)
	}

	if string(want) != string(got) {
		t.Errorf("golden mismatch for %s:\nwant: %s\ngot:  %s", name, want, got)
	}
}

// AssertJSONEqual compares two JSON values for structural equality.
func AssertJSONEqual(t *testing.T, want, got string) {
	t.Helper()

	var wantVal, gotVal interface{}
	if err := json.Unmarshal([]byte(want), &wantVal); err != nil {
		t.Fatalf("invalid want JSON: %v\n%s", err, want)
	}
	if err := json.Unmarshal([]byte(got), &gotVal); err != nil {
		t.Fatalf("invalid got JSON: %v\n%s", err, got)
	}

	wantNorm, _ := json.Marshal(wantVal)
	gotNorm, _ := json.Marshal(gotVal)

	if string(wantNorm) != string(gotNorm) {
		wantPretty, _ := json.MarshalIndent(wantVal, "", "  ")
		gotPretty, _ := json.MarshalIndent(gotVal, "", "  ")
		t.Errorf("JSON mismatch:\nwant: %s\ngot:  %s", wantPretty, gotPretty)
	}
}
