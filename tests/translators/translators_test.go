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

// scopedOnlyDirs are output directories only created under specific conditions.
// They are absent from a minimal test config and must not fail TestAllTranslatorsGenerate.
var scopedOnlyDirs = map[string]bool{
	".claude/rules/":        true,
	".github/instructions/": true,
	".claude/agents/":       true, // only written when a persona has a claude: block
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
					if scopedOnlyDirs[f] {
						continue // only created when scoped_rules are present
					}
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

	// Claude Code uses @file import syntax
	if !strings.Contains(content, "@.agents/rules/general.md") {
		t.Error("CLAUDE.md should use @file import syntax for rules")
	}
	if !strings.Contains(content, "@.agents/skills/git.md") {
		t.Error("CLAUDE.md should use @file import syntax for skills")
	}
	// Must NOT contain bullet-list references
	if strings.Contains(content, "- `.agents/") {
		t.Error("CLAUDE.md should not use bullet-list file references")
	}
}

func TestInlineToolsEmbedContent(t *testing.T) {
	dir := t.TempDir()

	// Create the referenced files so inlineFiles can read them
	os.MkdirAll(filepath.Join(dir, ".agents", "rules"), 0o755)
	os.MkdirAll(filepath.Join(dir, ".agents", "skills"), 0o755)
	os.WriteFile(filepath.Join(dir, ".agents/rules/general.md"), []byte("# General Rules\n\nAlways test your code."), 0o644)
	os.WriteFile(filepath.Join(dir, ".agents/rules/code-style.md"), []byte("# Code Style\n\nMatch the surrounding style."), 0o644)
	os.WriteFile(filepath.Join(dir, ".agents/skills/git.md"), []byte("# Git\n\nUse feature branches."), 0o644)
	os.WriteFile(filepath.Join(dir, ".agents/skills/testing.md"), []byte("# Testing\n\nWrite tests first."), 0o644)

	cfg := testConfig()

	for _, tc := range []struct {
		name    string
		tr      translators.Syncer
		outFile string
	}{
		{"copilot", &translators.CopilotTranslator{}, ".github/copilot-instructions.md"},
		{"windsurf", &translators.WindsurfTranslator{}, ".windsurf/rules/agents.md"},
		{"cline", &translators.ClineTranslator{}, ".clinerules"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.tr.Generate(cfg, dir); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(dir, tc.outFile))
			if err != nil {
				t.Fatalf("output file %s not found: %v", tc.outFile, err)
			}
			content := string(data)
			if !strings.Contains(content, "Always test your code.") {
				t.Errorf("%s: expected rule content to be inlined, got:\n%s", tc.name, content)
			}
			if !strings.Contains(content, "Use feature branches.") {
				t.Errorf("%s: expected skill content to be inlined, got:\n%s", tc.name, content)
			}
			// Must NOT contain raw bullet-list path references
			if strings.Contains(content, "- `.agents/") {
				t.Errorf("%s: should not contain bullet-list file references", tc.name)
			}
		})
	}
}

// --- Scoped rules tests ---

func testScopedConfig(t *testing.T, dir string) *config.Config {
	t.Helper()
	os.MkdirAll(filepath.Join(dir, ".agents", "rules"), 0o755)
	os.WriteFile(filepath.Join(dir, ".agents/rules/frontend.md"), []byte("# Frontend Rules\n\nUse React hooks."), 0o644)

	cfg := testConfig()
	cfg.ScopedRules = []config.ScopedRule{
		{Name: "frontend", Globs: []string{"**/*.tsx", "**/*.css"}, Path: ".agents/rules/frontend.md"},
	}
	return cfg
}

func TestCursorScopedRules(t *testing.T) {
	dir := t.TempDir()
	cfg := testScopedConfig(t, dir)

	tr := &translators.CursorTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".cursor/rules/frontend.mdc"))
	if err != nil {
		t.Fatalf(".cursor/rules/frontend.mdc not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "globs: **/*.tsx, **/*.css") {
		t.Error(".cursor/rules/frontend.mdc should contain globs frontmatter")
	}
	if !strings.Contains(content, "alwaysApply: false") {
		t.Error(".cursor/rules/frontend.mdc should have alwaysApply: false")
	}
	if !strings.Contains(content, "Use React hooks.") {
		t.Error(".cursor/rules/frontend.mdc should contain inlined rule content")
	}
}

