package config

// Config is the canonical ajolote configuration stored in .agents/config.json.
type Config struct {
	Project Project            `json:"project"`
	MCP     MCP                `json:"mcp"`
	Rules   Rules              `json:"rules"`
	Skills  []string           `json:"skills"`
	Tools   map[string]bool    `json:"tools"`
}

type Project struct {
	Name       string `json:"name"`
	Stack      string `json:"stack"`
	RepoType   string `json:"repoType"`
	Language   string `json:"language"`
	TestRunner string `json:"testRunner"`
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

type Rules struct {
	General  []string `json:"general"`
	CodeStyle []string `json:"codeStyle"`
	Testing  []string `json:"testing"`
	Security []string `json:"security"`
	Commits  []string `json:"commits"`
}

// AllTools is the ordered list of supported tool names.
var AllTools = []string{"claude", "cursor", "windsurf", "copilot", "cline", "aider"}

// EnabledTools returns tool names that are set to true.
func (c *Config) EnabledTools() []string {
	var enabled []string
	for _, name := range AllTools {
		if c.Tools[name] {
			enabled = append(enabled, name)
		}
	}
	return enabled
}
