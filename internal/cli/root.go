package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/cli/commands"
)

var rootCmd = &cobra.Command{
	Use:   "ajolote",
	Short: "Shared AI agent configuration for multi-tool teams",
	Long: `ajolote manages a canonical AI agent configuration (.agents/config.json)
and translates it into each developer's preferred AI tool format.

Tool-specific files are gitignored; only the canonical config is committed.`,
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(commands.InitCmd())
	rootCmd.AddCommand(commands.SyncCmd())
	rootCmd.AddCommand(commands.IgnoreCmd())
	rootCmd.AddCommand(commands.AddCmd())
	rootCmd.AddCommand(commands.StatusCmd())
}
