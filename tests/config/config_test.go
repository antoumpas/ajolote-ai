package config_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// ── Config inheritance (Resolve) ─────────────────────────────────────────────

// TestResolveWithNoExtends verifies that Resolve is a no-op when Extends is empty.
func TestResolveWithNoExtends(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	if err := config.Save(dir, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := config.Resolve(loaded, dir, false)
	if err != nil {
		t.Fatalf("Resolve with no extends: %v", err)
	}
	if resolved != loaded {
		t.Error("Resolve with no extends should return the same pointer")
	}
}

// TestResolveWithLocalFileExtends verifies that a local file base is merged correctly.
func TestResolveWithLocalFileExtends(t *testing.T) {
	// Create a base project with an org-wide rule.
	baseProject := t.TempDir()
	os.MkdirAll(filepath.Join(baseProject, ".agents", "rules"), 0o755)
	os.MkdirAll(filepath.Join(baseProject, ".agents", "mcp"), 0o755)

	baseCfgJSON := `{
		"mcp": {"servers": {"org-mcp": {"command": "npx", "args": ["org-tool"]}}},
		"rules": [".agents/rules/org-style.md"],
		"skills": [],
		"personas": [],
		"context": []
	}`
	os.WriteFile(filepath.Join(baseProject, ".agents", "config.json"), []byte(baseCfgJSON), 0o644)
	os.WriteFile(filepath.Join(baseProject, ".agents", "rules", "org-style.md"), []byte("# Org style"), 0o644)

	// Create the inheriting project.
	project := t.TempDir()
	os.MkdirAll(filepath.Join(project, ".agents"), 0o755)
	os.WriteFile(filepath.Join(project, ".agents", "config.json"), []byte(`{
		"extends": "`+filepath.ToSlash(baseProject)+`",
		"mcp": {"servers": {}},
		"rules": [],
		"skills": [],
		"personas": [],
		"context": []
	}`), 0o644)

	cfg, err := config.Load(project)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := config.Resolve(cfg, project, true)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	// Inherited MCP server should be present.
	if _, ok := resolved.MCP.Servers["org-mcp"]; !ok {
		t.Error("inherited org-mcp server should be in resolved config")
	}

	// Inherited rule should be rebased to .agents/.base/
	found := false
	for _, r := range resolved.Rules {
		if filepath.Base(r) == "org-style.md" {
			if r[:len(".agents/.base/")] != ".agents/.base/" {
				t.Errorf("inherited rule should be under .agents/.base/, got %q", r)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("inherited rule org-style.md not found in resolved rules: %v", resolved.Rules)
	}
}

// TestResolveExtendsLocalWins verifies that a local config overrides the base.
func TestResolveExtendsLocalWins(t *testing.T) {
	// Base has an MCP server and a rule.
	baseProject := t.TempDir()
	os.MkdirAll(filepath.Join(baseProject, ".agents", "rules"), 0o755)
	baseCfgJSON := `{
		"mcp": {"servers": {"shared": {"command": "base-cmd"}}},
		"rules": [".agents/rules/general.md"],
		"skills": [],
		"personas": [],
		"context": []
	}`
	os.WriteFile(filepath.Join(baseProject, ".agents", "config.json"), []byte(baseCfgJSON), 0o644)
	os.WriteFile(filepath.Join(baseProject, ".agents", "rules", "general.md"), []byte("# Base general"), 0o644)

	// Local project overrides the same MCP server and the same rule filename.
	project := t.TempDir()
	os.MkdirAll(filepath.Join(project, ".agents", "rules"), 0o755)
	os.WriteFile(filepath.Join(project, ".agents", "config.json"), []byte(`{
		"extends": "`+filepath.ToSlash(baseProject)+`",
		"mcp": {"servers": {"shared": {"command": "local-cmd"}}},
		"rules": [".agents/rules/general.md"],
		"skills": [],
		"personas": [],
		"context": []
	}`), 0o644)
	os.WriteFile(filepath.Join(project, ".agents", "rules", "general.md"), []byte("# Local general"), 0o644)

	cfg, err := config.Load(project)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := config.Resolve(cfg, project, true)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	// Local MCP server wins.
	if resolved.MCP.Servers["shared"].Command != "local-cmd" {
		t.Errorf("local MCP server should win, got %q", resolved.MCP.Servers["shared"].Command)
	}

	// No duplicate general.md — local wins, base copy suppressed.
	count := 0
	for _, r := range resolved.Rules {
		if filepath.Base(r) == "general.md" {
			count++
			if r != ".agents/rules/general.md" {
				t.Errorf("general.md should come from local, got %q", r)
			}
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 general.md in rules, got %d: %v", count, resolved.Rules)
	}
}

// TestResolveHTTPExtends verifies that an HTTPS base source is fetched and merged.
func TestResolveHTTPExtends(t *testing.T) {
	files := map[string]string{
		"/.agents/config.json":  `{"mcp":{"servers":{"http-server":{"command":"npx"}}},"rules":[".agents/rules/http.md"],"skills":[],"personas":[],"context":[]}`,
		"/.agents/rules/http.md": "# HTTP Rule",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if content, ok := files[r.URL.Path]; ok {
			w.Write([]byte(content))
		} else {
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	project := t.TempDir()
	os.MkdirAll(filepath.Join(project, ".agents"), 0o755)
	os.WriteFile(filepath.Join(project, ".agents", "config.json"), []byte(`{
		"extends": "`+srv.URL+`",
		"mcp": {"servers": {}},
		"rules": [],
		"skills": [],
		"personas": [],
		"context": []
	}`), 0o644)


	cfg, _ := config.Load(project)
	resolved, err := config.Resolve(cfg, project, true)
	if err != nil {
		t.Fatalf("Resolve (HTTP): %v", err)
	}

	if _, ok := resolved.MCP.Servers["http-server"]; !ok {
		t.Error("http-server MCP should be inherited")
	}

	found := false
	for _, r := range resolved.Rules {
		if filepath.Base(r) == "http.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("http.md rule not found in resolved rules: %v", resolved.Rules)
	}
}

// TestResolvePreservesExtendsInRawConfig verifies Save/Load round-trip keeps extends.
func TestResolvePreservesExtendsInRawConfig(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".agents"), 0o755)
	os.WriteFile(filepath.Join(dir, ".agents", "config.json"), []byte(`{
		"extends": "https://example.com/standards",
		"mcp": {"servers": {}},
		"rules": [],
		"skills": [],
		"personas": [],
		"context": []
	}`), 0o644)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Extends != "https://example.com/standards" {
		t.Errorf("Extends should be preserved by Load, got %q", cfg.Extends)
	}

	// Save and reload — extends must survive round-trip.
	if err := config.Save(dir, cfg); err != nil {
		t.Fatal(err)
	}
	reloaded, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Extends != "https://example.com/standards" {
		t.Errorf("Extends should survive Save/Load round-trip, got %q", reloaded.Extends)
	}
}
