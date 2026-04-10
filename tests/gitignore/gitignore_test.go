package gitignore_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/gitignore"
)

func TestUpdateCreatesFile(t *testing.T) {
	dir := t.TempDir()

	entries := []string{"CLAUDE.md", ".claude/settings.json"}
	if err := gitignore.Update(dir, entries); err != nil {
		t.Fatalf("Update: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}

	content := string(data)
	for _, e := range entries {
		if !strings.Contains(content, e) {
			t.Errorf("expected %q in .gitignore", e)
		}
	}
}

func TestUpdateIsIdempotent(t *testing.T) {
	dir := t.TempDir()

	entries := []string{"CLAUDE.md", ".cursor/mcp.json"}

	for i := 0; i < 3; i++ {
		if err := gitignore.Update(dir, entries); err != nil {
			t.Fatalf("Update iteration %d: %v", i, err)
		}
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	content := string(data)

	// Each entry should appear exactly once
	for _, e := range entries {
		count := strings.Count(content, e)
		if count != 1 {
			t.Errorf("entry %q appears %d times, want exactly 1", e, count)
		}
	}
}

func TestUpdatePreservesExistingLines(t *testing.T) {
	dir := t.TempDir()

	existing := "node_modules/\ndist/\n.env\n"
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := gitignore.Update(dir, []string{"CLAUDE.md"}); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	for _, line := range []string{"node_modules/", "dist/", ".env"} {
		if !strings.Contains(content, line) {
			t.Errorf("existing line %q was removed", line)
		}
	}
}

func TestEntriesInBlock(t *testing.T) {
	dir := t.TempDir()

	entries := []string{"CLAUDE.md", ".cursor/mcp.json"}
	if err := gitignore.Update(dir, entries); err != nil {
		t.Fatal(err)
	}

	got, ok := gitignore.EntriesInBlock(dir)
	if !ok {
		t.Fatal("expected block to exist")
	}

	if len(got) != len(entries) {
		t.Errorf("got %d entries, want %d", len(got), len(entries))
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, ".gitignore")
	existing := "dist/\n.env\n"
	os.WriteFile(path, []byte(existing), 0o644)

	gitignore.Update(dir, []string{"CLAUDE.md"})
	gitignore.Remove(dir)

	data, _ := os.ReadFile(path)
	content := string(data)

	if strings.Contains(content, "CLAUDE.md") {
		t.Error("CLAUDE.md should have been removed")
	}
	if !strings.Contains(content, "dist/") {
		t.Error("dist/ should still be present")
	}
}
