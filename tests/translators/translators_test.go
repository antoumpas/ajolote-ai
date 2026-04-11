package translators_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/translators"
)

func testConfig() *config.Config {
	return config.DefaultConfig()
}

// seedCommand writes a command file into .agents/commands/ under dir.
func seedCommand(t *testing.T, dir, name, content string) {
	t.Helper()
	cmdDir := filepath.Join(dir, ".agents", "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, name+".md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAllTranslatorsGenerate(t *testing.T) {
	for _, tr := range translators.All() {
		tr := tr
		t.Run(tr.Name(), func(t *testing.T) {
			dir := t.TempDir()
			if err := tr.Generate(testConfig(), dir); err != nil {
				t.Fatalf("%s Generate: %v", tr.Name(), err)
			}
			for _, f := range tr.OutputFiles() {
				full := filepath.Join(dir, f)
				if strings.HasSuffix(f, "/") {
					// Directory pattern — verify at least one file was written inside
					entries, err := os.ReadDir(full)
					if err != nil {
						t.Errorf("expected output directory %s: %v", f, err)
						continue
					}
					hasFile := false
					for _, e := range entries {
						if !e.IsDir() {
							hasFile = true
							break
						}
					}
					if !hasFile {
						t.Errorf("output directory %s is empty", f)
					}
				} else {
					data, err := os.ReadFile(full)
					if err != nil {
						t.Errorf("expected output file %s: %v", f, err)
						continue
					}
					if len(data) == 0 {
						t.Errorf("output file %s is empty", f)
					}
				}
			}
		})
	}
}

func TestClaudeTranslatorContent(t *testing.T) {
	dir := t.TempDir()
	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	content := string(data)

	if !strings.Contains(content, ".agents/rules/general.md") {
		t.Error("CLAUDE.md should reference rules files")
	}
	if !strings.Contains(content, ".agents/skills/git.md") {
		t.Error("CLAUDE.md should reference skill files")
	}
}

func TestCursorMCPContent(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	cfg.MCP.Servers["filesystem"] = config.MCPServer{
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
	}

	tr := &translators.CursorTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".cursor/mcp.json"))
	if !strings.Contains(string(data), "filesystem") {
		t.Error(".cursor/mcp.json should contain filesystem server")
	}
}

func TestRegistryGet(t *testing.T) {
	for _, name := range []string{"claude", "cursor", "windsurf", "copilot", "cline", "aider"} {
		if _, err := translators.Get(name); err != nil {
			t.Errorf("Get(%q) failed: %v", name, err)
		}
	}
	if _, err := translators.Get("unknown"); err == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestNames(t *testing.T) {
	names := translators.Names()
	for _, expected := range []string{"claude", "cursor", "windsurf"} {
		if !strings.Contains(names, expected) {
			t.Errorf("Names() missing %q, got: %s", expected, names)
		}
	}
}

// --- Commands translation tests ---

func TestCommandsGeneratedForClaude(t *testing.T) {
	dir := t.TempDir()
	seedCommand(t, dir, "deploy", "---\ndescription: Deploy to staging\n---\n\nRun deploy.sh\n")

	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude/commands/deploy.md"))
	if err != nil {
		t.Fatalf(".claude/commands/deploy.md not generated: %v", err)
	}
	if !strings.Contains(string(data), "Run deploy.sh") {
		t.Error(".claude/commands/deploy.md missing command content")
	}
}

func TestCommandsGeneratedForCursor(t *testing.T) {
	dir := t.TempDir()
	seedCommand(t, dir, "deploy", "---\ndescription: Deploy to staging\n---\n\nRun deploy.sh\n")

	tr := &translators.CursorTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".cursor/rules/deploy.mdc"))
	if err != nil {
		t.Fatalf(".cursor/rules/deploy.mdc not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Deploy to staging") {
		t.Error(".cursor/rules/deploy.mdc missing description")
	}
	if !strings.Contains(content, "Run deploy.sh") {
		t.Error(".cursor/rules/deploy.mdc missing command content")
	}
	if !strings.Contains(content, "alwaysApply: false") {
		t.Error(".cursor/rules/deploy.mdc missing alwaysApply frontmatter")
	}
}

func TestCommandsGeneratedForCline(t *testing.T) {
	dir := t.TempDir()
	seedCommand(t, dir, "deploy", "---\ndescription: Deploy to staging\n---\n\nRun deploy.sh\n")

	tr := &translators.ClineTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".roo/rules/deploy.md"))
	if err != nil {
		t.Fatalf(".roo/rules/deploy.md not generated: %v", err)
	}
	if !strings.Contains(string(data), "Run deploy.sh") {
		t.Error(".roo/rules/deploy.md missing command content")
	}
}

