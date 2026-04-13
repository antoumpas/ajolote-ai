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
		Short: "Show which tool configs have been generated locally",
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

	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	_ = cfg // loaded for validation; project name no longer in config
	fmt.Printf("\n")
	fmt.Println(bold("Tool configs (local, gitignored):"))
	fmt.Println()

	for _, t := range translators.All() {
		anyPresent := false
		for _, f := range t.OutputFiles() {
			if _, err := os.Stat(filepath.Join(projectRoot, f)); err == nil {
				anyPresent = true
				break
			}
		}

		if anyPresent {
			fmt.Printf("  %s %-12s\n", green("✔"), t.Name())
		} else {
			fmt.Printf("  %s %-12s  %s\n", yellow("○"), t.Name(), yellow("run: ajolote use "+t.Name()))
		}
	}

	fmt.Println()
	return nil
}
