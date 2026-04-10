package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/config"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig("test-project")
	cfg.Project.Language = "Go"

	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, config.ConfigPath)); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Project.Name != "test-project" {
		t.Errorf("got name %q, want %q", loaded.Project.Name, "test-project")
	}
	if loaded.Project.Language != "Go" {
		t.Errorf("got language %q, want %q", loaded.Project.Language, "Go")
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := config.Load(dir)
	if err == nil {
		t.Fatal("expected error loading from empty dir")
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if config.Exists(dir) {
		t.Fatal("should not exist yet")
	}

	if err := config.Save(dir, config.DefaultConfig("test")); err != nil {
		t.Fatal(err)
	}

	if !config.Exists(dir) {
		t.Fatal("should exist after save")
	}
}
