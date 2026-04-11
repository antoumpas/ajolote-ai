package config

// Config is the canonical ajolote configuration stored in .agents/config.json.
type Config struct {
	MCP      MCP      `json:"mcp"`
	Rules    []string `json:"rules"`
	Skills   []string `json:"skills"`
	Personas []string `json:"personas"`
	Context  []string `json:"context"`
}

type MCP struct {
	Servers map[string]MCPServer `json:"servers"`
}

type MCPServer struct {
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env,omitempty"`
	Description string            `json:"description,omitempty"`
}
