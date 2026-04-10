package translators

import (
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