func TestCommandsGeneratedForWindsurf(t *testing.T) {
	dir := t.TempDir()
	seedCommand(t, dir, "deploy", "---\ndescription: Deploy to staging\n---\n\nRun deploy.sh\n")

	tr := &translators.WindsurfTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".windsurf/workflows/deploy.yaml"))
	if err != nil {
		t.Fatalf(".windsurf/workflows/deploy.yaml not generated: %v", err)
	}
	if !strings.Contains(string(data), "Run deploy.sh") {
		t.Error(".windsurf/workflows/deploy.yaml missing command content")
	}
}

// --- Import / two-way sync tests ---

func TestClaudeImportMCPServers(t *testing.T) {
	dir := t.TempDir()
	settingsJSON := `{"mcpServers":{"github":{"command":"npx","args":["-y","@modelcontextprotocol/server-github"]}}}`
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, ".claude/settings.json"), []byte(settingsJSON), 0o644)

	tr := &translators.ClaudeTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil ImportResult")
	}
	if _, ok := result.NewMCPServers["github"]; !ok {
		t.Error("expected github MCP server in ImportResult")
	}
}

func TestClaudeImportCommands(t *testing.T) {
	dir := t.TempDir()

	// Simulate .claude/settings.json existing (required for Import to return non-nil)
	os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude/settings.json"), []byte(`{"mcpServers":{}}`), 0o644)

	// A user-created command
	os.WriteFile(filepath.Join(dir, ".claude/commands/deploy.md"),
		[]byte("---\ndescription: Deploy\n---\n\nRun deploy.sh\n"), 0o644)
	// ajolote-managed file — must be skipped
	os.WriteFile(filepath.Join(dir, ".claude/commands/ajolote-sync.md"), []byte("sync content"), 0o644)

	tr := &translators.ClaudeTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil ImportResult")
	}
	if len(result.NewCommands) != 1 {
		t.Fatalf("expected 1 new command, got %d", len(result.NewCommands))
	}
	if result.NewCommands[0].Name != "deploy" {
		t.Errorf("expected command name 'deploy', got %q", result.NewCommands[0].Name)
	}
	if result.NewCommands[0].Description != "Deploy" {
		t.Errorf("expected description 'Deploy', got %q", result.NewCommands[0].Description)
	}
}

func TestClaudeImportSkipsExistingCommands(t *testing.T) {
	dir := t.TempDir()

	// deploy.md already in .agents/commands/
	seedCommand(t, dir, "deploy", "existing content")

	os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude/settings.json"), []byte(`{"mcpServers":{}}`), 0o644)
	os.WriteFile(filepath.Join(dir, ".claude/commands/deploy.md"), []byte("tool version"), 0o644)

	tr := &translators.ClaudeTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NewCommands) != 0 {
		t.Errorf("expected no new commands (deploy already in .agents/commands/), got %d", len(result.NewCommands))
	}
}

func TestCursorImportCommands(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0o755)
	os.WriteFile(filepath.Join(dir, ".cursor/mcp.json"), []byte(`{"mcpServers":{}}`), 0o644)

	// User-created MDC command
	os.WriteFile(filepath.Join(dir, ".cursor/rules/deploy.mdc"),
		[]byte("---\ndescription: Deploy\nalwaysApply: false\n---\n\nRun deploy.sh\n"), 0o644)
	// ajolote-managed files — must be skipped
	os.WriteFile(filepath.Join(dir, ".cursor/rules/agents.mdc"), []byte("rules"), 0o644)
	os.WriteFile(filepath.Join(dir, ".cursor/rules/ajolote-sync.mdc"), []byte("sync"), 0o644)

	tr := &translators.CursorTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NewCommands) != 1 {
		t.Fatalf("expected 1 new command, got %d", len(result.NewCommands))
	}
	if result.NewCommands[0].Name != "deploy" {
		t.Errorf("expected 'deploy', got %q", result.NewCommands[0].Name)
	}
	if !strings.Contains(result.NewCommands[0].Content, "Run deploy.sh") {
		t.Error("cursor command import should strip MDC frontmatter and preserve body")
	}
}

func TestClineImportCommands(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".roo", "rules"), 0o755)
	os.WriteFile(filepath.Join(dir, ".roo/mcp.json"), []byte(`{"mcpServers":{}}`), 0o644)

	// Cline command format: "# name\n\ncontent"
	os.WriteFile(filepath.Join(dir, ".roo/rules/deploy.md"),
		[]byte("# deploy\n\nRun deploy.sh\n"), 0o644)
	os.WriteFile(filepath.Join(dir, ".roo/rules/ajolote-sync.md"), []byte("sync"), 0o644)

	tr := &translators.ClineTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NewCommands) != 1 {
		t.Fatalf("expected 1 new command, got %d", len(result.NewCommands))
	}
	if result.NewCommands[0].Name != "deploy" {
		t.Errorf("expected 'deploy', got %q", result.NewCommands[0].Name)
	}
	if !strings.Contains(result.NewCommands[0].Content, "Run deploy.sh") {
		t.Error("cline command import should strip '# name' header and preserve body")
	}
}

