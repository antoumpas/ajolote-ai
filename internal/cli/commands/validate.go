package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
)

func ValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Check that all files and MCP servers in config.json are valid",
		Long: `Validate checks the canonical config for common problems before syncing:

  - All rule, skill, persona, context, and scoped-rule files exist and are non-empty
  - Scoped rules define at least one glob pattern
  - MCP stdio servers have a resolvable command in PATH
  - MCP http/sse servers have a URL configured

Exits with status 1 if any error is found. Warnings (e.g. a command not found
in the current PATH) are printed but do not affect the exit code.`,
		SilenceUsage: true,
		RunE:         runValidate,
	}
}

func runValidate(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	v := &validator{projectRoot: projectRoot}
	v.run(cfg)

	fmt.Println()
	if v.errors > 0 {
		msg := fmt.Sprintf("%d error(s)", v.errors)
		if v.warnings > 0 {
			msg += fmt.Sprintf(", %d warning(s)", v.warnings)
		}
		color.Red(msg)
		return fmt.Errorf("validation failed — fix the errors above before running ajolote sync")
	}
	if v.warnings > 0 {
		color.Yellow("%d warning(s) — review above", v.warnings)
		return nil
	}
	color.Green("All checks passed.")
	return nil
}

type validator struct {
	projectRoot string
	errors      int
	warnings    int
}

func (v *validator) run(cfg *config.Config) {
	v.checkFileSection("Rules", cfg.Rules)
	v.checkFileSection("Skills", cfg.Skills)
	v.checkPersonas(cfg.Personas)
	v.checkFileSection("Context", cfg.Context)
	v.checkScopedRules(cfg.ScopedRules)
	v.checkMCPServers(cfg.MCP.Servers)
}

func (v *validator) checkFileSection(heading string, paths []string) {
	if len(paths) == 0 {
		return
	}
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("\n%s\n", bold(heading))
	for _, p := range paths {
		v.checkFile(p)
	}
}

func (v *validator) checkPersonas(personas []config.Persona) {
	if len(personas) == 0 {
		return
	}
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("\n%s\n", bold("Personas"))
	for _, p := range personas {
		v.checkFile(p.Path)
	}
}

func (v *validator) checkScopedRules(rules []config.ScopedRule) {
	if len(rules) == 0 {
		return
	}
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("\n%s\n", bold("Scoped Rules"))
	for _, sr := range rules {
		if len(sr.Globs) == 0 {
			v.fail(sr.Name, "no glob patterns defined")
			continue
		}
		v.checkFile(sr.Path)
	}
}

func (v *validator) checkMCPServers(servers map[string]config.MCPServer) {
	if len(servers) == 0 {
		return
	}
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("\n%s\n", bold("MCP Servers"))

	// Sort for deterministic output
	names := make([]string, 0, len(servers))
	for name := range servers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		v.checkMCPServer(name, servers[name])
	}
}

func (v *validator) checkMCPServer(name string, srv config.MCPServer) {
	transport := srv.Transport
	if transport == "" {
		transport = "stdio"
	}

	switch transport {
	case "http", "sse":
		if srv.URL == "" {
			v.fail(name, fmt.Sprintf("transport %q requires a url", transport))
		} else {
			v.ok(fmt.Sprintf("%s (%s)", name, transport))
		}
	default: // stdio
		if srv.Command == "" {
			v.fail(name, "stdio server requires a command")
			return
		}
		if _, err := exec.LookPath(srv.Command); err != nil {
			v.warn(name, fmt.Sprintf("command not in PATH: %s", srv.Command))
		} else {
			v.ok(fmt.Sprintf("%s (%s)", name, srv.Command))
		}
	}
}

// checkFile verifies a project-relative path exists and is non-empty.
func (v *validator) checkFile(relPath string) {
	full := filepath.Join(v.projectRoot, relPath)
	info, err := os.Stat(full)
	if err != nil {
		v.fail(relPath, "file not found")
		return
	}
	if info.Size() == 0 {
		v.fail(relPath, "file is empty")
		return
	}
	data, err := os.ReadFile(full)
	if err != nil {
		v.fail(relPath, fmt.Sprintf("cannot read: %v", err))
		return
	}
	if strings.TrimSpace(string(data)) == "" {
		v.fail(relPath, "file contains only whitespace")
		return
	}
	v.ok(relPath)
}

func (v *validator) ok(label string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("  %s %s\n", green("✔"), label)
}

func (v *validator) warn(label, msg string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("  %s %s — %s\n", yellow("⚠"), label, msg)
	v.warnings++
}

func (v *validator) fail(label, msg string) {
	red := color.New(color.FgRed).SprintFunc()
	fmt.Printf("  %s %s — %s\n", red("✘"), label, msg)
	v.errors++
}
