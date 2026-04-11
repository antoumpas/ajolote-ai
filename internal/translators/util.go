package translators

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ajolote-ai/ajolote/internal/config"
)

// mcpServerJSON is the tool-facing MCP server representation.
// It omits ajolote-only fields (Scope) that must not appear in generated tool configs.
type mcpServerJSON struct {
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Description string            `json:"description,omitempty"`
	Transport   string            `json:"transport,omitempty"`
	URL         string            `json:"url,omitempty"`
}

func toMCPJSON(srv config.MCPServer) mcpServerJSON {
	return mcpServerJSON{
		Command:     srv.Command,
		Args:        srv.Args,
		Env:         srv.Env,
		Description: srv.Description,
		Transport:   srv.Transport,
		URL:         srv.URL,
	}
}

// projectScopedServers returns servers whose scope is "" or "project".
func projectScopedServers(servers map[string]config.MCPServer) map[string]config.MCPServer {
	out := map[string]config.MCPServer{}
	for name, srv := range servers {
		if srv.Scope == "" || srv.Scope == "project" {
			out[name] = srv
		}
	}
	return out
}

// userScopedServers returns servers whose scope is "user".
func userScopedServers(servers map[string]config.MCPServer) map[string]config.MCPServer {
	out := map[string]config.MCPServer{}
	for name, srv := range servers {
		if srv.Scope == "user" {
			out[name] = srv
		}
	}
	return out
}

// mergeUserMCPConfig reads an existing {"mcpServers":{...}} JSON file at path,
// adds any servers not already present, and writes back only if changed.
// Creates the file (and parent dirs) if it doesn't exist. Never removes existing servers.
func mergeUserMCPConfig(path string, servers map[string]config.MCPServer) error {
	if len(servers) == 0 {
		return nil
	}

	existing := map[string]json.RawMessage{}
	if data, err := os.ReadFile(path); err == nil {
		var wrapper struct {
			MCPServers map[string]json.RawMessage `json:"mcpServers"`
		}
		if json.Unmarshal(data, &wrapper) == nil && wrapper.MCPServers != nil {
			existing = wrapper.MCPServers
		}
	}

	changed := false
	for name, srv := range servers {
		if _, ok := existing[name]; ok {
			continue // already present — don't overwrite
		}
		b, err := json.Marshal(toMCPJSON(srv))
		if err != nil {
			return err
		}
		existing[name] = json.RawMessage(b)
		changed = true
	}
	if !changed {
		return nil
	}

	type out struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	data, err := json.MarshalIndent(out{MCPServers: existing}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// writeFile writes content to path (relative to projectRoot), creating dirs as needed.
func writeFile(projectRoot, relPath, content string) error {
	full := filepath.Join(projectRoot, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", relPath, err)
	}
	return os.WriteFile(full, []byte(content), 0o644)
}

// parseMCPFile reads a {"mcpServers": {...}} JSON file and returns the servers map.
// Returns nil map (no error) when the file does not exist.
func parseMCPFile(path string) (map[string]config.MCPServer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var wrapper struct {
		MCPServers map[string]config.MCPServer `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return wrapper.MCPServers, nil
}

// Command is a team-defined agent command loaded from .agents/commands/*.md.
type Command struct {
	Name        string
	Description string
	Content     string
}

// readCommands reads all *.md files from .agents/commands/ under projectRoot.
// Returns nil (no error) if the directory does not exist.
func readCommands(projectRoot string) ([]Command, error) {
	dir := filepath.Join(projectRoot, ".agents", "commands")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading .agents/commands: %w", err)
	}

	var commands []Command
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		desc, content := parseFrontmatter(string(data))
		commands = append(commands, Command{Name: name, Description: desc, Content: content})
	}
	return commands, nil
}

// parseFrontmatterFull parses a --- YAML frontmatter block and returns all key→value
// pairs plus the body content after the closing ---. Returns an empty map and the
// full raw string as body when no frontmatter is present.
func parseFrontmatterFull(raw string) (fields map[string]string, body string) {
	fields = map[string]string{}
	if !strings.HasPrefix(raw, "---\n") {
		return fields, strings.TrimSpace(raw)
	}
	rest := raw[4:]
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return fields, strings.TrimSpace(raw)
	}
	for _, line := range strings.Split(rest[:end], "\n") {
		if k, v, ok := strings.Cut(line, ":"); ok {
			fields[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"`)
		}
	}
	return fields, strings.TrimSpace(rest[end+5:])
}

