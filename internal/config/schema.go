package config

import "encoding/json"

// Config is the canonical ajolote configuration stored in .agents/config.json.
type Config struct {
	MCP         MCP          `json:"mcp"`
	Rules       []string     `json:"rules"`
	ScopedRules []ScopedRule `json:"scoped_rules,omitempty"`
	Skills      []string     `json:"skills"`
	Personas    []Persona    `json:"personas"`
	Context     []string     `json:"context"`
}

// Persona is a role-based behaviour file. It may optionally carry Claude-specific
// metadata to generate a proper .claude/agents/ subagent file.
type Persona struct {
	Path   string       `json:"path"`
	Claude *ClaudeAgent `json:"claude,omitempty"`
}

// UnmarshalJSON accepts both the legacy string form (".agents/personas/reviewer.md")
// and the new object form ({"path": "...", "claude": {...}}) so existing configs
// continue to parse without modification.
func (p *Persona) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		p.Path = s
		return nil
	}
	type alias Persona
	return json.Unmarshal(data, (*alias)(p))
}

// ClaudeAgent holds Claude Code-specific subagent metadata for a persona.
type ClaudeAgent struct {
	Model       string   `json:"model,omitempty"`       // "haiku" | "sonnet" | "opus" | full model ID
	Tools       []string `json:"tools,omitempty"`       // Claude Code tool names (e.g. "Read", "Grep", "Glob")
	Description string   `json:"description,omitempty"` // auto-invocation trigger; derived from file if omitted
}

// ScopedRule is a rule that only applies to files matching specific glob patterns.
// Each AI tool renders it in its native scoped-rule format.
type ScopedRule struct {
	Name  string   `json:"name"`
	Globs []string `json:"globs"`
	Path  string   `json:"path"`
}

type MCP struct {
	Servers map[string]MCPServer `json:"servers"`
}

type MCPServer struct {
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Description string            `json:"description,omitempty"`
	Transport   string            `json:"transport,omitempty"` // "stdio" (default) | "http" | "sse"
	URL         string            `json:"url,omitempty"`       // for http/sse transport
	Scope       string            `json:"scope,omitempty"`     // "project" (default) | "user"
}
