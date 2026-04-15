package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/scanning"
)

// ScanCmd returns the cobra command for `ajolote scan`.
func ScanCmd() *cobra.Command {
	var format string
	var failOnWarn bool

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Detect leaked secrets and prompt-injection payloads in .agents/ files",
		Long: `Scan inspects the content of every .md file under .agents/rules/ and
.agents/skills/ (plus all files referenced in config.json) for two classes
of security risk:

  secrets    — API keys, tokens, private keys, database URIs, etc.
  injection  — Prompt-injection payloads: "ignore previous instructions",
               persona hijacks, control-character tricks, special tokens

Exit codes:
  0  — no findings
  1  — one or more errors (secrets or injections)
  2  — warnings only (reserved for future use)

Use --fail-on-warn to treat warnings as errors (exit 1).`,
		SilenceUsage: true,
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	cmd.Flags().BoolVar(&failOnWarn, "fail-on-warn", false, "Exit 1 even when only warnings are found")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runScan(format, failOnWarn)
	}
	return cmd
}

func runScan(format string, failOnWarn bool) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	d := scanning.NewDetector()

	// 1. Scan all .md files under .agents/rules/, .agents/skills/, etc.
	agentsDir := filepath.Join(projectRoot, ".agents")
	dirFindings, err := d.ScanDir(agentsDir)
	if err != nil {
		return fmt.Errorf("scan dir: %w", err)
	}

	// 2. Also scan config-referenced files outside .agents/ (e.g. custom paths).
	var referencedPaths []string
	referencedPaths = append(referencedPaths, cfg.Rules...)
	referencedPaths = append(referencedPaths, cfg.Skills...)
	referencedPaths = append(referencedPaths, cfg.Context...)
	for _, p := range cfg.Personas {
		referencedPaths = append(referencedPaths, p.Path)
	}
	for _, sr := range cfg.ScopedRules {
		referencedPaths = append(referencedPaths, sr.Path)
	}

	// Deduplicate: mark files already scanned by ScanDir so we don't double-report.
	scannedByDir := map[string]bool{}
	for _, f := range dirFindings {
		scannedByDir[f.File] = true
	}
	var extraPaths []string
	for _, rel := range referencedPaths {
		if !scannedByDir[rel] {
			extraPaths = append(extraPaths, rel)
		}
	}

	extraFindings, err := d.ScanFiles(projectRoot, extraPaths)
	if err != nil {
		return fmt.Errorf("scan files: %w", err)
	}

	all := append(dirFindings, extraFindings...)

	// Sort findings for deterministic output: file, then line.
	sort.Slice(all, func(i, j int) bool {
		if all[i].File != all[j].File {
			return all[i].File < all[j].File
		}
		return all[i].Line < all[j].Line
	})

	switch format {
	case "json":
		return outputJSON(all)
	default:
		return outputText(all, failOnWarn)
	}
}

// outputText prints human-readable findings to stdout and returns an error
// (to set the exit code) when errors are present.
func outputText(findings []scanning.Finding, failOnWarn bool) error {
	bold := color.New(color.Bold).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	if len(findings) == 0 {
		fmt.Printf("\n%s\n", bold("Content Scan"))
		fmt.Printf("  %s No secrets or prompt-injection patterns detected.\n", green("✔"))
		fmt.Println()
		color.Green("All checks passed.")
		return nil
	}

	// Group findings by file.
	byFile := map[string][]scanning.Finding{}
	var fileOrder []string
	for _, f := range findings {
		if _, seen := byFile[f.File]; !seen {
			fileOrder = append(fileOrder, f.File)
		}
		byFile[f.File] = append(byFile[f.File], f)
	}

	fmt.Printf("\n%s\n", bold("Content Scan"))

	errorCount := 0
	for _, file := range fileOrder {
		fmt.Printf("\n  %s\n", bold(file))
		for _, f := range byFile[file] {
			loc := ""
			if f.Line > 0 {
				loc = fmt.Sprintf(":%d", f.Line)
			}
			marker := red("✘")
			tag := red(fmt.Sprintf("[%s/%s]", f.Kind, f.Rule))
			fmt.Printf("    %s %s%s %s — %s\n", marker, file, loc, tag, f.Snippet)
			errorCount++
		}
	}

	fmt.Println()
	summary := fmt.Sprintf("%d issue(s) found", errorCount)
	color.Red(summary)

	_ = yellow // available for future warning severity tier
	_ = failOnWarn

	return fmt.Errorf("scan failed — %d issue(s) detected; review and remove sensitive content before committing", errorCount)
}

// outputJSON writes findings as a JSON array to stdout.
func outputJSON(findings []scanning.Finding) error {
	type jsonFinding struct {
		File    string `json:"file"`
		Line    int    `json:"line,omitempty"`
		Kind    string `json:"kind"`
		Rule    string `json:"rule"`
		Snippet string `json:"snippet"`
	}

	out := make([]jsonFinding, 0, len(findings))
	for _, f := range findings {
		out = append(out, jsonFinding{
			File:    f.File,
			Line:    f.Line,
			Kind:    f.Kind,
			Rule:    f.Rule,
			Snippet: f.Snippet,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}

	if len(findings) > 0 {
		return fmt.Errorf("scan failed — %d issue(s) detected", len(findings))
	}
	return nil
}
