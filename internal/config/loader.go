package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// safeName matches identifiers safe for use in TOML table headers, YAML
// frontmatter, and filesystem paths. Allows letters, digits, hyphens,
// underscores, and dots (for dotted command names like "speckit.analyze").
var safeName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)

// safeEnvKey matches valid environment variable names.
var safeEnvKey = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

const ConfigPath = ".agents/config.json"

// maxConfigSize is the maximum allowed size for config.json (1 MB).
const maxConfigSize = 1 << 20

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

	if len(data) > maxConfigSize {
		return nil, fmt.Errorf("config file exceeds maximum size (%d bytes)", maxConfigSize)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.validatePaths(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validatePaths checks that all file paths in the config are relative, stay
// within the project directory, and don't use path traversal (SEC-001).
func (cfg *Config) validatePaths() error {
	for _, p := range cfg.Rules {
		if err := validatePath(p); err != nil {
			return fmt.Errorf("rules: %w", err)
		}
	}
	for _, p := range cfg.Skills {
		if err := validatePath(p); err != nil {
			return fmt.Errorf("skills: %w", err)
		}
	}
	for _, p := range cfg.Context {
		if err := validatePath(p); err != nil {
			return fmt.Errorf("context: %w", err)
		}
	}
	for _, p := range cfg.Personas {
		if err := validatePath(p.Path); err != nil {
			return fmt.Errorf("personas: %w", err)
		}
	}
	for _, sr := range cfg.ScopedRules {
		if err := validatePath(sr.Path); err != nil {
			return fmt.Errorf("scoped_rules[%s]: %w", sr.Name, err)
		}
		// SEC-005: Validate scoped rule names for safe use in filenames and YAML frontmatter.
		if sr.Name != "" && !safeName.MatchString(sr.Name) {
			return fmt.Errorf("scoped_rules name %q contains invalid characters (allowed: letters, digits, hyphens, underscores, dots)", sr.Name)
		}
	}
	// SEC-003: Validate MCP server names and env keys for safe use in TOML/JSON output.
	for name, srv := range cfg.MCP.Servers {
		if !safeName.MatchString(name) {
			return fmt.Errorf("mcp server name %q contains invalid characters (allowed: letters, digits, hyphens, underscores, dots)", name)
		}
		for key := range srv.Env {
			if !safeEnvKey.MatchString(key) {
				return fmt.Errorf("mcp server %q env key %q contains invalid characters (allowed: letters, digits, underscores)", name, key)
			}
		}
	}
	return nil
}

// validatePath rejects absolute paths and paths that traverse outside the
// project directory using ".." segments.
func validatePath(p string) error {
	if p == "" {
		return fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(p) {
		return fmt.Errorf("path %q must be relative, not absolute", p)
	}
	clean := filepath.Clean(p)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path %q must not traverse outside the project directory", p)
	}
	return nil
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