func TestCopilotScopedRules(t *testing.T) {
	dir := t.TempDir()
	cfg := testScopedConfig(t, dir)

	tr := &translators.CopilotTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".github/instructions/frontend.instructions.md"))
	if err != nil {
		t.Fatalf(".github/instructions/frontend.instructions.md not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "applyTo:") {
		t.Error("copilot instructions file should contain applyTo: frontmatter")
	}
	if !strings.Contains(content, "**/*.tsx") {
		t.Error("copilot instructions file should contain glob patterns")
	}
	if !strings.Contains(content, "Use React hooks.") {
		t.Error("copilot instructions file should contain inlined rule content")
	}
}

func TestClaudeScopedRules(t *testing.T) {
	dir := t.TempDir()
	cfg := testScopedConfig(t, dir)

	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude/rules/frontend.md"))
	if err != nil {
		t.Fatalf(".claude/rules/frontend.md not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "globs: **/*.tsx, **/*.css") {
		t.Error(".claude/rules/frontend.md should contain globs frontmatter")
	}
	if !strings.Contains(content, "@.agents/rules/frontend.md") {
		t.Error(".claude/rules/frontend.md should use @import syntax")
	}
}

func TestWindsurfScopedRules(t *testing.T) {
	dir := t.TempDir()
	cfg := testScopedConfig(t, dir)

	tr := &translators.WindsurfTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".windsurf/rules/frontend.md"))
	if err != nil {
		t.Fatalf(".windsurf/rules/frontend.md not generated: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "globs: **/*.tsx, **/*.css") {
		t.Error(".windsurf/rules/frontend.md should contain globs frontmatter")
	}
	if !strings.Contains(content, "Use React hooks.") {
		t.Error(".windsurf/rules/frontend.md should contain inlined rule content")
	}
}

func TestClineScopedRules(t *testing.T) {
	dir := t.TempDir()
	cfg := testScopedConfig(t, dir)

	tr := &translators.ClineTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".roo/rules/frontend.md"))
	if err != nil {
		t.Fatalf(".roo/rules/frontend.md not generated: %v", err)
	}
	if !strings.Contains(string(data), "Use React hooks.") {
		t.Error(".roo/rules/frontend.md should contain inlined rule content")
	}
}

func TestAiderScopedRulesInReadList(t *testing.T) {
	dir := t.TempDir()
	cfg := testScopedConfig(t, dir)

	tr := &translators.AiderTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".aider.conf.yml"))
	if !strings.Contains(string(data), ".agents/rules/frontend.md") {
		t.Error(".aider.conf.yml should include scoped rule path in read: list")
	}
}

func TestCursorImportScopedRules(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0o755)

	// User-authored scoped rule — has globs, not ajolote-generated
	os.WriteFile(filepath.Join(dir, ".cursor/rules/frontend.mdc"),
		[]byte("---\ndescription: frontend\nglobs: **/*.tsx, **/*.css\nalwaysApply: false\n---\n\nUse React hooks.\n"), 0o644)
	// Global rule — has no globs, should be skipped as scoped rule
	os.WriteFile(filepath.Join(dir, ".cursor/rules/other.mdc"),
		[]byte("---\ndescription: other\nalwaysApply: false\n---\n\nSome rule.\n"), 0o644)

	tr := &translators.CursorTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NewScopedRules) != 1 {
		t.Fatalf("expected 1 scoped rule, got %d", len(result.NewScopedRules))
	}
	sr := result.NewScopedRules[0]
	if sr.Name != "frontend" {
		t.Errorf("expected name 'frontend', got %q", sr.Name)
	}
	if len(sr.Globs) != 2 {
		t.Errorf("expected 2 globs, got %v", sr.Globs)
	}
}

func TestCopilotImportScopedRules(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".github", "instructions"), 0o755)

	os.WriteFile(filepath.Join(dir, ".github/instructions/frontend.instructions.md"),
		[]byte("---\napplyTo: \"**/*.tsx,**/*.css\"\n---\n\nUse React hooks.\n"), 0o644)

	tr := &translators.CopilotTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NewScopedRules) != 1 {
		t.Fatalf("expected 1 scoped rule, got %d", len(result.NewScopedRules))
	}
	sr := result.NewScopedRules[0]
	if sr.Name != "frontend" {
		t.Errorf("expected name 'frontend', got %q", sr.Name)
	}
	if len(sr.Globs) != 2 {
		t.Errorf("expected 2 globs, got %v", sr.Globs)
	}
}

