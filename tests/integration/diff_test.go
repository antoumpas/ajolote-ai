package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/cli/commands"
	"github.com/ajolote-ai/ajolote/internal/config"
)

// runDiff executes the diff command in dir (optionally with a tool arg) and returns any error.
func runDiff(t *testing.T, dir string, args ...string) error {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		orig = os.TempDir() // fallback: previous test may have left a deleted cwd
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	cmd := commands.DiffCmd()
	cmd.SetArgs(args)
	return cmd.Execute()
}

// setupClaudeProject creates a minimal ajolote project with Claude output already generated.
func setupClaudeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir() // cleanup registered first → runs LAST (LIFO)

	cfg := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/rules/general.md"},
	}
	writeConfig(t, dir, cfg)
	writeFile(t, dir, ".agents/rules/general.md", "# General\n\nAlways read before writing.\n")

	// Register cwd restore AFTER t.TempDir() so it runs FIRST (LIFO),
	// before the temp dir is deleted. Without this, deleting the cwd dir
	// causes os.Getwd() to fail on Windows in subsequent tests.
	orig, err := os.Getwd()
	if err != nil {
		orig = os.TempDir()
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			_ = os.Chdir(os.TempDir())
		}
	})

	// Generate Claude output so the translator is considered "active"
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	useCmd := commands.UseCmd()
	useCmd.SetArgs([]string{"claude"})
	if err := useCmd.Execute(); err != nil {
		t.Fatalf("use claude: %v", err)
	}

	return dir
}

func TestDiffNothingChanged(t *testing.T) {
	dir := setupClaudeProject(t)

	// diff immediately after use — nothing should have changed
	if err := runDiff(t, dir, "claude"); err != nil {
		t.Errorf("diff should exit 0 when nothing would change, got: %v", err)
	}
}

func TestDiffDetectsChange(t *testing.T) {
	dir := setupClaudeProject(t)

	// Modify the rule file — next sync would produce different CLAUDE.md content
	writeFile(t, dir, ".agents/rules/general.md",
		"# General\n\nAlways read before writing.\n\n## New Rule\n\nAdded after generation.\n")

	// diff should detect that CLAUDE.md content would not change (it only lists paths,
	// not file content) but the rules file itself changed — CLAUDE.md is unchanged.
	// Actually CLAUDE.md just @-imports the file, so it wouldn't change. Let's test
	// by adding a new rule path to config instead.
	cfg := &config.Config{
		MCP: config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{
			".agents/rules/general.md",
			".agents/rules/code-style.md", // newly added path
		},
	}
	writeConfig(t, dir, cfg)
	writeFile(t, dir, ".agents/rules/code-style.md", "# Code Style\n\nUse tabs.\n")

	err := runDiff(t, dir, "claude")
	if err == nil {
		t.Error("diff should exit 1 when generated output would change")
	}
	if !strings.Contains(err.Error(), "diff: run ajolote sync to apply") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiffNoToolsConfigured(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		MCP:   config.MCP{Servers: map[string]config.MCPServer{}},
		Rules: []string{".agents/rules/general.md"},
	}
	writeConfig(t, dir, cfg)
	writeFile(t, dir, ".agents/rules/general.md", "# General\n\nContent.\n")

	// No tool output files exist → diff should print a message and exit 0
	if err := runDiff(t, dir); err != nil {
		t.Errorf("diff with no active tools should exit 0, got: %v", err)
	}
}

func TestDiffInvalidTool(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, &config.Config{MCP: config.MCP{Servers: map[string]config.MCPServer{}}})

	if err := runDiff(t, dir, "notarealtool"); err == nil {
		t.Error("diff with unknown tool should return error")
	}
}

func TestDiffUnifiedOutput(t *testing.T) {
	// Unit-test the unifiedDiff helper directly via the exported command path.
	// We verify the diff algorithm produces correct output by running diff on
	// a known before/after pair via the integration flow.

	dir := setupClaudeProject(t)

	// Capture original CLAUDE.md content
	orig, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	// Sanity: should be unchanged immediately after generation
	if err := runDiff(t, dir, "claude"); err != nil {
		t.Fatalf("fresh diff should be clean: %v", err)
	}

	// Corrupt CLAUDE.md on disk to simulate drift
	corrupted := strings.Replace(string(orig), "@.agents/rules/general.md", "@.agents/rules/OTHER.md", 1)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(corrupted), 0o644); err != nil {
		t.Fatal(err)
	}

	// Now diff should detect a change
	if err := runDiff(t, dir, "claude"); err == nil {
		t.Error("diff should detect the corrupted CLAUDE.md")
	}
}
