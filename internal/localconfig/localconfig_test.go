package localconfig_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/localconfig"
)

// ── Load ─────────────────────────────────────────────────────────────────────

func TestLoadAbsent(t *testing.T) {
	dir := t.TempDir()
	lc, err := localconfig.Load(dir)
	if err != nil {
		t.Fatalf("expected no error for absent file, got: %v", err)
	}
	if lc == nil {
		t.Fatal("expected non-nil config for absent file")
	}
	if len(lc.Protect) != 0 {
		t.Errorf("expected empty protect list, got: %v", lc.Protect)
	}
}

func TestLoadPresent(t *testing.T) {
	dir := t.TempDir()
	writeLocalConfig(t, dir, &localconfig.LocalConfig{
		Protect: []string{"CLAUDE.md", ".cursor/rules/*.md"},
	})

	lc, err := localconfig.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lc.Protect) != 2 {
		t.Errorf("expected 2 protect entries, got %d: %v", len(lc.Protect), lc.Protect)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeRaw(t, dir, "not valid json {{{")

	_, err := localconfig.Load(dir)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

// ── IsProtected ───────────────────────────────────────────────────────────────

func TestIsProtectedNilConfig(t *testing.T) {
	var lc *localconfig.LocalConfig
	if lc.IsProtected("CLAUDE.md") {
		t.Error("nil config should never protect anything")
	}
}

func TestIsProtectedEmptyProtectList(t *testing.T) {
	lc := &localconfig.LocalConfig{}
	if lc.IsProtected("CLAUDE.md") {
		t.Error("empty protect list should not protect anything")
	}
}

func TestIsProtectedExactMatch(t *testing.T) {
	lc := &localconfig.LocalConfig{Protect: []string{"CLAUDE.md"}}
	if !lc.IsProtected("CLAUDE.md") {
		t.Error("exact match should be protected")
	}
}

func TestIsProtectedExactMatchNegative(t *testing.T) {
	lc := &localconfig.LocalConfig{Protect: []string{"CLAUDE.md"}}
	if lc.IsProtected("cursor.md") {
		t.Error("different file should not be protected")
	}
}

func TestIsProtectedGlob(t *testing.T) {
	lc := &localconfig.LocalConfig{Protect: []string{".claude/commands/*.md"}}
	cases := []struct {
		path      string
		protected bool
	}{
		{".claude/commands/review.md", true},
		{".claude/commands/deploy.md", true},
		{".claude/commands/nested/review.md", false}, // * doesn't cross dirs
		{".cursor/commands/review.md", false},
		{"CLAUDE.md", false},
	}
	for _, c := range cases {
		got := lc.IsProtected(c.path)
		if got != c.protected {
			t.Errorf("IsProtected(%q) = %v, want %v", c.path, got, c.protected)
		}
	}
}

func TestIsProtectedDirectoryPrefix(t *testing.T) {
	lc := &localconfig.LocalConfig{Protect: []string{".claude/commands/"}}
	cases := []struct {
		path      string
		protected bool
	}{
		{".claude/commands/review.md", true},
		{".claude/commands/deploy.md", true},
		{".claude/commands/sub/file.md", true}, // directory prefix catches all depths
		{".claude/settings.json", false},
		{"CLAUDE.md", false},
	}
	for _, c := range cases {
		got := lc.IsProtected(c.path)
		if got != c.protected {
			t.Errorf("IsProtected(%q) = %v, want %v", c.path, got, c.protected)
		}
	}
}

func TestIsProtectedMultiplePatterns(t *testing.T) {
	lc := &localconfig.LocalConfig{Protect: []string{
		"CLAUDE.md",
		".cursor/rules/my-rule.md",
	}}
	if !lc.IsProtected("CLAUDE.md") {
		t.Error("CLAUDE.md should be protected")
	}
	if !lc.IsProtected(".cursor/rules/my-rule.md") {
		t.Error(".cursor/rules/my-rule.md should be protected")
	}
	if lc.IsProtected(".cursor/rules/other.md") {
		t.Error(".cursor/rules/other.md should not be protected")
	}
}

func TestIsProtectedNoMatch(t *testing.T) {
	lc := &localconfig.LocalConfig{Protect: []string{"CLAUDE.md"}}
	unrelated := []string{
		".cursor/mcp.json",
		".claude/settings.json",
		"AGENTS.md",
		".cursorrules",
	}
	for _, p := range unrelated {
		if lc.IsProtected(p) {
			t.Errorf("unexpected protection for %q", p)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeLocalConfig(t *testing.T, dir string, lc *localconfig.LocalConfig) {
	t.Helper()
	agentsDir := filepath.Join(dir, ".agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(lc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "config.local.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeRaw(t *testing.T, dir, content string) {
	t.Helper()
	agentsDir := filepath.Join(dir, ".agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "config.local.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