func TestNoScopedRulesNoExtraFiles(t *testing.T) {
	// Configs without scoped_rules must not generate any extra files
	dir := t.TempDir()
	cfg := testConfig() // no ScopedRules

	for _, tr := range translators.All() {
		if err := tr.Generate(cfg, dir); err != nil {
			t.Fatalf("%s Generate: %v", tr.Name(), err)
		}
	}

	for _, path := range []string{
		".cursor/rules/frontend.mdc",
		".github/instructions/frontend.instructions.md",
		".claude/rules/frontend.md",
		".windsurf/rules/frontend.md",
		".roo/rules/frontend.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, path)); err == nil {
			t.Errorf("scoped rule file %s should not exist when config has no scoped_rules", path)
		}
	}
}

func TestAiderUsesReadList(t *testing.T) {
	dir := t.TempDir()
	tr := &translators.AiderTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".aider.conf.yml"))
	content := string(data)

	if !strings.Contains(content, "read:") {
		t.Error(".aider.conf.yml should contain a read: list")
	}
	if !strings.Contains(content, ".agents/rules/general.md") {
		t.Error(".aider.conf.yml read: list should include rules files")
	}
	if !strings.Contains(content, ".agents/skills/git.md") {
		t.Error(".aider.conf.yml read: list should include skills files")
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
	for _, name := range []string{"claude", "cursor", "windsurf", "copilot", "cline", "aider", "gemini", "codex", "agents-md"} {
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

func TestRoomodesGenerated(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()

	// Seed persona files so renderRoomodes can read them
	seedPersonaFile(t, dir, ".agents/personas/reviewer.md",
		"# Persona: Code Reviewer\n\nReview code carefully.\n")
	seedPersonaFile(t, dir, ".agents/personas/architect.md",
		"# Persona: Architect\n\nThink in trade-offs.\n")

	tr := &translators.ClineTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".roomodes"))
	if err != nil {
		t.Fatalf(".roomodes not generated: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, `"slug": "reviewer"`) {
		t.Error(".roomodes should contain reviewer slug")
	}
	if !strings.Contains(content, `"name": "Reviewer"`) {
		t.Error(".roomodes should contain title-cased name")
	}
	if !strings.Contains(content, "Review code carefully") {
		t.Error(".roomodes should embed persona file content as roleDefinition")
	}
	if !strings.Contains(content, `"source": "project"`) {
		t.Error(".roomodes should set source to project")
	}
	if !strings.Contains(content, `"read"`) || !strings.Contains(content, `"edit"`) {
		t.Error(".roomodes should include tool groups")
	}
	// Both personas should appear
	if !strings.Contains(content, `"slug": "architect"`) {
		t.Error(".roomodes should contain architect slug")
	}
}

func TestRoomodesEmptyWithoutPersonas(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	cfg.Personas = nil

	tr := &translators.ClineTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".roomodes"))
	if err != nil {
		t.Fatalf(".roomodes not generated: %v", err)
	}
	if !strings.Contains(string(data), `"customModes": []`) {
		t.Errorf(".roomodes with no personas should have empty customModes, got:\n%s", string(data))
	}
}

