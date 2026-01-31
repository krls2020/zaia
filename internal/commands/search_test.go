package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zeropsio/zaia/internal/output"
)

func TestSearchCmd_ReturnsResults(t *testing.T) {
	cmd := NewSearch()
	cmd.SetArgs([]string{"postgresql", "connection", "string"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	if data["query"] != "postgresql connection string" {
		t.Errorf("query = %v, want 'postgresql connection string'", data["query"])
	}
	results := data["results"].([]interface{})
	if len(results) == 0 {
		t.Error("expected non-empty results for 'postgresql connection string'")
	}
	if data["topResult"] == nil {
		t.Error("expected topResult for high-scoring query")
	}
}

func TestSearchCmd_NoArgs_Error(t *testing.T) {
	cmd := NewSearch()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no query")
	}
}

func TestSearch_Get_Found(t *testing.T) {
	cmd := NewSearch()
	cmd.SetArgs([]string{"--get", "zerops://docs/services/postgresql"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "sync" {
		t.Errorf("type = %v, want sync", resp["type"])
	}
	data := resp["data"].(map[string]interface{})
	if data["uri"] != "zerops://docs/services/postgresql" {
		t.Errorf("uri = %v, want zerops://docs/services/postgresql", data["uri"])
	}
	if data["title"] == nil || data["title"] == "" {
		t.Error("expected non-empty title")
	}
	if data["content"] == nil || data["content"] == "" {
		t.Error("expected non-empty content")
	}
}

func TestSearch_Get_NotFound(t *testing.T) {
	cmd := NewSearch()
	cmd.SetArgs([]string{"--get", "zerops://docs/nonexistent"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent URI")
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	if resp["type"] != "error" {
		t.Errorf("type = %v, want error", resp["type"])
	}
	if resp["code"] != "NOT_FOUND" {
		t.Errorf("code = %v, want NOT_FOUND", resp["code"])
	}
}

func TestSearchCmd_UnsupportedService(t *testing.T) {
	cmd := NewSearch()
	cmd.SetArgs([]string{"mongodb"})

	var stdout bytes.Buffer
	output.SetWriter(&stdout)
	defer output.ResetWriter()

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(stdout.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	suggestions := data["suggestions"].([]interface{})
	if len(suggestions) == 0 {
		t.Error("expected suggestions for unsupported 'mongodb'")
	}
}
