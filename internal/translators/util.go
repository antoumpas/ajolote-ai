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

// rulesMarkdown renders the rules section of the config as a markdown bullet list.
func rulesMarkdown(cfg *config.Config) string {
	var sb strings.Builder

	sections := []struct {
		heading string
		items   []string
	}{
		{"General", cfg.Rules.General},
		{"Code Style", cfg.Rules.CodeStyle},
		{"Testing", cfg.Rules.Testing},
		{"Security", cfg.Rules.Security},
		{"Commits", cfg.Rules.Commits},
	}

	for _, s := range sections {
		if len(s.items) == 0 {
			continue
		}
		sb.WriteString("### " + s.heading + "\n\n")
		for _, item := range s.items {
			sb.WriteString("- " + item + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
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