func TestRoomodesSkipsMissingPersonaFile(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	// Only seed one of the two persona files
	seedPersonaFile(t, dir, ".agents/personas/reviewer.md", "# Reviewer\n\nReview.\n")
	// .agents/personas/architect.md intentionally missing

	tr := &translators.ClineTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".roomodes"))
	content := string(data)
	if !strings.Contains(content, `"slug": "reviewer"`) {
		t.Error("reviewer should appear when its file exists")
	}
	if strings.Contains(content, `"slug": "architect"`) {
		t.Error("architect should be skipped when its file is missing")
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

// mcpScopingConfig returns a config with one project-scoped and one user-scoped MCP server.
func mcpScopingConfig() *config.Config {
	cfg := testConfig()
	cfg.MCP.Servers = map[string]config.MCPServer{
		"filesystem": {
			Command:     "npx",
			Args:        []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
			Description: "Read/write access to this repo",
			Scope:       "project",
		},
		"personal-figma": {
			Command:     "npx",
			Args:        []string{"-y", "figma-mcp"},
			Description: "Figma MCP (personal token)",
			Scope:       "user",
		},
	}
	return cfg
}

// httpMCPConfig returns a config with an HTTP-transport MCP server (no command/args).
func httpMCPConfig() *config.Config {
	cfg := testConfig()
	cfg.MCP.Servers = map[string]config.MCPServer{
		"remote": {
			Transport: "http",
			URL:       "https://mcp.example.com/api",
			Scope:     "project",
		},
	}
	return cfg
}

func TestClaudeMCPScoping(t *testing.T) {
	dir := t.TempDir()
	cfg := mcpScopingConfig()

	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	settings := string(data)

	// Project-scoped server should be present
	if !strings.Contains(settings, "filesystem") {
		t.Error("settings.json should contain project-scoped 'filesystem' server")
	}
	// User-scoped server must NOT be in project settings
	if strings.Contains(settings, "personal-figma") {
		t.Error("settings.json must not contain user-scoped 'personal-figma' server")
	}
	// Scope field must not appear in generated config
	if strings.Contains(settings, `"scope"`) {
		t.Error("settings.json must not contain 'scope' field")
	}
}

func TestCursorMCPScoping(t *testing.T) {
	dir := t.TempDir()
	cfg := mcpScopingConfig()

	tr := &translators.CursorTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".cursor", "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	mcpJSON := string(data)

	if !strings.Contains(mcpJSON, "filesystem") {
		t.Error(".cursor/mcp.json should contain project-scoped 'filesystem' server")
	}
	if strings.Contains(mcpJSON, "personal-figma") {
		t.Error(".cursor/mcp.json must not contain user-scoped 'personal-figma' server")
	}
	if strings.Contains(mcpJSON, `"scope"`) {
		t.Error(".cursor/mcp.json must not contain 'scope' field")
	}
}

func TestClineMCPScoping(t *testing.T) {
	dir := t.TempDir()
	cfg := mcpScopingConfig()

	tr := &translators.ClineTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".roo", "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	mcpJSON := string(data)

	if !strings.Contains(mcpJSON, "filesystem") {
		t.Error(".roo/mcp.json should contain project-scoped 'filesystem' server")
	}
	if strings.Contains(mcpJSON, "personal-figma") {
		t.Error(".roo/mcp.json must not contain user-scoped 'personal-figma' server")
	}
	if strings.Contains(mcpJSON, `"scope"`) {
		t.Error(".roo/mcp.json must not contain 'scope' field")
	}
}

func TestHTTPServerSerialization(t *testing.T) {
	dir := t.TempDir()
	cfg := httpMCPConfig()

	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	settings := string(data)

	if !strings.Contains(settings, "remote") {
		t.Error("settings.json should contain 'remote' server")
	}
	if !strings.Contains(settings, "https://mcp.example.com/api") {
		t.Error("settings.json should contain the server URL")
	}
	// HTTP server has no command — must not appear
	if strings.Contains(settings, `"command"`) {
		t.Error("HTTP server must not emit a 'command' field")
	}
	if strings.Contains(settings, `"scope"`) {
		t.Error("settings.json must not contain 'scope' field")
	}
}

func TestMergeUserMCPConfig(t *testing.T) {
	dir := t.TempDir()
	userMCPPath := filepath.Join(dir, ".claude.json")

	// Write an existing user config with one server already present
	existing := `{"mcpServers":{"existing-server":{"command":"existing","args":[]}}}`
	if err := os.WriteFile(userMCPPath, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		MCP: config.MCP{
			Servers: map[string]config.MCPServer{
				"existing-server": {Command: "different-command", Scope: "user"}, // already present — must not overwrite
				"new-server":      {Command: "new-cmd", Args: []string{"--flag"}, Scope: "user"},
			},
		},
	}

	tr := &translators.ClaudeTranslator{}
	// We test Generate() pointing at our fake home via the path that mergeUserMCPConfig writes to.
	// Since mergeUserMCPConfig writes to os.UserHomeDir()/.claude.json we can't redirect it here,
	// so we test the outcome via the cursor translator writing to a known temp path indirectly.
	// Instead, verify the scoping logic: project settings must not contain user servers.
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	settings, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(settings), "new-server") {
		t.Error("project settings.json must not contain user-scoped servers")
	}
	if strings.Contains(string(settings), "existing-server") {
		t.Error("project settings.json must not contain user-scoped servers")
	}
}

// seedPersonaFile writes a minimal persona markdown file to dir at the given path.
func seedPersonaFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// agentPersonaConfig returns a config where the reviewer persona has a claude: block.
func agentPersonaConfig() *config.Config {
	cfg := testConfig()
	cfg.Personas = []config.Persona{
		{
			Path: ".agents/personas/reviewer.md",
			Claude: &config.ClaudeAgent{
				Model:       "haiku",
				Tools:       []string{"Read", "Grep", "Glob"},
				Description: "Expert code reviewer focused on correctness and coverage.",
			},
		},
		{Path: ".agents/personas/architect.md"}, // no claude block
	}
	return cfg
}

