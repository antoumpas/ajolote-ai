package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// safeName matches identifiers safe for use in TOML table headers, YAML
// frontmatter, and filesystem paths. Allows letters, digits, hyphens,
// underscores, and dots (for dotted command names like "speckit.analyze").
var safeName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)

// safeEnvKey matches valid environment variable names.
var safeEnvKey = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

const ConfigPath = ".agents/config.json"

// baseCacheDir is the directory inside .agents/ where inherited files are cached.
const baseCacheDir = ".agents/.base"

// defaultCacheTTL is how long the base config cache remains fresh.
const defaultCacheTTL = time.Hour

// maxConfigSize is the maximum allowed size for config.json (1 MB).
const maxConfigSize = 1 << 20

// cacheMeta is stored as .agents/.base/.meta.json to track cache freshness.
type cacheMeta struct {
	Source    string `json:"source"`
	FetchedAt string `json:"fetched_at"`
}

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

// Resolve returns the effective config after applying inheritance. If cfg.Extends
// is set, the base project's .agents/ directory is fetched (and cached under
// .agents/.base/), its config is merged with the local config, and the merged
// result is returned. Local values always win over inherited ones.
//
// Cache TTL defaults to 1 hour. Override with AJOLOTE_CACHE_TTL_SECONDS env var
// (set to 0 to always re-fetch).
//
// If cfg.Extends is empty, cfg is returned unchanged.
func Resolve(cfg *Config, projectRoot string) (*Config, error) {
	if cfg.Extends == "" {
		return cfg, nil
	}

	source := resolveSource(cfg.Extends, projectRoot)
	destDir := filepath.Join(projectRoot, baseCacheDir)

	if err := refreshBaseCache(source, destDir); err != nil {
		return nil, fmt.Errorf("fetching base config from %q: %w", cfg.Extends, err)
	}

	// Parse the cached base config.json
	baseCfgPath := filepath.Join(destDir, "config.json")
	baseCfgData, err := os.ReadFile(baseCfgPath)
	if err != nil {
		return nil, fmt.Errorf("reading cached base config: %w", err)
	}
	if len(baseCfgData) > maxConfigSize {
		return nil, fmt.Errorf("base config exceeds maximum size (%d bytes)", maxConfigSize)
	}

	var baseCfg Config
	if err := json.Unmarshal(baseCfgData, &baseCfg); err != nil {
		return nil, fmt.Errorf("parsing base config: %w", err)
	}

	// Rewrite all base config file paths from ".agents/..." to ".agents/.base/..."
	// so they resolve correctly inside the inheriting project.
	rebased := RebaseConfigPaths(&baseCfg)

	return MergeConfigs(rebased, cfg), nil
}

// resolveSource converts a relative local path to an absolute path using
// projectRoot as the base. URLs and git@ addresses are returned unchanged.
func resolveSource(source, projectRoot string) string {
	if strings.Contains(source, "://") || strings.HasPrefix(source, "git@") {
		return source
	}
	if !filepath.IsAbs(source) {
		return filepath.Join(projectRoot, source)
	}
	return source
}

// refreshBaseCache ensures .agents/.base/ contains a fresh copy of the base
// source's .agents/ directory. Returns immediately if the cache is still within TTL.
func refreshBaseCache(source, destDir string) error {
	ttl := defaultCacheTTL
	if s := os.Getenv("AJOLOTE_CACHE_TTL_SECONDS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			ttl = time.Duration(n) * time.Second
		}
	}

	// Check whether the cache is still fresh.
	if ttl > 0 {
		metaPath := filepath.Join(destDir, ".meta.json")
		if data, err := os.ReadFile(metaPath); err == nil {
			var meta cacheMeta
			if json.Unmarshal(data, &meta) == nil && meta.Source == source {
				if t, err := time.Parse(time.RFC3339, meta.FetchedAt); err == nil {
					if time.Since(t) < ttl {
						return nil // cache is fresh
					}
				}
			}
		}
	}

	// Fetch into a temporary sibling directory so we can atomically swap it in.
	tmpDir, err := os.MkdirTemp(filepath.Dir(destDir), ".base-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}

	fetchErr := FetchAgentsDir(source, tmpDir)
	if fetchErr != nil {
		os.RemoveAll(tmpDir)
		// If a stale cache already exists, warn and reuse it rather than failing.
		if _, statErr := os.Stat(filepath.Join(destDir, "config.json")); statErr == nil {
			fmt.Fprintf(os.Stderr, "warning: could not refresh base config from %q (using cached copy): %v\n", source, fetchErr)
			return nil
		}
		return fetchErr
	}

	// Write cache metadata.
	meta := cacheMeta{Source: source, FetchedAt: time.Now().UTC().Format(time.RFC3339)}
	if metaData, err := json.Marshal(meta); err == nil {
		os.WriteFile(filepath.Join(tmpDir, ".meta.json"), metaData, 0o644)
	}

	// Atomically replace the old cache with the new one.
	os.RemoveAll(destDir)
	if err := os.Rename(tmpDir, destDir); err != nil {
		// Rename can fail across device boundaries; fall back to a recursive copy.
		os.RemoveAll(destDir)
		if err2 := copyDirContents(tmpDir, destDir); err2 != nil {
			os.RemoveAll(tmpDir)
			return fmt.Errorf("installing base cache: %w", err2)
		}
		os.RemoveAll(tmpDir)
	}
	return nil
}

// copyDirContents recursively copies all files from src into dst.
func copyDirContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

// validatePaths checks that all file paths in the config are relative, stay
// within the project directory, and don't use path traversal (SEC-001).
// Note: the Extends field is a URL or external path and is NOT validated here.
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
	for _, p := range cfg.Commands {
		if err := validatePath(p); err != nil {
			return fmt.Errorf("commands: %w", err)
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
	if filepath.IsAbs(p) || p[0] == '/' {
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
