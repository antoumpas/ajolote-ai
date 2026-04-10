package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/cli/commands"
)

// runInit executes the init command in dir and returns any error.
func runInit(t *testing.T, dir string) error {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	cmd := commands.InitCmd()
	cmd.SetArgs(nil)
	return cmd.Execute()
}

func TestInitImportsCommandsFromClaude(t *testing.T) {
	// Regression: init silently skipped commands when settings.json was absent.
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude/commands/deploy.md"),
		[]byte("---\ndescription: Deploy to staging\n---\n\nRun deploy.sh\n"), 0o644)
	os.WriteFile(filepath.Join(dir, ".claude/commands/speckit.analyze.md"),
		[]byte("---\ndescription: Analyze spec\n---\n\nAnalyze the spec.\n"), 0o644)

	if err := runInit(t, dir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	for _, name := range []string{"deploy", "speckit.analyze"} {
		path := filepath.Join(dir, ".agents", "commands", name+".md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected .agents/commands/%s.md to be imported, but it is missing", name)
		}
	}
}

func TestInitImportsMCPServersFromClaude(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
	settingsJSON := `{"mcpServers":{
		"github":{"command":"npx","args":["-y","@modelcontextprotocol/server-github"]},
		"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}
	}}`
	os.WriteFile(filepath.Join(dir, ".claude/settings.json"), []byte(settingsJSON), 0o644)

	if err := runInit(t, dir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".agents", "config.json"))
	if err != nil {
		t.Fatal("config.json not created")
	}
	var cfg struct {
		MCP struct {
			Servers map[string]interface{} `json:"servers"`
		} `json:"mcp"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"github", "filesystem"} {
		if _, ok := cfg.MCP.Servers[name]; !ok {
			t.Errorf("expected MCP server %q to be imported into config.json", name)
		}
	}
}

func TestInitDeduplicatesCommandsAcrossTools(t *testing.T) {
	// If both Claude and Cursor have a command with the same name,
	// only one file should be written to .agents/commands/.
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755)
	os.WriteFile(filepath.Join(dir, ".claude/commands/deploy.md"),
		[]byte("Claude version of deploy\n"), 0o644)

	os.MkdirAll(filepath.Join(dir, ".cursor", "rules"), 0o755)
	os.WriteFile(filepath.Join(dir, ".cursor/mcp.json"), []byte(`{"mcpServers":{}}`), 0o644)
	os.WriteFile(filepath.Join(dir, ".cursor/rules/deploy.mdc"),
		[]byte("---\ndescription: Deploy\nalwaysApply: false\n---\n\nCursor version of deploy\n"), 0o644)

	if err := runInit(t, dir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(dir, ".agents", "commands"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, e := range entries {
		if e.Name() == "deploy.md" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 deploy.md in .agents/commands/, found %d", count)
	}
}

func TestInitSkipsWhenConfigExists(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".agents")
	os.MkdirAll(agentsDir, 0o755)
	os.WriteFile(filepath.Join(agentsDir, "config.json"), []byte(`{"project":{"name":"existing"}}`), 0o644)

	if err := runInit(t, dir); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Original config must not be overwritten
	data, _ := os.ReadFile(filepath.Join(agentsDir, "config.json"))
	var cfg struct {
		Project struct{ Name string } `json:"project"`
	}
	json.Unmarshal(data, &cfg)
	if cfg.Project.Name != "existing" {
		t.Errorf("init overwrote existing config.json (project name changed to %q)", cfg.Project.Name)
	}
}
