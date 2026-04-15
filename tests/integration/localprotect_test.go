package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/cli/commands"
	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/localconfig"
)

// runUse executes `ajolote use <tool>` in dir.
func runUse(t *testing.T, dir, tool string) error {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		orig = os.TempDir()
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	cmd := commands.UseCmd()
	cmd.SetArgs([]string{tool})
	return cmd.Execute()
}

// writeLocalConfig writes .agents/config.local.json.
func writeLocalConfig(t *testing.T, dir string, lc *localconfig.LocalConfig) {
	t.Helper()
	data, err := json.MarshalIndent(lc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, localconfig.Filename, string(data))
}

// minimalConfig returns a bare-minimum config that lets `use claude` succeed.
func minimalConfig() *config.Config {
	return &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{}},
	}
}

// ── Core protection behaviour ─────────────────────────────────────────────────

func TestProtectFileNotOverwritten(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, minimalConfig())

	// First generation — creates CLAUDE.md
	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("first use failed: %v", err)
	}

	// Developer customises the file.
	customContent := "# My personal notes — do not overwrite\n"
	writeFile(t, dir, "CLAUDE.md", customContent)

	// Protect CLAUDE.md in the local config.
	writeLocalConfig(t, dir, &localconfig.LocalConfig{
		Protect: []string{"CLAUDE.md"},
	})

	// Second generation — CLAUDE.md must be left intact.
	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("second use failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}
	if string(data) != customContent {
		t.Errorf("CLAUDE.md was overwritten.\ngot:\n%s\nwant:\n%s", string(data), customContent)
	}
}

func TestUnprotectedFileIsOverwritten(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, minimalConfig())

	// Protect only a non-existent file — other files must still regenerate.
	writeLocalConfig(t, dir, &localconfig.LocalConfig{
		Protect: []string{"some-other-file.md"},
	})

	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("use failed: %v", err)
	}

	// CLAUDE.md must exist and contain ajolote-generated content.
	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md not created: %v", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		t.Error("CLAUDE.md is empty — generation did not run")
	}
}

func TestAbsentLocalConfigDefaultBehaviour(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, minimalConfig())
	// No .agents/config.local.json — normal behaviour, nothing protected.

	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("use failed without local config: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md should be generated when no local config exists")
	}
}

// ── Glob pattern ──────────────────────────────────────────────────────────────

func TestProtectGlobPattern(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		MCP:      config.MCP{Servers: map[string]config.MCPServer{}},
		Commands: []string{},
	}
	writeConfig(t, dir, cfg)

	// First generation — create Claude files including a command.
	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("first use failed: %v", err)
	}

	// Drop a custom command file.
	customCmd := "# My custom command\n\nDo something special.\n"
	writeFile(t, dir, ".claude/commands/my-cmd.md", customCmd)

	// Protect the entire commands directory via glob.
	writeLocalConfig(t, dir, &localconfig.LocalConfig{
		Protect: []string{".claude/commands/*.md"},
	})

	// Second generation — my-cmd.md must be left alone.
	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("second use failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude/commands/my-cmd.md"))
	if err != nil {
		t.Fatalf("reading my-cmd.md: %v", err)
	}
	if string(data) != customCmd {
		t.Errorf("my-cmd.md was overwritten.\ngot:\n%s\nwant:\n%s", string(data), customCmd)
	}
}

func TestProtectDirectoryPrefix(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, minimalConfig())

	// First generation.
	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("first use failed: %v", err)
	}

	// Customise a settings file inside .claude/.
	customSettings := `{"mcpServers":{"my-local-server":{"command":"my-tool"}}}`
	writeFile(t, dir, ".claude/settings.json", customSettings)

	// Protect everything under .claude/ via directory prefix.
	writeLocalConfig(t, dir, &localconfig.LocalConfig{
		Protect: []string{".claude/"},
	})

	// Second generation — .claude/settings.json must survive.
	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("second use failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude/settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}
	if string(data) != customSettings {
		t.Errorf(".claude/settings.json was overwritten.\ngot:\n%s\nwant:\n%s", string(data), customSettings)
	}
}

// ── Multiple files ────────────────────────────────────────────────────────────

func TestProtectMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, minimalConfig())

	// First generation.
	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("first use failed: %v", err)
	}

	// Customise two separate files.
	customCLAUDE := "# Personal CLAUDE notes\n"
	customSettings := `{"mcpServers":{},"personalNote":"mine"}`
	writeFile(t, dir, "CLAUDE.md", customCLAUDE)
	writeFile(t, dir, ".claude/settings.json", customSettings)

	// Protect both.
	writeLocalConfig(t, dir, &localconfig.LocalConfig{
		Protect: []string{"CLAUDE.md", ".claude/settings.json"},
	})

	// Second generation — both must survive.
	if err := runUse(t, dir, "claude"); err != nil {
		t.Fatalf("second use failed: %v", err)
	}

	claudeData, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(claudeData) != customCLAUDE {
		t.Errorf("CLAUDE.md was overwritten: got %q", string(claudeData))
	}

	settingsData, _ := os.ReadFile(filepath.Join(dir, ".claude/settings.json"))
	if string(settingsData) != customSettings {
		t.Errorf(".claude/settings.json was overwritten: got %q", string(settingsData))
	}
}

// ── Protection survives across tools ─────────────────────────────────────────

func TestProtectWorksForCursorToo(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, minimalConfig())

	// First generation — creates .cursor/mcp.json.
	if err := runUse(t, dir, "cursor"); err != nil {
		t.Fatalf("first use failed: %v", err)
	}

	// Developer customises the MCP config.
	customMCP := `{"mcpServers":{"my-dev-server":{"command":"my-tool"}}}`
	writeFile(t, dir, ".cursor/mcp.json", customMCP)

	// Protect it.
	writeLocalConfig(t, dir, &localconfig.LocalConfig{
		Protect: []string{".cursor/mcp.json"},
	})

	// Second generation — .cursor/mcp.json must not be touched.
	if err := runUse(t, dir, "cursor"); err != nil {
		t.Fatalf("second use failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".cursor/mcp.json"))
	if err != nil {
		t.Fatalf("reading .cursor/mcp.json: %v", err)
	}
	if string(data) != customMCP {
		t.Errorf(".cursor/mcp.json was overwritten: got %q", string(data))
	}
}
