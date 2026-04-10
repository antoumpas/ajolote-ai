package translators

import "github.com/ajolote-ai/ajolote/internal/config"

// ImportResult describes what was found in the tool's existing config files
// that is not yet present in .agents/config.json.
type ImportResult struct {
	// NewMCPServers are servers found in the tool's config but absent from config.mcp.servers.
	NewMCPServers map[string]config.MCPServer
}

// HasChanges reports whether the import found anything new.
func (r *ImportResult) HasChanges() bool {
	return len(r.NewMCPServers) > 0
}

// Syncer extends Translator with the ability to read a tool's existing config
// and return anything not yet present in the canonical config.
type Syncer interface {
	Translator
	// Import reads the tool's existing files under projectRoot and returns
	// whatever it found that is not already represented in the canonical config.
	// Returns a non-nil result with no changes (and nil error) when files exist
	// but nothing new was found. Returns nil result when no tool files exist at all.
	Import(projectRoot string) (*ImportResult, error)
}
