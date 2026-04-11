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

func SyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [<tool>]",
		Short: "Two-way sync between .agents/config.json and your tool's config",
		Long: `Runs in both directions:
  ↑ Reads the tool's existing config files and merges any new MCP servers into .agents/config.json
  ↓ Regenerates the tool's files from the (now updated) canonical config

Without a tool argument, syncs every tool whose config files are already present on disk.`,
		Args:    cobra.MaximumNArgs(1),
		RunE:    runSync,
		Example: "  ajolote sync\n  ajolote sync cursor\n  ajolote sync claude",
	}
}

func runSync(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	// Determine which tools to sync
	var targets []translators.Syncer
	if len(args) == 1 {
		t, err := translators.Get(args[0])
		if err != nil {
			return err
		}
		targets = []translators.Syncer{t}
	} else {
		// Sync all tools that have at least one output file present
		for _, t := range translators.All() {
			for _, f := range t.OutputFiles() {
				if _, err := os.Stat(filepath.Join(projectRoot, f)); err == nil {
					targets = append(targets, t)
					break
				}
			}
		}
		if len(targets) == 0 {
			fmt.Println("No tool configs found on disk. Run 'ajolote use <tool>' first.")
			return nil
		}
	}

	up := color.New(color.FgCyan).SprintFunc()
	down := color.New(color.FgGreen).SprintFunc()
	added := color.New(color.FgYellow).SprintFunc()

	configChanged := false

	for _, t := range targets {
		fmt.Printf("\n%s\n", color.New(color.Bold).Sprint(t.Name()))

		// ↑ Import
		result, err := t.Import(projectRoot)
		if err != nil {
			return fmt.Errorf("importing from %s: %w", t.Name(), err)
		}

		if result == nil {
			fmt.Printf("  %s no config files found — skipping import\n", up("↑"))
		} else if !result.HasChanges() {
			fmt.Printf("  %s nothing new to import\n", up("↑"))
		} else {
			if cfg.MCP.Servers == nil {
				cfg.MCP.Servers = map[string]config.MCPServer{}
			}
			for name, srv := range result.NewMCPServers {
				if _, exists := cfg.MCP.Servers[name]; !exists {
					cfg.MCP.Servers[name] = srv
					fmt.Printf("  %s %s  %s\n", up("↑"), name, added("(new MCP server added to config)"))
					configChanged = true
				}
			}
			for _, cmd := range result.NewCommands {
				if err := writeAgentsCommand(projectRoot, cmd); err != nil {
					return fmt.Errorf("writing .agents/commands/%s.md: %w", cmd.Name, err)
				}
				fmt.Printf("  %s .agents/commands/%s.md  %s\n", up("↑"), cmd.Name, added("(new command added)"))
				configChanged = true
			}
			for name, content := range result.NewRuleFiles {
				path := filepath.Join(projectRoot, ".agents", "rules", name)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
						return fmt.Errorf("creating .agents/rules/: %w", err)
					}
					if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
						return fmt.Errorf("writing .agents/rules/%s: %w", name, err)
					}
					fmt.Printf("  %s .agents/rules/%s  %s\n", up("↑"), name, added("(new rule file imported)"))
					configChanged = true
				}
			}
		}

		// ↓ Export
		if err := t.Generate(cfg, projectRoot); err != nil {
			return fmt.Errorf("generating %s config: %w", t.Name(), err)
		}
		for _, f := range t.OutputFiles() {
			fmt.Printf("  %s %s\n", down("↓"), f)
		}
	}

	fmt.Println()

	if configChanged {
		if err := config.Save(projectRoot, cfg); err != nil {
			return err
		}
		color.Yellow(".agents/ updated — review and commit when ready.")
	} else {
		color.Green("Done. Nothing new imported.")
	}

	return nil
}

// writeAgentsCommand writes a command to .agents/commands/<name>.md, creating the
// directory if needed and adding the path to .gitignore if it isn't already covered.
func writeAgentsCommand(projectRoot string, cmd translators.Command) error {
	dir := filepath.Join(projectRoot, ".agents", "commands")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var content string
	if cmd.Description != "" {
		content = "---\ndescription: " + cmd.Description + "\n---\n\n" + cmd.Content + "\n"
	} else {
		content = cmd.Content + "\n"
	}
	return os.WriteFile(filepath.Join(dir, cmd.Name+".md"), []byte(content), 0o644)
}
