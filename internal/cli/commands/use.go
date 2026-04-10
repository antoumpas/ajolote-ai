package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/translators"
)

func UseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <tool>",
		Short: "Generate local config files for your AI tool",
		Long: `Reads .agents/config.json and generates the config files for the specified tool.
Generated files are gitignored — they stay local to your machine.

Commands defined in .agents/commands/ are translated into the tool's native format.

Supported tools: claude, cursor, windsurf, copilot, cline, aider`,
		Args:    cobra.ExactArgs(1),
		RunE:    runUse,
		Example: "  ajolote use claude\n  ajolote use cursor",
	}
}

func runUse(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	t, err := translators.Get(toolName)
	if err != nil {
		return err
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	if err := t.Generate(cfg, projectRoot); err != nil {
		return fmt.Errorf("generating %s config: %w", toolName, err)
	}

	// Print generated files — walk directories for accurate listing
	for _, f := range t.OutputFiles() {
		if strings.HasSuffix(f, "/") {
			// Directory pattern — list what's actually inside
			dir := filepath.Join(projectRoot, f)
			entries, err := os.ReadDir(dir)
			if err == nil {
				for _, e := range entries {
					if !e.IsDir() {
						printOK(filepath.Join(f, e.Name()))
					}
				}
			}
		} else {
			printOK(f)
		}
	}

	fmt.Println()
	color.Green("Ready. %s is configured for this project.", toolName)
	return nil
}