func TestClaudeImportCommandsWithoutSettingsJSON(t *testing.T) {
	// Regression: Import() used to return nil if settings.json was absent,
	// skipping commands even when .claude/commands/ had content.
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755)
	// No settings.json — project uses Claude but has no MCP servers configured

	os.WriteFile(filepath.Join(dir, ".claude/commands/deploy.md"),
		[]byte("---\ndescription: Deploy\n---\n\nRun deploy.sh\n"), 0o644)

	tr := &translators.ClaudeTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil ImportResult when .claude/ exists but settings.json is absent")
	}
	if len(result.NewCommands) != 1 {
		t.Fatalf("expected 1 imported command, got %d", len(result.NewCommands))
	}
	if result.NewCommands[0].Name != "deploy" {
		t.Errorf("expected command 'deploy', got %q", result.NewCommands[0].Name)
	}
}

func TestCursorImportCommandsWithoutMCPJson(t *testing.T) {
	// Regression: Import() used to return nil if .cursor/mcp.json was absent,
	// skipping commands even when .cursor/rules/ had content.
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0o755)
	// No mcp.json — project uses Cursor but has no MCP servers configured

	os.WriteFile(filepath.Join(dir, ".cursor/rules/deploy.mdc"),
		[]byte("---\ndescription: Deploy\nalwaysApply: false\n---\n\nRun deploy.sh\n"), 0o644)

	tr := &translators.CursorTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil ImportResult when .cursor/ exists but mcp.json is absent")
	}
	if len(result.NewCommands) != 1 {
		t.Fatalf("expected 1 imported command, got %d", len(result.NewCommands))
	}
	if result.NewCommands[0].Name != "deploy" {
		t.Errorf("expected command 'deploy', got %q", result.NewCommands[0].Name)
	}
	if !strings.Contains(result.NewCommands[0].Content, "Run deploy.sh") {
		t.Error("expected command body to contain 'Run deploy.sh'")
	}
}

func TestClineImportCommandsWithoutMCPJson(t *testing.T) {
	// Regression: Import() used to return nil if .roo/mcp.json was absent,
	// skipping commands even when .roo/rules/ had content.
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".roo", "rules"), 0o755)
	// No mcp.json — project uses Cline but has no MCP servers configured

	os.WriteFile(filepath.Join(dir, ".roo/rules/deploy.md"),
		[]byte("# deploy\n\nRun deploy.sh\n"), 0o644)

	tr := &translators.ClineTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil ImportResult when .roo/ exists but mcp.json is absent")
	}
	if len(result.NewCommands) != 1 {
		t.Fatalf("expected 1 imported command, got %d", len(result.NewCommands))
	}
	if result.NewCommands[0].Name != "deploy" {
		t.Errorf("expected command 'deploy', got %q", result.NewCommands[0].Name)
	}
	if !strings.Contains(result.NewCommands[0].Content, "Run deploy.sh") {
		t.Error("expected command body to contain 'Run deploy.sh'")
	}
}

func TestImportNoFilesReturnsNil(t *testing.T) {
	// Tools with no files on disk should return nil (not empty result)
	for _, tr := range []translators.Syncer{
		&translators.ClaudeTranslator{},
		&translators.CursorTranslator{},
		&translators.ClineTranslator{},
	} {
		dir := t.TempDir()
		result, err := tr.Import(dir)
		if err != nil {
			t.Errorf("%s Import with no files: unexpected error: %v", tr.Name(), err)
		}
		if result != nil {
			t.Errorf("%s Import with no files: expected nil result, got %+v", tr.Name(), result)
		}
	}
}

func TestImportIdempotent(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude/settings.json"), []byte(`{"mcpServers":{}}`), 0o644)
	os.WriteFile(filepath.Join(dir, ".claude/commands/deploy.md"), []byte("Run deploy.sh\n"), 0o644)

	tr := &translators.ClaudeTranslator{}

	// First import — deploy is new
	result1, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result1.NewCommands) != 1 {
		t.Fatalf("first import: expected 1 new command, got %d", len(result1.NewCommands))
	}

	// Simulate writing it to .agents/commands/
	seedCommand(t, dir, "deploy", "Run deploy.sh\n")

	// Second import — deploy is already in .agents/commands/, should be skipped
	result2, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result2.NewCommands) != 0 {
		t.Errorf("second import: expected 0 new commands (idempotent), got %d", len(result2.NewCommands))
	}
}
