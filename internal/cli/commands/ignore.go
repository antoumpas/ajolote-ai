package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
)

func IgnoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ignore",
		Short: "Update the ajolote-ai block in .gitignore",
		Long:  "Rewrites the managed <ajolote-ai> block in .gitignore based on currently enabled tools.",
		RunE:  runIgnore,
	}
}

func runIgnore(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	if err := runIgnoreWith(projectRoot, cfg); err != nil {
		return err
	}

	fmt.Println()
	color.Green(".gitignore updated.")
	return nil
}