func TestClaudeAgentFileGenerated(t *testing.T) {
	dir := t.TempDir()
	cfg := agentPersonaConfig()

	seedPersonaFile(t, dir, ".agents/personas/reviewer.md", "# Persona: Code Reviewer\n\nReview code with care.\n")
	seedPersonaFile(t, dir, ".agents/personas/architect.md", "# Persona: Architect\n\nThink in trade-offs.\n")

	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	agentPath := filepath.Join(dir, ".claude", "agents", "reviewer.md")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("expected .claude/agents/reviewer.md to be created: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "name: Reviewer") {
		t.Errorf("agent file should contain 'name: Reviewer' (derived from filename), got:\n%s", content)
	}
	if !strings.Contains(content, "description: Expert code reviewer focused on correctness and coverage.") {
		t.Errorf("agent file should contain explicit description, got:\n%s", content)
	}
	if !strings.Contains(content, "model: claude-haiku-4-5-20251001") {
		t.Errorf("agent file should contain resolved model ID, got:\n%s", content)
	}
	if !strings.Contains(content, "  - Read") || !strings.Contains(content, "  - Grep") {
		t.Errorf("agent file should contain tools list, got:\n%s", content)
	}
	if !strings.Contains(content, "@.agents/personas/reviewer.md") {
		t.Errorf("agent file should @-import the persona file, got:\n%s", content)
	}
}

func TestClaudeAgentModelResolution(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	cfg.Personas = []config.Persona{
		{Path: ".agents/personas/reviewer.md", Claude: &config.ClaudeAgent{Model: "haiku"}},
		{Path: ".agents/personas/architect.md", Claude: &config.ClaudeAgent{Model: "claude-opus-4-6"}},
	}
	seedPersonaFile(t, dir, ".agents/personas/reviewer.md", "# Reviewer\n\nReview.\n")
	seedPersonaFile(t, dir, ".agents/personas/architect.md", "# Architect\n\nDesign.\n")

	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	reviewerData, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "reviewer.md"))
	if !strings.Contains(string(reviewerData), "model: claude-haiku-4-5-20251001") {
		t.Errorf("'haiku' shorthand should resolve to full ID, got:\n%s", string(reviewerData))
	}

	architectData, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "architect.md"))
	if !strings.Contains(string(architectData), "model: claude-opus-4-6") {
		t.Errorf("full model ID should pass through unchanged, got:\n%s", string(architectData))
	}
}

func TestClaudeAgentDescriptionFallback(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	cfg.Personas = []config.Persona{
		{
			Path:   ".agents/personas/reviewer.md",
			Claude: &config.ClaudeAgent{Model: "sonnet"}, // no Description set
		},
	}
	// First non-heading, non-blank line should be extracted as description
	seedPersonaFile(t, dir, ".agents/personas/reviewer.md",
		"# Persona: Code Reviewer\n\nWhen acting as a code reviewer, adopt these priorities.\n")

	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "agents", "reviewer.md"))
	if !strings.Contains(string(data), "description: When acting as a code reviewer") {
		t.Errorf("should derive description from persona file first paragraph, got:\n%s", string(data))
	}
}

func TestClaudeNoAgentWithoutBlock(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig() // default config has no claude: block on any persona

	seedPersonaFile(t, dir, ".agents/personas/reviewer.md", "# Reviewer\n\nReview.\n")
	seedPersonaFile(t, dir, ".agents/personas/architect.md", "# Architect\n\nDesign.\n")

	tr := &translators.ClaudeTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	agentsDir := filepath.Join(dir, ".claude", "agents")
	if _, err := os.Stat(agentsDir); !os.IsNotExist(err) {
		t.Error(".claude/agents/ should not be created when no persona has a claude: block")
	}
}

func TestAllTranslatorsWithAgentPersona(t *testing.T) {
	cfg := agentPersonaConfig()
	for _, tr := range translators.All() {
		tr := tr
		t.Run(tr.Name(), func(t *testing.T) {
			dir := t.TempDir()
			seedPersonaFile(t, dir, ".agents/personas/reviewer.md", "# Reviewer\n\nReview.\n")
			seedPersonaFile(t, dir, ".agents/personas/architect.md", "# Architect\n\nDesign.\n")
			if err := tr.Generate(cfg, dir); err != nil {
				t.Fatalf("%s Generate with agent persona: %v", tr.Name(), err)
			}
		})
	}
}

