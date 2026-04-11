package config

// Config is the canonical ajolote configuration stored in .agents/config.json.
type Config struct {
	MCP         MCP          `json:"mcp"`
	Rules       []string     `json:"rules"`
	ScopedRules []ScopedRule `json:"scoped_rules,omitempty"`
	Skills      []string     `json:"skills"`
	Personas    []string     `json:"personas"`
	Context     []string     `json:"context"`
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