// parseFrontmatter splits a markdown file into its description (from --- frontmatter)
// and body content. If no frontmatter is present, the whole string is the body.
func parseFrontmatter(raw string) (description, body string) {
	if !strings.HasPrefix(raw, "---\n") {
		return "", strings.TrimSpace(raw)
	}
	rest := raw[4:]
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return "", strings.TrimSpace(raw)
	}
	for _, line := range strings.Split(rest[:end], "\n") {
		if after, ok := strings.CutPrefix(line, "description:"); ok {
			description = strings.TrimSpace(after)
		}
	}
	return description, strings.TrimSpace(rest[end+5:])
}

// importCommandFiles scans dir for files with the given extension, skips names in skip or existing,
// and uses parse to convert each file into a Command.
func importCommandFiles(dir, ext string, skip, existing map[string]bool, parse func(name, raw string) Command) ([]Command, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	var cmds []Command
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ext)
		if skip[name] || existing[name] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, parse(name, string(data)))
	}
	return cmds, nil
}

// existingCommandNames returns a set of command names already in .agents/commands/.
func existingCommandNames(projectRoot string) (map[string]bool, error) {
	cmds, err := readCommands(projectRoot)
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool, len(cmds))
	for _, c := range cmds {
		names[c.Name] = true
	}
	return names, nil
}

// atFileList renders paths as Claude Code @file imports under a heading.
func atFileList(heading string, paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## " + heading + "\n\n")
	for _, p := range paths {
		sb.WriteString("@" + p + "\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

// inlineFiles reads each file in paths and embeds its content under heading.
// If a file cannot be read it falls back to a comment so generation never fails.
func inlineFiles(heading, projectRoot string, paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## " + heading + "\n\n")
	for _, p := range paths {
		data, err := os.ReadFile(filepath.Join(projectRoot, p))
		if err != nil {
			sb.WriteString(fmt.Sprintf("<!-- %s -->\n\n", p))
			continue
		}
		sb.WriteString(fmt.Sprintf("<!-- %s -->\n\n", p))
		sb.WriteString(strings.TrimSpace(string(data)))
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// importScopedRulesFromMDC scans dir for .mdc files that have a globs: frontmatter field
// and are not ajolote-generated. Returns them as ScopedRule values with the body as content.
// skip is the same set used for command imports (e.g. "agents", "ajolote-sync").
func importScopedRulesFromMDC(dir string, skip map[string]bool) []config.ScopedRule {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var rules []config.ScopedRule
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".mdc") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".mdc")
		if skip[name] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil || isAjoloteGenerated(string(data)) {
			continue
		}
		fields, _ := parseFrontmatterFull(string(data))
		globStr := fields["globs"]
		if globStr == "" {
			continue // not a scoped rule
		}
		var globs []string
		for _, g := range strings.Split(globStr, ",") {
			if g = strings.TrimSpace(g); g != "" {
				globs = append(globs, g)
			}
		}
		rules = append(rules, config.ScopedRule{
			Name:  name,
			Globs: globs,
			Path:  ".agents/rules/" + name + ".md",
		})
	}
	return rules
}

// isAjoloteGenerated reports whether content was produced by ajolote (not user-authored).
// Files generated by ajolote contain the standard "Generated by ajolote-ai" marker.
func isAjoloteGenerated(content string) bool {
	return strings.Contains(content, "Generated by ajolote-ai")
}

// stripFrontmatter removes a leading --- YAML block (if present) and returns the body.
func stripFrontmatter(raw string) string {
	if !strings.HasPrefix(raw, "---\n") {
		return raw
	}
	rest := raw[4:]
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return raw
	}
	return strings.TrimSpace(rest[end+5:])
}

// personaPaths extracts the Path field from each Persona, returning a plain []string
// of file paths suitable for use with atFileList, inlineFiles, and similar helpers.
func personaPaths(personas []config.Persona) []string {
	paths := make([]string, len(personas))
	for i, p := range personas {
		paths[i] = p.Path
	}
	return paths
}

// claudeModelAliases maps shorthand names to full Claude model IDs.
var claudeModelAliases = map[string]string{
	"haiku":  "claude-haiku-4-5-20251001",
	"sonnet": "claude-sonnet-4-6",
	"opus":   "claude-opus-4-6",
}

// resolveModel maps a model shorthand to its full ID, or returns the input unchanged
// if it is already a full model ID or unknown.
func resolveModel(model string) string {
	if full, ok := claudeModelAliases[model]; ok {
		return full
	}
	return model
}

// fileListMarkdown renders a list of file paths as markdown bullets.
func fileListMarkdown(heading string, paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## " + heading + "\n\n")
	for _, p := range paths {
		sb.WriteString(fmt.Sprintf("- `%s`\n", p))
	}
	sb.WriteString("\n")
	return sb.String()
}
