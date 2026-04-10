package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/translators"
)

func AddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <tool>",
		Short: "Enable a tool and generate its config",
		Long:  "Adds a tool to .agents/config.json, generates its config files, and updates .gitignore.",
		Args:  cobra.ExactArgs(1),
		RunE:  runAdd,
	}
}

func runAdd(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	// Validate tool name
	if _, err := translators.Get(toolName); err != nil {
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

	if cfg.Tools[toolName] {
		fmt.Printf("%s is already enabled.\n", toolName)
		return nil
	}

	cfg.Tools[toolName] = true

	if err := config.Save(projectRoot, cfg); err != nil {
		return err
	}
	printOK(".agents/config.json")

	if err := runSyncWith(projectRoot, cfg, toolName); err != nil {
		return err
	}

	if err := runIgnoreWith(projectRoot, cfg); err != nil {
		return err
	}

	fmt.Println()
	color.Green("Added %s.", toolName)
	return nil
}