// --- Gemini CLI translator tests ---

func TestGeminiGEMINIMdGenerated(t *testing.T) {
	dir := t.TempDir()
	tr := &translators.GeminiTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "GEMINI.md"))
	if err != nil {
		t.Fatalf("GEMINI.md not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "@.agents/rules/general.md") {
		t.Error("GEMINI.md should contain @file import for general.md")
	}
}

func TestGeminiUsesAtFileImports(t *testing.T) {
	dir := t.TempDir()
	tr := &translators.GeminiTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "GEMINI.md"))
	content := string(data)

	// Should use @file imports, not inlined content
	if !strings.Contains(content, "@.agents/") {
		t.Error("GEMINI.md should use @file imports")
	}
	// Should not inline file contents (no "# General" heading from general.md)
	if strings.Contains(content, "Always read before writing") {
		t.Error("GEMINI.md should not inline file content — use @file imports")
	}
}

// --- Codex CLI translator tests ---

func TestCodexTOMLAlwaysCreated(t *testing.T) {
	// .codex/config.toml is always written (even without MCP servers) so the
	// translator is detectable as active by ajolote sync and ajolote diff.
	dir := t.TempDir()
	tr := &translators.CodexTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".codex", "config.toml")); err != nil {
		t.Error(".codex/config.toml should always be created")
	}
}

func TestCodexTOMLGenerated(t *testing.T) {
	dir := t.TempDir()
	cfg := testConfig()
	cfg.MCP.Servers["shell"] = config.MCPServer{
		Command: "npx",
		Args:    []string{"-y", "@some/mcp-server"},
	}

	tr := &translators.CodexTranslator{}
	if err := tr.Generate(cfg, dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf(".codex/config.toml not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "[mcp.servers.shell]") {
		t.Error(".codex/config.toml should contain [mcp.servers.shell] entry")
	}
	if !strings.Contains(content, `command = "npx"`) {
		t.Error(".codex/config.toml should contain command field")
	}
}

func TestRegistryIncludesGeminiCodexAndAgentsMD(t *testing.T) {
	for _, name := range []string{"gemini", "codex", "agents-md"} {
		if _, err := translators.Get(name); err != nil {
			t.Errorf("Get(%q) failed: %v", name, err)
		}
	}
}

// --- agents-md translator tests ---

func TestAgentsMDGenerated(t *testing.T) {
	dir := t.TempDir()
	// Write rule file so inlineFiles can read it
	ruleDir := filepath.Join(dir, ".agents", "rules")
	if err := os.MkdirAll(ruleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ruleDir, "general.md"), []byte("# General\n\nAlways read before writing.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	tr := &translators.AgentsMDTranslator{}
	if err := tr.Generate(testConfig(), dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md not created: %v", err)
	}
	if !strings.Contains(string(data), "Always read before writing") {
		t.Error("AGENTS.md should contain inlined rule content")
	}
}

func TestAgentsMDIsCommitted(t *testing.T) {
	tr := &translators.AgentsMDTranslator{}
	co, ok := any(tr).(translators.CommittedOutput)
	if !ok {
		t.Fatal("AgentsMDTranslator should implement CommittedOutput")
	}
	if !co.Committed() {
		t.Error("AgentsMDTranslator.Committed() should return true")
	}
}

func TestAgentsMDImportBootstraps(t *testing.T) {
	dir := t.TempDir()
	// Place an existing user-authored AGENTS.md
	content := "# Project Rules\n\nAlways write tests.\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	tr := &translators.AgentsMDTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || result.NewRuleFiles["general.md"] == "" {
		t.Error("Import should return existing AGENTS.md content as a rule file")
	}
	if !strings.Contains(result.NewRuleFiles["general.md"], "Always write tests") {
		t.Error("imported content should match original AGENTS.md")
	}
}

func TestAgentsMDImportSkipsAjoloteGenerated(t *testing.T) {
	dir := t.TempDir()
	// ajolote-generated AGENTS.md should not be re-imported
	content := "<!-- AGENTS.md — Generated by ajolote-ai from `.agents/config.json` -->\n\n## Rules\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	tr := &translators.AgentsMDTranslator{}
	result, err := tr.Import(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NewRuleFiles) > 0 {
		t.Error("Import should skip ajolote-generated AGENTS.md")
	}
}
