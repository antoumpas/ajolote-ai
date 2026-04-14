package config_test

import (
	"path/filepath"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/config"
)

// ── rebaseConfigPaths ────────────────────────────────────────────────────────

func TestRebaseConfigPathsRules(t *testing.T) {
	cfg := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/rules/general.md", ".agents/rules/style.md"},
	}
	rebased := config.RebaseConfigPaths(cfg)
	for _, p := range rebased.Rules {
		if filepath.ToSlash(p)[:len(".agents/.base/")] != ".agents/.base/" {
			t.Errorf("expected rebased prefix, got %q", p)
		}
	}
}

func TestRebaseConfigPathsPersonas(t *testing.T) {
	cfg := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{}},
		Personas: []config.Persona{
			{Path: ".agents/personas/reviewer.md"},
		},
	}
	rebased := config.RebaseConfigPaths(cfg)
	if rebased.Personas[0].Path != ".agents/.base/personas/reviewer.md" {
		t.Errorf("persona path not rebased: %q", rebased.Personas[0].Path)
	}
}

func TestRebaseConfigPathsScopedRules(t *testing.T) {
	cfg := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{}},
		ScopedRules: []config.ScopedRule{
			{Name: "frontend", Globs: []string{"*.tsx"}, Path: ".agents/rules/frontend.md"},
		},
	}
	rebased := config.RebaseConfigPaths(cfg)
	if rebased.ScopedRules[0].Path != ".agents/.base/rules/frontend.md" {
		t.Errorf("scoped rule path not rebased: %q", rebased.ScopedRules[0].Path)
	}
}

func TestRebaseConfigPathsCommands(t *testing.T) {
	cfg := &config.Config{
		MCP:      config.MCP{Servers: map[string]config.MCPServer{}},
		Commands: []string{".agents/commands/review.md"},
	}
	rebased := config.RebaseConfigPaths(cfg)
	if rebased.Commands[0] != ".agents/.base/commands/review.md" {
		t.Errorf("command path not rebased: %q", rebased.Commands[0])
	}
}

// ── mergeConfigs ─────────────────────────────────────────────────────────────

func TestMergeLocalWinsOnMCPConflict(t *testing.T) {
	base := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{
			"github": {Command: "base-github", Scope: "project"},
		}},
	}
	local := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{
			"github": {Command: "local-github", Scope: "user"},
		}},
	}
	merged := config.MergeConfigs(base, local)
	if merged.MCP.Servers["github"].Command != "local-github" {
		t.Errorf("local MCP server should win, got %q", merged.MCP.Servers["github"].Command)
	}
}

func TestMergeBaseAddsNewMCPServers(t *testing.T) {
	base := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{
			"org-server": {Command: "npx", Args: []string{"org-mcp"}},
		}},
	}
	local := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{
			"local-server": {Command: "npx", Args: []string{"local-mcp"}},
		}},
	}
	merged := config.MergeConfigs(base, local)
	if _, ok := merged.MCP.Servers["org-server"]; !ok {
		t.Error("base MCP server should be present in merged config")
	}
	if _, ok := merged.MCP.Servers["local-server"]; !ok {
		t.Error("local MCP server should be present in merged config")
	}
}

func TestMergeRulesLocalFirst(t *testing.T) {
	base := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/.base/rules/org.md"},
	}
	local := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/rules/general.md"},
	}
	merged := config.MergeConfigs(base, local)
	if len(merged.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(merged.Rules))
	}
	if merged.Rules[0] != ".agents/rules/general.md" {
		t.Errorf("local rule should be first, got %q", merged.Rules[0])
	}
}

func TestMergeRulesDeduplicatesByFilename(t *testing.T) {
	// Base has "general.md" — local also has "general.md" → only local copy survives.
	base := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/.base/rules/general.md", ".agents/.base/rules/org.md"},
	}
	local := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/rules/general.md"},
	}
	merged := config.MergeConfigs(base, local)
	// Should have local general.md + base org.md (no duplicate general)
	if len(merged.Rules) != 2 {
		t.Fatalf("expected 2 rules (no duplicate), got %d: %v", len(merged.Rules), merged.Rules)
	}
	foundLocal := false
	for _, r := range merged.Rules {
		if r == ".agents/rules/general.md" {
			foundLocal = true
		}
		if r == ".agents/.base/rules/general.md" {
			t.Error("base general.md should be suppressed by local general.md")
		}
	}
	if !foundLocal {
		t.Error("local general.md should be in merged rules")
	}
}

func TestMergeScopedRulesLocalWins(t *testing.T) {
	base := &config.Config{
		MCP:         config.MCP{Servers: map[string]config.MCPServer{}},
		ScopedRules: []config.ScopedRule{{Name: "frontend", Globs: []string{"*.tsx"}, Path: ".agents/.base/rules/frontend.md"}},
	}
	local := &config.Config{
		MCP:         config.MCP{Servers: map[string]config.MCPServer{}},
		ScopedRules: []config.ScopedRule{{Name: "frontend", Globs: []string{"*.jsx", "*.tsx"}, Path: ".agents/rules/frontend.md"}},
	}
	merged := config.MergeConfigs(base, local)
	if len(merged.ScopedRules) != 1 {
		t.Fatalf("expected 1 scoped rule, got %d", len(merged.ScopedRules))
	}
	if merged.ScopedRules[0].Path != ".agents/rules/frontend.md" {
		t.Errorf("local scoped rule should win, got %q", merged.ScopedRules[0].Path)
	}
}

func TestMergePersonasDeduplicatedByFilename(t *testing.T) {
	base := &config.Config{
		MCP:      config.MCP{Servers: map[string]config.MCPServer{}},
		Personas: []config.Persona{{Path: ".agents/.base/personas/reviewer.md"}, {Path: ".agents/.base/personas/architect.md"}},
	}
	local := &config.Config{
		MCP:      config.MCP{Servers: map[string]config.MCPServer{}},
		Personas: []config.Persona{{Path: ".agents/personas/reviewer.md"}},
	}
	merged := config.MergeConfigs(base, local)
	// Should have local reviewer + base architect (no duplicate reviewer)
	if len(merged.Personas) != 2 {
		t.Fatalf("expected 2 personas, got %d", len(merged.Personas))
	}
	for _, p := range merged.Personas {
		if p.Path == ".agents/.base/personas/reviewer.md" {
			t.Error("base reviewer should be suppressed by local reviewer")
		}
	}
}

func TestMergeExtendsNotInherited(t *testing.T) {
	base := &config.Config{
		MCP:     config.MCP{Servers: map[string]config.MCPServer{}},
		Extends: "https://example.com/grandparent",
	}
	local := &config.Config{
		MCP:     config.MCP{Servers: map[string]config.MCPServer{}},
		Extends: "https://example.com/parent",
	}
	merged := config.MergeConfigs(base, local)
	if merged.Extends != "" {
		t.Errorf("merged config should have Extends cleared (no chaining), got %q", merged.Extends)
	}
}

func TestMergeEmptyBase(t *testing.T) {
	base := &config.Config{MCP: config.MCP{Servers: map[string]config.MCPServer{}}}
	local := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{"s": {Command: "x"}}},
		Rules: []string{".agents/rules/general.md"},
	}
	merged := config.MergeConfigs(base, local)
	if len(merged.Rules) != 1 || merged.Rules[0] != ".agents/rules/general.md" {
		t.Errorf("unexpected rules: %v", merged.Rules)
	}
	if _, ok := merged.MCP.Servers["s"]; !ok {
		t.Error("local MCP server missing from merged")
	}
}
