package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ConfigPath = ".agents/config.json"

// Load reads and parses .agents/config.json from the given project root.
func Load(projectRoot string) (*Config, error) {
	path := filepath.Join(projectRoot, ConfigPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no ajolote config found at %s — run `ajolote init` first", ConfigPath)
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Tools == nil {
		cfg.Tools = map[string]bool{}
	}

	return &cfg, nil
}

// Save writes cfg to .agents/config.json under projectRoot.
func Save(projectRoot string, cfg *Config) error {
	dir := filepath.Join(projectRoot, ".agents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating .agents dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	path := filepath.Join(projectRoot, ConfigPath)
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// Exists reports whether .agents/config.json exists under projectRoot.
func Exists(projectRoot string) bool {
	_, err := os.Stat(filepath.Join(projectRoot, ConfigPath))
	return err == nil
}
