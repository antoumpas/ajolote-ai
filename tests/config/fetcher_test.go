package config_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/config"
)

// TestLocalFetcherCopiesAgentsDir verifies that the local fetcher copies the
// base source's .agents/ directory into destDir.
func TestLocalFetcherCopiesAgentsDir(t *testing.T) {
	// Create a fake base project with a .agents/ directory.
	baseProject := t.TempDir()
	agentsDir := filepath.Join(baseProject, ".agents")
	os.MkdirAll(filepath.Join(agentsDir, "rules"), 0o755)
	os.WriteFile(filepath.Join(agentsDir, "config.json"), []byte(`{"mcp":{"servers":{}},"rules":[".agents/rules/org.md"],"skills":[],"personas":[],"context":[]}`), 0o644)
	os.WriteFile(filepath.Join(agentsDir, "rules", "org.md"), []byte("# Org Rules"), 0o644)

	destDir := t.TempDir()

	if err := config.FetchAgentsDir(baseProject, destDir); err != nil {
		t.Fatalf("FetchAgentsDir: %v", err)
	}

	// Verify config.json was copied
	if _, err := os.Stat(filepath.Join(destDir, "config.json")); err != nil {
		t.Error("config.json not found in destDir")
	}

	// Verify rules/org.md was copied
	if _, err := os.Stat(filepath.Join(destDir, "rules", "org.md")); err != nil {
		t.Error("rules/org.md not found in destDir")
	}
}

// TestLocalFetcherAbsolutePath verifies absolute paths are handled correctly.
func TestLocalFetcherAbsolutePath(t *testing.T) {
	baseProject := t.TempDir()
	agentsDir := filepath.Join(baseProject, ".agents")
	os.MkdirAll(agentsDir, 0o755)
	os.WriteFile(filepath.Join(agentsDir, "config.json"), []byte(`{"mcp":{"servers":{}},"rules":[],"skills":[],"personas":[],"context":[]}`), 0o644)

	destDir := t.TempDir()
	if err := config.FetchAgentsDir(baseProject, destDir); err != nil {
		t.Fatalf("FetchAgentsDir with absolute path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "config.json")); err != nil {
		t.Error("config.json not found")
	}
}

// TestLocalFetcherMissingAgentsDir returns an error when the source has no .agents/.
func TestLocalFetcherMissingAgentsDir(t *testing.T) {
	empty := t.TempDir()
	destDir := t.TempDir()
	if err := config.FetchAgentsDir(empty, destDir); err == nil {
		t.Fatal("expected error for missing .agents/ dir, got nil")
	}
}

// TestHTTPFetcherGetsConfigAndFiles uses an httptest server to verify the
// HTTP fetcher downloads config.json and all referenced files.
func TestHTTPFetcherGetsConfigAndFiles(t *testing.T) {
	files := map[string]string{
		"/.agents/config.json":       `{"mcp":{"servers":{}},"rules":[".agents/rules/org.md"],"skills":[],"personas":[],"context":[]}`,
		"/.agents/rules/org.md":      "# Org Rules",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(content))
	}))
	defer srv.Close()

	destDir := t.TempDir()
	if err := config.FetchAgentsDir(srv.URL, destDir); err != nil {
		t.Fatalf("FetchAgentsDir (HTTP): %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "config.json")); err != nil {
		t.Error("config.json not found")
	}
	if _, err := os.Stat(filepath.Join(destDir, "rules", "org.md")); err != nil {
		t.Error("rules/org.md not found")
	}
}

// TestHTTPFetcherConfigNotFound returns an error when config.json is missing.
func TestHTTPFetcherConfigNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	destDir := t.TempDir()
	if err := config.FetchAgentsDir(srv.URL, destDir); err == nil {
		t.Fatal("expected error for missing config.json, got nil")
	}
}

// TestFetchSourceUnsupportedScheme returns an error for unknown schemes.
func TestFetchSourceUnsupportedScheme(t *testing.T) {
	destDir := t.TempDir()
	if err := config.FetchAgentsDir("s3://bucket/path", destDir); err == nil {
		t.Fatal("expected error for unsupported scheme, got nil")
	}
}

// TestCanFetchSource verifies which schemes are recognised.
func TestCanFetchSource(t *testing.T) {
	cases := []struct {
		source string
		want   bool
	}{
		{"/absolute/path", true},
		{"./relative/path", true},
		{"../parent/path", true},
		{"file:///some/path", true},
		{"https://example.com/project", true},
		{"http://example.com/project", true},
		{"https://github.com/org/repo.git", true},
		{"git@github.com:org/repo.git", true},
		{"ftp://ftp.example.com/standards", true},
		{"s3://bucket/key", false},
		{"ssh://host/path", false},
	}

	for _, tc := range cases {
		got := config.CanFetchSource(tc.source)
		if got != tc.want {
			t.Errorf("CanFetchSource(%q) = %v, want %v", tc.source, got, tc.want)
		}
	}
}

// TestHTTPFetcherSkipsCommands verifies that the HTTPS fetcher does NOT fetch
// command files unless they are listed in the config's commands field (since
// directories cannot be listed over HTTP).
func TestHTTPFetcherSkipsUnlistedCommands(t *testing.T) {
	files := map[string]string{
		"/.agents/config.json":            `{"mcp":{"servers":{}},"rules":[],"skills":[],"personas":[],"context":[],"commands":[".agents/commands/review.md"]}`,
		"/.agents/commands/review.md":     "# Review",
		"/.agents/commands/unlisted.md":   "# Unlisted",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(content))
	}))
	defer srv.Close()

	destDir := t.TempDir()
	if err := config.FetchAgentsDir(srv.URL, destDir); err != nil {
		t.Fatalf("FetchAgentsDir: %v", err)
	}

	// Listed command should be present
	if _, err := os.Stat(filepath.Join(destDir, "commands", "review.md")); err != nil {
		t.Error("commands/review.md should have been fetched (it is listed in commands)")
	}

	// Unlisted command should NOT be present (HTTPS can't list directories)
	if _, err := os.Stat(filepath.Join(destDir, "commands", "unlisted.md")); err == nil {
		t.Error("commands/unlisted.md should NOT have been fetched (not in commands list)")
	}
}
