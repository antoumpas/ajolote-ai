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

// SEC-001: Path traversal rejected at load time.
func TestLoadRejectsPathTraversal(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{"rule with ..", `{"mcp":{"servers":{}},"rules":["../../etc/passwd"],"skills":[],"personas":[],"context":[]}`},
		{"absolute rule", `{"mcp":{"servers":{}},"rules":["/etc/passwd"],"skills":[],"personas":[],"context":[]}`},
		{"scoped_rule path", `{"mcp":{"servers":{}},"rules":[],"skills":[],"personas":[],"context":[],"scoped_rules":[{"name":"x","globs":["*"],"path":"../../.env"}]}`},
		{"persona path", `{"mcp":{"servers":{}},"rules":[],"skills":[],"personas":["../../.ssh/id_rsa"],"context":[]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			os.MkdirAll(filepath.Join(dir, ".agents"), 0o755)
			os.WriteFile(filepath.Join(dir, ".agents", "config.json"), []byte(tc.json), 0o644)

			_, err := config.Load(dir)
			if err == nil {
				t.Fatal("expected error for path traversal, got nil")
			}
		})
	}
}

// SEC-003: Invalid MCP server names and env keys rejected at load time.
func TestLoadRejectsInvalidServerNames(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{"server name with ]", `{"mcp":{"servers":{"evil]":{"command":"x"}}},"rules":[],"skills":[],"personas":[],"context":[]}`},
		{"env key with newline", `{"mcp":{"servers":{"ok":{"command":"x","env":{"KEY\nINJECT":"v"}}}},"rules":[],"skills":[],"personas":[],"context":[]}`},
		{"env key with =", `{"mcp":{"servers":{"ok":{"command":"x","env":{"K=V":"v"}}}},"rules":[],"skills":[],"personas":[],"context":[]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			os.MkdirAll(filepath.Join(dir, ".agents"), 0o755)
			os.WriteFile(filepath.Join(dir, ".agents", "config.json"), []byte(tc.json), 0o644)

			_, err := config.Load(dir)
			if err == nil {
				t.Fatal("expected error for invalid name, got nil")
			}
		})
	}
}

// SEC-009: Oversized config rejected.
func TestLoadRejectsOversizedConfig(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".agents"), 0o755)
	// Write a file larger than 1 MB
	big := make([]byte, 1<<20+1)
	for i := range big {
		big[i] = ' '
	}
	os.WriteFile(filepath.Join(dir, ".agents", "config.json"), big, 0o644)

	_, err := config.Load(dir)
	if err == nil {
		t.Fatal("expected error for oversized config, got nil")
	}
}

// SEC-001: Valid paths accepted.
func TestLoadAcceptsValidPaths(t *testing.T) {
	cfgJSON := `{"mcp":{"servers":{}},"rules":[".agents/rules/general.md"],"skills":[".agents/skills/git.md"],"personas":[".agents/personas/reviewer.md"],"context":[".agents/context/arch.md"]}`
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".agents"), 0o755)
	os.WriteFile(filepath.Join(dir, ".agents", "config.json"), []byte(cfgJSON), 0o644)

	_, err := config.Load(dir)
	if err != nil {
		t.Fatalf("expected valid config to load, got: %v", err)
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
