package commands

import (
	"fmt"
	"os"

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

	for _, f := range t.OutputFiles() {
		printOK(f)
	}

	fmt.Println()
	color.Green("Ready. %s is configured for this project.", toolName)
	return nil
}
