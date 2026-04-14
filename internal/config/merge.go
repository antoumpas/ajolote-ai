package config

import (
	"path/filepath"
	"strings"
)

const (
	agentsPrefix = ".agents/"
	basePrefix   = ".agents/.base/"
)

// rebaseConfigPaths rewrites all file paths in cfg that start with ".agents/"
// to start with ".agents/.base/" instead. This makes the base config's file
// references valid when resolved inside the inheriting project's directory.
func RebaseConfigPaths(cfg *Config) *Config {
	out := *cfg // shallow copy of the struct value
	out.Rules = rebasePaths(cfg.Rules)
	out.Skills = rebasePaths(cfg.Skills)
	out.Context = rebasePaths(cfg.Context)
	out.Commands = rebasePaths(cfg.Commands)

	out.Personas = make([]Persona, len(cfg.Personas))
	for i, p := range cfg.Personas {
		out.Personas[i] = p
		out.Personas[i].Path = rebasePath(p.Path)
	}

	out.ScopedRules = make([]ScopedRule, len(cfg.ScopedRules))
	for i, sr := range cfg.ScopedRules {
		out.ScopedRules[i] = sr
		out.ScopedRules[i].Path = rebasePath(sr.Path)
	}

	return &out
}

func rebasePath(p string) string {
	if strings.HasPrefix(p, agentsPrefix) {
		return basePrefix + p[len(agentsPrefix):]
	}
	return p
}

func rebasePaths(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = rebasePath(p)
	}
	return out
}

// mergeConfigs returns a new Config where base provides defaults and local
// values always win on conflict. The Extends field is always cleared in the
// returned config (inheritance does not chain).
func MergeConfigs(base, local *Config) *Config {
	merged := &Config{
		MCP: MCP{Servers: make(map[string]MCPServer)},
	}

	// MCP servers: base provides defaults, local overrides by server name.
	for k, v := range base.MCP.Servers {
		merged.MCP.Servers[k] = v
	}
	for k, v := range local.MCP.Servers {
		merged.MCP.Servers[k] = v
	}

	// File-path lists: local entries first, then base entries whose filename
	// is not already represented in local (local wins by filename).
	merged.Rules = mergeFilePaths(base.Rules, local.Rules)
	merged.Skills = mergeFilePaths(base.Skills, local.Skills)
	merged.Context = mergeFilePaths(base.Context, local.Context)
	merged.Commands = mergeFilePaths(base.Commands, local.Commands)

	// Personas: deduplicated by the base filename of the .Path field.
	merged.Personas = mergePersonas(base.Personas, local.Personas)

	// ScopedRules: deduplicated by Name (local wins).
	merged.ScopedRules = mergeScopedRules(base.ScopedRules, local.ScopedRules)

	// Extends is intentionally NOT inherited — no chaining.
	merged.Extends = ""

	return merged
}

// mergeFilePaths returns local paths followed by any base paths whose
// filepath.Base is not already present in local. This preserves local
// ordering while appending inherited additions.
func mergeFilePaths(base, local []string) []string {
	result := append([]string{}, local...)
	localNames := make(map[string]bool, len(local))
	for _, p := range local {
		localNames[filepath.Base(p)] = true
	}
	for _, p := range base {
		if !localNames[filepath.Base(p)] {
			result = append(result, p)
		}
	}
	return result
}

// mergePersonas merges base and local personas, deduplicating by the base
// filename of each persona's Path. Local entries take precedence.
func mergePersonas(base, local []Persona) []Persona {
	result := append([]Persona{}, local...)
	localNames := make(map[string]bool, len(local))
	for _, p := range local {
		localNames[filepath.Base(p.Path)] = true
	}
	for _, p := range base {
		if !localNames[filepath.Base(p.Path)] {
			result = append(result, p)
		}
	}
	return result
}

// mergeScopedRules merges base and local scoped rules, deduplicating by Name.
// Local entries take precedence.
func mergeScopedRules(base, local []ScopedRule) []ScopedRule {
	result := append([]ScopedRule{}, local...)
	localNames := make(map[string]bool, len(local))
	for _, sr := range local {
		localNames[sr.Name] = true
	}
	for _, sr := range base {
		if !localNames[sr.Name] {
			result = append(result, sr)
		}
	}
	return result
}
