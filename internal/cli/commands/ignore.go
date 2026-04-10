package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func IgnoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ignore",
		Short: "Refresh the ajolote-ai block in .gitignore",
		Long:  "Rewrites the managed <ajolote-ai> block in .gitignore to cover all supported tool files.",
		RunE:  runIgnore,
	}
}

func runIgnore(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := ignoreAllTools(projectRoot); err != nil {
		return err
	}

	printOK(".gitignore")
	fmt.Println()
	color.Green(".gitignore updated.")
	return nil
}
