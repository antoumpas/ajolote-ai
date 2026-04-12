package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/cli/commands"
	"github.com/ajolote-ai/ajolote/internal/config"
)

// runValidate executes the validate command in dir and returns any error.
func runValidate(t *testing.T, dir string) error {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		orig = os.TempDir() // fallback: previous test may have left a deleted cwd
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	cmd := commands.ValidateCmd()
	cmd.SetArgs(nil)
	return cmd.Execute()
}

// writeConfig serialises cfg into .agents/config.json under dir.
func writeConfig(t *testing.T, dir string, cfg *config.Config) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".agents", "config.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeFile creates a file at relPath (relative to dir) with the given content.
func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAllFilesPresent(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{
			".agents/rules/general.md",
		},
		Skills: []string{
			".agents/skills/git.md",
		},
		Personas: []config.Persona{
			{Path: ".agents/personas/reviewer.md"},
		},
		Context: []string{
			".agents/context/arch.md",
		},
	}
	writeConfig(t, dir, cfg)
	writeFile(t, dir, ".agents/rules/general.md", "# General\n\nAlways read before writing.\n")
	writeFile(t, dir, ".agents/skills/git.md", "# Git\n\nUse feature branches.\n")
	writeFile(t, dir, ".agents/personas/reviewer.md", "# Reviewer\n\nReview code carefully.\n")
	writeFile(t, dir, ".agents/context/arch.md", "# Architecture\n\nMonolith for now.\n")

	if err := runValidate(t, dir); err != nil {
		t.Errorf("validate should pass when all files are present, got: %v", err)
	}
}

func TestValidateMissingFile(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/rules/general.md"}, // file not created
	}
	writeConfig(t, dir, cfg)

	if err := runValidate(t, dir); err == nil {
		t.Error("validate should fail when a referenced file is missing")
	}
}

func TestValidateEmptyFile(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/rules/general.md"},
	}
	writeConfig(t, dir, cfg)
	writeFile(t, dir, ".agents/rules/general.md", "") // empty

	if err := runValidate(t, dir); err == nil {
		t.Error("validate should fail when a referenced file is empty")
	}
}

func TestValidateWhitespaceOnlyFile(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/rules/general.md"},
	}
	writeConfig(t, dir, cfg)
	writeFile(t, dir, ".agents/rules/general.md", "   \n\n  \t\n") // whitespace only

	if err := runValidate(t, dir); err == nil {
		t.Error("validate should fail when a referenced file contains only whitespace")
	}
}

func TestValidateScopedRuleNoGlobs(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{}},
		ScopedRules: []config.ScopedRule{
			{Name: "frontend", Globs: nil, Path: ".agents/rules/frontend.md"},
		},
	}
	writeConfig(t, dir, cfg)
	writeFile(t, dir, ".agents/rules/frontend.md", "# Frontend\n\nUse TypeScript.\n")

	if err := runValidate(t, dir); err == nil {
		t.Error("validate should fail when a scoped rule has no glob patterns")
	}
}

func TestValidateScopedRuleWithGlobs(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{}},
		ScopedRules: []config.ScopedRule{
			{Name: "frontend", Globs: []string{"**/*.tsx"}, Path: ".agents/rules/frontend.md"},
		},
	}
	writeConfig(t, dir, cfg)
	writeFile(t, dir, ".agents/rules/frontend.md", "# Frontend\n\nUse TypeScript.\n")

	if err := runValidate(t, dir); err != nil {
		t.Errorf("validate should pass for a valid scoped rule, got: %v", err)
	}
}

func TestValidateMCPHTTPMissingURL(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP: config.MCP{
			Servers: map[string]config.MCPServer{
				"remote": {Transport: "http"}, // no URL
			},
		},
	}
	writeConfig(t, dir, cfg)

	if err := runValidate(t, dir); err == nil {
		t.Error("validate should fail when an http server has no url")
	}
}

func TestValidateMCPHTTPWithURL(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP: config.MCP{
			Servers: map[string]config.MCPServer{
				"remote": {Transport: "http", URL: "https://mcp.example.com/api"},
			},
		},
	}
	writeConfig(t, dir, cfg)

	if err := runValidate(t, dir); err != nil {
		t.Errorf("validate should pass for a valid http server, got: %v", err)
	}
}

func TestValidateMCPStdioNoCommand(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP: config.MCP{
			Servers: map[string]config.MCPServer{
				"broken": {Transport: "stdio"}, // no command
			},
		},
	}
	writeConfig(t, dir, cfg)

	if err := runValidate(t, dir); err == nil {
		t.Error("validate should fail when a stdio server has no command")
	}
}

func TestValidateMCPStdioCommandInPath(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP: config.MCP{
			Servers: map[string]config.MCPServer{
				// "sh" is guaranteed to be in PATH on any Unix system
				"shell-server": {Command: "sh", Args: []string{"-c", "echo hi"}},
			},
		},
	}
	writeConfig(t, dir, cfg)

	if err := runValidate(t, dir); err != nil {
		t.Errorf("validate should pass when command is in PATH, got: %v", err)
	}
}

func TestValidateEmptyConfig(t *testing.T) {
	dir := t.TempDir()

	cfg := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{}},
	}
	writeConfig(t, dir, cfg)

	// Empty config with no files referenced should pass (nothing to validate)
	if err := runValidate(t, dir); err != nil {
		t.Errorf("validate should pass for a config with no referenced files, got: %v", err)
	}
}
