package translators

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ajolote-ai/ajolote/internal/config"
)

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
