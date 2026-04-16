package translators

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/localconfig"
)

// mcpServerJSON is the tool-facing MCP server representation.
// It omits ajolote-only fields (Scope) that must not appear in generated tool configs.
type mcpServerJSON struct {
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Description string            `json:"description,omitempty"`
	Transport   string            `json:"transport,omitempty"`
	URL         *string           `json:"url,omitempty"`
}

func toMCPJSON(srv config.MCPServer) mcpServerJSON {
	var env map[string]string
	if len(srv.Env) > 0 {
		env = make(map[string]string, len(srv.Env))
		for k, v := range srv.Env {
			env[k] = expandEnv(v)
		}
	}
	// URL: use a pointer so that an empty expansion of a non-empty placeholder
	// (e.g. ${EMPTY_VAR}="") still emits "url":"" rather than being omitted.
	// A truly absent URL (stdio server) is represented as nil and omitted.
	var url *string
	if srv.URL != "" {
		expanded := expandEnv(srv.URL)
		url = &expanded
	}
	return mcpServerJSON{
		Command:     srv.Command,
		Args:        srv.Args,
		Env:         env,
		Description: srv.Description,
		Transport:   srv.Transport,
		URL:         url,
	}
}

// expandEnv substitutes ${VAR} and $VAR references with their environment values.
// Variables that are not set in the environment are left as-is (not replaced with
// an empty string), so the placeholder is visible in generated output and the team
// knows a variable needs to be set.
func expandEnv(s string) string {
	return os.Expand(s, func(key string) string {
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return "${" + key + "}"
	})
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
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	// SEC-002: Refuse to follow symlinks on home-directory writes.
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write to symlink: %s", path)
	}
	// SEC-007: Home-directory files may contain resolved secrets; use 0o600.
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// writeFile writes content to path (relative to projectRoot), creating dirs as needed.
// If the developer has listed relPath in .agents/config.local.json "protect", the
// write is silently skipped — the file is left exactly as the developer left it.
func writeFile(projectRoot, relPath, content string) error {
	lc, _ := localconfig.Load(projectRoot)
	if lc.IsProtected(relPath) {
		return nil // protected — leave the file untouched
	}
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
// It also scans .agents/.base/commands/ for inherited commands and appends any
// that are not already present by name (local commands win on name conflict).
// Returns nil (no error) if neither directory exists.
func readCommands(projectRoot string) ([]Command, error) {
	localDir := filepath.Join(projectRoot, ".agents", "commands")
	baseDir := filepath.Join(projectRoot, ".agents", ".base", "commands")

	local, err := readCommandDir(localDir)
	if err != nil {
		return nil, err
	}
	base, err := readCommandDir(baseDir)
	if err != nil {
		return nil, err
	}

	// Merge: local first, then base entries not already in local.
	seen := make(map[string]bool, len(local))
	for _, c := range local {
		seen[c.Name] = true
	}
	commands := append([]Command{}, local...)
	for _, c := range base {
		if !seen[c.Name] {
			commands = append(commands, c)
		}
	}
	return commands, nil
}

// readCommandDir reads all *.md files from a single directory and parses them
// as commands. Returns nil (no error) if the directory does not exist.
func readCommandDir(dir string) ([]Command, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", dir, err)
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

// existingSkillNames returns a set of skill file names (without extension) already
// present in .agents/skills/.
func existingSkillNames(projectRoot string) (map[string]bool, error) {
	dir := filepath.Join(projectRoot, ".agents", "skills")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	names := make(map[string]bool, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names[strings.TrimSuffix(e.Name(), ".md")] = true
		}
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
// and are not ajolote-generated. Returns the ScopedRule metadata and a map of
// name → body content so callers can write the rule file with real content.
// skip is the same set used for command imports (e.g. "agents", "ajolote-sync").
func importScopedRulesFromMDC(dir string, skip map[string]bool) ([]config.ScopedRule, map[string]string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}
	var rules []config.ScopedRule
	contents := map[string]string{}
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
		fields, body := parseFrontmatterFull(string(data))
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
		contents[name] = body
	}
	return rules, contents
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

// agentName returns the filename stem from a persona path.
// ".agents/personas/code-reviewer.md" → "code-reviewer"
func agentName(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// toTitle converts a hyphen/underscore-separated slug to Title Case.
// "code-reviewer" → "Code Reviewer"
func toTitle(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
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

// parseWindsurfWorkflow extracts the description and body from a Windsurf
// workflow YAML file. It scans for a top-level "description:" key and collects
// the content of all "say: |" multi-line blocks. No full YAML parser is needed
// because the format is our own predictable generated output.
func parseWindsurfWorkflow(raw string) (description, body string) {
	lines := strings.Split(raw, "\n")
	var bodyLines []string
	inSay := false
	sayIndent := 0

	for _, line := range lines {
		if inSay {
			if line == "" {
				bodyLines = append(bodyLines, "")
				continue
			}
			indent := len(line) - len(strings.TrimLeft(line, " "))
			if indent >= sayIndent {
				bodyLines = append(bodyLines, line[sayIndent:])
			} else {
				inSay = false
			}
			continue
		}
		trimmed := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(trimmed, "description: "); ok {
			description = strings.TrimSpace(after)
			continue
		}
		if strings.HasSuffix(trimmed, "say: |") {
			inSay = true
			// Content lines are indented two spaces beyond the "say:" key.
			idx := strings.Index(line, "say:")
			sayIndent = idx + 2
		}
	}
	body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	return
}

// parseCodexTOML parses a .codex/config.toml file and returns the MCP servers
// it defines. The format is ajolote's own generated TOML so a full library is
// not required — a simple line-by-line state machine is sufficient.
// Returns nil map (no error) when the file does not exist.
func parseCodexTOML(path string) (map[string]config.MCPServer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	servers := map[string]config.MCPServer{}
	var current string // current server name
	inEnv := false

	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inner := line[1 : len(line)-1]
			const prefix = "mcp.servers."
			if strings.HasPrefix(inner, prefix) {
				rest := inner[len(prefix):]
				if strings.HasSuffix(rest, ".env") {
					current = rest[:len(rest)-4]
					inEnv = true
				} else if !strings.Contains(rest, ".") {
					current = rest
					inEnv = false
					if _, exists := servers[current]; !exists {
						servers[current] = config.MCPServer{}
					}
				}
			} else {
				current = ""
				inEnv = false
			}
			continue
		}

		if current == "" {
			continue
		}
		k, v, ok := strings.Cut(line, " = ")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)

		srv := servers[current]
		if inEnv {
			if srv.Env == nil {
				srv.Env = map[string]string{}
			}
			srv.Env[k] = strings.Trim(v, `"`)
		} else {
			switch k {
			case "command":
				srv.Command = strings.Trim(v, `"`)
			case "transport":
				srv.Transport = strings.Trim(v, `"`)
			case "url":
				srv.URL = strings.Trim(v, `"`)
			case "args":
				srv.Args = parseTomlStringArray(v)
			}
		}
		servers[current] = srv
	}
	return servers, nil
}

// parseTomlStringArray parses a TOML inline array literal like ["a", "-y", "b"]
// into a Go string slice.
func parseTomlStringArray(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var result []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
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
