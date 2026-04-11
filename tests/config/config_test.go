package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/config"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.MCP.Servers["test"] = config.MCPServer{Command: "npx", Args: []string{"test"}}

	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, config.ConfigPath)); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, ok := loaded.MCP.Servers["test"]; !ok {
		t.Error("expected MCP server 'test' to survive save/load round-trip")
	}
	if len(loaded.Rules) == 0 {
		t.Error("expected Rules to be non-empty after round-trip")
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := config.Load(dir)
	if err == nil {
		t.Fatal("expected error loading from empty dir")
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if config.Exists(dir) {
		t.Fatal("should not exist yet")
	}

	if err := config.Save(dir, config.DefaultConfig()); err != nil {
		t.Fatal(err)
	}

	if !config.Exists(dir) {
		t.Fatal("should exist after save")
	}
}

func TestPersonaBackwardCompat(t *testing.T) {
	// Old format: personas is a JSON array of plain strings
	oldJSON := `{
		"mcp": {"servers": {}},
		"rules": [],
		"skills": [],
		"personas": [".agents/personas/reviewer.md", ".agents/personas/architect.md"],
		"context": []
	}`

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".agents")
	if err := os.MkdirAll(cfgPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgPath, "config.json"), []byte(oldJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load of old-format config failed: %v", err)
	}
	if len(loaded.Personas) != 2 {
		t.Fatalf("expected 2 personas, got %d", len(loaded.Personas))
	}
	if loaded.Personas[0].Path != ".agents/personas/reviewer.md" {
		t.Errorf("expected first persona path '.agents/personas/reviewer.md', got %q", loaded.Personas[0].Path)
	}
	if loaded.Personas[0].Claude != nil {
		t.Error("old-format persona should have nil Claude block")
	}
}

func TestPersonaRoundTrip(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.Personas = []config.Persona{
		{Path: ".agents/personas/reviewer.md"}, // simple form
		{
			Path: ".agents/personas/architect.md",
			Claude: &config.ClaudeAgent{
				Model:       "haiku",
				Tools:       []string{"Read", "Grep"},
				Description: "Software architect for design decisions.",
			},
		},
	}

	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify the JSON uses the object form for personas with claude blocks
	data, _ := os.ReadFile(filepath.Join(dir, config.ConfigPath))
	var raw struct {
		Personas []json.RawMessage `json:"personas"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.Personas) != 2 {
		t.Fatalf("expected 2 personas in JSON, got %d", len(raw.Personas))
	}

	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Personas) != 2 {
		t.Fatalf("expected 2 personas after round-trip, got %d", len(loaded.Personas))
	}
	arch := loaded.Personas[1]
	if arch.Path != ".agents/personas/architect.md" {
		t.Errorf("wrong path: %q", arch.Path)
	}
	if arch.Claude == nil {
		t.Fatal("claude block should survive round-trip")
	}
	if arch.Claude.Model != "haiku" {
		t.Errorf("expected model 'haiku', got %q", arch.Claude.Model)
	}
	if arch.Claude.Description != "Software architect for design decisions." {
		t.Errorf("wrong description: %q", arch.Claude.Description)
	}
}
