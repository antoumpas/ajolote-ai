package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/translators"
)

func StatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show configuration status for each tool",
		Long:  "Reports which tools are enabled, which generated files exist, and whether they are present.",
		RunE:  runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Printf("\n%s: %s\n", bold("Project"), cfg.Project.Name)
	fmt.Printf("%s: %s\n\n", bold("Config"), config.ConfigPath)

	fmt.Println(bold("Tool Status:"))
	fmt.Println()

	for _, t := range translators.All() {
		enabled := cfg.Tools[t.Name()]

		statusLabel := red("disabled")
		if enabled {
			statusLabel = green("enabled")
		}

		fmt.Printf("  %-12s %s\n", t.Name(), statusLabel)

		if enabled {
			for _, f := range t.OutputFiles() {
				full := filepath.Join(projectRoot, f)
				if _, err := os.Stat(full); err == nil {
					fmt.Printf("    %s %s\n", green("✔"), f)
				} else {
					fmt.Printf("    %s %s %s\n", yellow("!"), f, yellow("(not generated — run `ajolote sync`)"))
				}
			}
		}
	}

	fmt.Println()
	return nil
}
