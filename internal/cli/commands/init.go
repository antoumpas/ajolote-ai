package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/gitignore"
	"github.com/ajolote-ai/ajolote/internal/translators"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Scaffold ajolote config in the current project",
		Long: `Creates .agents/config.json (the shared source of truth), seeds starter skill files,
and adds all AI tool files to .gitignore.

If any supported AI tool is already configured in this project, its MCP servers
and command files are imported into .agents/ automatically.

Edit .agents/config.json to match your project, then commit the .agents/ directory.
Each developer then runs 'ajolote use <tool>' to generate their own local tool config.`,
		RunE: runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	if config.Exists(projectRoot) {
		fmt.Println(".agents/config.json already exists.")
		fmt.Println("Edit it directly, or delete it and re-run 'ajolote init' to start fresh.")
		return nil
	}

	// Create config scaffold
	cfg := config.DefaultConfig(filepath.Base(projectRoot))

	// Detect any existing tool configs and import from them before saving
	imported := importFromExistingTools(projectRoot, cfg)

	if err := config.Save(projectRoot, cfg); err != nil {
		return err
	}
	printOK(".agents/config.json")
	printImportSummary(imported)

	// Seed skill files
	skillsDir := filepath.Join(projectRoot, ".agents", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return err
	}
	if err := seedFile(filepath.Join(skillsDir, "git.md"), config.GitSkillContent); err != nil {
		return err
	}
	printOK(".agents/skills/git.md")
	if err := seedFile(filepath.Join(skillsDir, "testing.md"), config.TestingSkillContent); err != nil {
		return err
	}
	printOK(".agents/skills/testing.md")

	// Seed persona files
	personasDir := filepath.Join(projectRoot, ".agents", "personas")
	if err := os.MkdirAll(personasDir, 0o755); err != nil {
		return err
	}
	if err := seedFile(filepath.Join(personasDir, "reviewer.md"), config.ReviewerPersonaContent); err != nil {
		return err
	}
	printOK(".agents/personas/reviewer.md")
	if err := seedFile(filepath.Join(personasDir, "architect.md"), config.ArchitectPersonaContent); err != nil {
		return err
	}
	printOK(".agents/personas/architect.md")

	// Seed context files
	contextDir := filepath.Join(projectRoot, ".agents", "context")
	if err := os.MkdirAll(contextDir, 0o755); err != nil {
		return err
	}
	if err := seedFile(filepath.Join(contextDir, "architecture.md"), config.ArchitectureContextContent); err != nil {
		return err
	}
	printOK(".agents/context/architecture.md")
	if err := seedFile(filepath.Join(contextDir, "data-model.md"), config.DataModelContextContent); err != nil {
		return err
	}
	printOK(".agents/context/data-model.md")
	if err := seedFile(filepath.Join(contextDir, "glossary.md"), config.GlossaryContextContent); err != nil {
		return err
	}
	printOK(".agents/context/glossary.md")

	// Seed commands — starter review.md plus any imported from existing tools
	commandsDir := filepath.Join(projectRoot, ".agents", "commands")
	if err := os.MkdirAll(commandsDir, 0o755); err != nil {
		return err
	}
	if err := seedFile(filepath.Join(commandsDir, "review.md"), config.ReviewCommandContent); err != nil {
		return err
	}
	printOK(".agents/commands/review.md")
	for _, c := range imported.commands {
		if err := writeAgentsCommand(projectRoot, c); err != nil {
			return fmt.Errorf("writing imported command %s: %w", c.Name, err)
		}
		printOK(".agents/commands/" + c.Name + ".md")
	}

	// Gitignore all tool output files upfront
	if err := ignoreAllTools(projectRoot); err != nil {
		return err
	}
	printOK(".gitignore")

	fmt.Println()
	color.Green("ajolote initialized.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit .agents/config.json with your project details and rules")
	fmt.Println("  2. Commit the .agents/ directory")
	fmt.Println("  3. Each developer runs: ajolote use <tool>")
	fmt.Printf("     Supported tools: %s\n", translators.Names())
	return nil
}

// toolImport holds what was found for a single tool during init.
type toolImport struct {
	name     string
	nServers int
	commands []translators.Command
}

// initImports is the aggregated result of scanning all tools.
type initImports struct {
	byTool   []toolImport
	commands []translators.Command // deduplicated across tools
}

// importFromExistingTools scans all translators for existing configs and merges
// discovered MCP servers into cfg. Returns a summary of what was found.
func importFromExistingTools(projectRoot string, cfg *config.Config) initImports {
	var result initImports
	seen := map[string]bool{} // deduplicate commands across tools

	for _, t := range translators.All() {
		ir, err := t.Import(projectRoot)
		if err != nil || ir == nil {
			continue
		}

		ti := toolImport{name: t.Name()}

		// Merge MCP servers
		if cfg.MCP.Servers == nil {
			cfg.MCP.Servers = map[string]config.MCPServer{}
		}
		for name, srv := range ir.NewMCPServers {
			if _, exists := cfg.MCP.Servers[name]; !exists {
				cfg.MCP.Servers[name] = srv
				ti.nServers++
			}
		}

		// Collect commands (first tool to define a name wins)
		for _, c := range ir.NewCommands {
			if !seen[c.Name] {
				seen[c.Name] = true
				ti.commands = append(ti.commands, c)
				result.commands = append(result.commands, c)
			}
		}

		if ti.nServers > 0 || len(ti.commands) > 0 {
			result.byTool = append(result.byTool, ti)
		}
	}

	return result
}

func printImportSummary(imported initImports) {
	if len(imported.byTool) == 0 {
		return
	}
	up := color.New(color.FgCyan).SprintFunc()
	fmt.Println()
	fmt.Println("  Detected existing tool configs:")
	for _, ti := range imported.byTool {
		parts := ""
		if ti.nServers == 1 {
			parts += "1 MCP server"
		} else if ti.nServers > 1 {
			parts += fmt.Sprintf("%d MCP servers", ti.nServers)
		}
		if len(ti.commands) > 0 {
			if parts != "" {
				parts += ", "
			}
			if len(ti.commands) == 1 {
				parts += "1 command"
			} else {
				parts += fmt.Sprintf("%d commands", len(ti.commands))
			}
		}
		fmt.Printf("    %s %s — %s imported\n", up("↑"), ti.name, parts)
	}
	fmt.Println()
}

func seedFile(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists — don't overwrite
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func ignoreAllTools(projectRoot string) error {
	var entries []string
	for _, t := range translators.All() {
		entries = append(entries, t.OutputFiles()...)
	}
	return gitignore.Update(projectRoot, entries)
}

func printOK(path string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("  %s %s\n", green("✔"), path)
}
