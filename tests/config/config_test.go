package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ajolote-ai/ajolote/internal/config"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.MCP.Servers["test"] = config.MCPServer{Command: "npx", Args: []string{"test"}}

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

	if _, ok := loaded.MCP.Servers["test"]; !ok {
		t.Error("expected MCP server 'test' to survive save/load round-trip")
	}
	if len(loaded.Rules) == 0 {
		t.Error("expected Rules to be non-empty after round-trip")
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

	if err := config.Save(dir, config.DefaultConfig()); err != nil {
		t.Fatal(err)
	}

	if !config.Exists(dir) {
		t.Fatal("should exist after save")
	}
}
