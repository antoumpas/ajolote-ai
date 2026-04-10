package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
)

func SyncCmd() *cobra.Command {
	var toolFilter string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Generate tool-specific config files from .agents/config.json",
		Long:  "Runs all enabled translators (or a single one with --tool) and writes tool-specific files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(toolFilter)
		},
	}

	cmd.Flags().StringVar(&toolFilter, "tool", "", "Only sync a specific tool (e.g. claude, cursor)")
	return cmd
}

func runSync(toolFilter string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	count := 0
	if err := runSyncWith(projectRoot, cfg, toolFilter); err != nil {
		return err
	}

	_ = count
	fmt.Println()
	color.Green("Sync complete.")
	return nil
}
