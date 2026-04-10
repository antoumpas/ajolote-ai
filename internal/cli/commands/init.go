package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/gitignore"
	"github.com/ajolote-ai/ajolote/internal/translators"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize ajolote in the current project",
		Long:  "Interactive setup that creates .agents/config.json, seeds skill files, generates tool configs, and updates .gitignore.",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	if config.Exists(projectRoot) {
		overwrite := false
		if err := survey.AskOne(&survey.Confirm{
			Message: ".agents/config.json already exists. Overwrite?",
			Default: false,
		}, &overwrite); err != nil {
			return err
		}
		if !overwrite {
			fmt.Println("Aborted.")
			return nil
		}
	}

	cfg := config.DefaultConfig()

	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Project name:",
				Default: filepath.Base(projectRoot),
			},
		},
		{
			Name:   "language",
			Prompt: &survey.Input{Message: "Primary language:", Default: ""},
		},
		{
			Name:   "stack",
			Prompt: &survey.Input{Message: "Stack (e.g. Python / FastAPI / PostgreSQL):", Default: ""},
		},
		{
			Name:   "repoType",
			Prompt: &survey.Select{Message: "Repo type:", Options: []string{"polyrepo", "monorepo"}, Default: "polyrepo"},
		},
		{
			Name:   "testRunner",
			Prompt: &survey.Input{Message: "Test runner (e.g. pytest, Vitest, Jest):", Default: ""},
		},
	}

	answers := struct {
		Name       string
		Language   string
		Stack      string
		RepoType   string
		TestRunner string
	}{}

	if err := survey.Ask(questions, &answers); err != nil {
		return err
	}

	cfg.Project.Name = answers.Name
	cfg.Project.Language = answers.Language
	cfg.Project.Stack = answers.Stack
	cfg.Project.RepoType = answers.RepoType
	cfg.Project.TestRunner = answers.TestRunner

	// Tool selection
	toolOptions := []string{"claude", "cursor", "windsurf", "copilot", "cline", "aider"}
	var selectedTools []string
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Which AI tools does your team use?",
		Options: toolOptions,
	}, &selectedTools); err != nil {
		return err
	}

	for _, name := range toolOptions {
		cfg.Tools[name] = false
	}
	for _, name := range selectedTools {
		cfg.Tools[name] = true
	}

	// Save config
	if err := config.Save(projectRoot, cfg); err != nil {
		return err
	}
	printOK(".agents/config.json")

	// Seed skill files
	skillsDir := filepath.Join(projectRoot, ".agents", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return err
	}
	if err := seedSkill(filepath.Join(skillsDir, "git.md"), config.GitSkillContent); err != nil {
		return err
	}
	printOK(".agents/skills/git.md")

	if err := seedSkill(filepath.Join(skillsDir, "testing.md"), config.TestingSkillContent); err != nil {
		return err
	}
	printOK(".agents/skills/testing.md")

	// Generate tool-specific files
	if err := runSyncWith(projectRoot, cfg, ""); err != nil {
		return err
	}

	// Update .gitignore
	if err := runIgnoreWith(projectRoot, cfg); err != nil {
		return err
	}

	fmt.Println()
	color.Green("ajolote initialized successfully.")
	fmt.Println()
	fmt.Println("Commit these files:")
	fmt.Println("  .agents/config.json")
	fmt.Println("  .agents/skills/")
	fmt.Println("  AGENTS.md (if present)")
	fmt.Println()
	fmt.Println("Generated tool files are gitignored automatically.")
	return nil
}

func seedSkill(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists — don't overwrite human edits
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func runSyncWith(projectRoot string, cfg *config.Config, toolFilter string) error {
	for _, t := range translators.All() {
		if toolFilter != "" && t.Name() != toolFilter {
			continue
		}
		if !cfg.Tools[t.Name()] {
			continue
		}
		if err := t.Generate(cfg, projectRoot); err != nil {
			return fmt.Errorf("generating %s config: %w", t.Name(), err)
		}
		for _, f := range t.OutputFiles() {
			printOK(f)
		}
	}
	return nil
}

func runIgnoreWith(projectRoot string, cfg *config.Config) error {
	var entries []string
	for _, t := range translators.All() {
		if !cfg.Tools[t.Name()] {
			continue
		}
		entries = append(entries, t.OutputFiles()...)
	}
	if err := gitignore.Update(projectRoot, entries); err != nil {
		return err
	}
	printOK(".gitignore")
	return nil
}

func printOK(path string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("  %s %s\n", green("✔"), path)
}
