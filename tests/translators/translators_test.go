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
	cfg := config.DefaultConfig("test-app")
	cfg.Project.Language = "Go"
	cfg.Project.Stack = "Go / Gin / PostgreSQL"
	cfg.Project.TestRunner = "go test"
	cfg.Rules.General = []string{"Read before writing."}
	return cfg
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
				data, err := os.ReadFile(filepath.Join(dir, f))
				if err != nil {
					t.Errorf("expected output file %s: %v", f, err)
					continue
				}
				if len(data) == 0 {
					t.Errorf("output file %s is empty", f)
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

	if !strings.Contains(content, "test-app") {
		t.Error("CLAUDE.md should contain project name")
	}
	if !strings.Contains(content, "Read before writing.") {
		t.Error("CLAUDE.md should contain rules")
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
