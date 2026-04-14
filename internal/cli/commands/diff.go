package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ajolote-ai/ajolote/internal/config"
	"github.com/ajolote-ai/ajolote/internal/translators"
)

func DiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff [<tool>]",
		Short: "Show what ajolote sync would change without writing anything",
		Long: `Generates each tool's config into a temporary directory and compares it
against the files currently on disk. Prints a unified diff for every file
that would be created or modified. Exits with status 1 if any file would
change — useful in CI to catch config drift before it reaches main.

Only tools that already have at least one output file on disk are checked.
User-scoped MCP server entries (~/.claude.json, ~/.cursor/mcp.json) are
excluded from the diff.`,
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		RunE:         runDiff,
		Example:      "  ajolote diff\n  ajolote diff cursor\n  ajolote diff claude",
	}
}

func runDiff(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectRoot)
	if err != nil {
		return err
	}

	// Resolve inheritance so the diff reflects what sync would actually generate.
	// This also populates .agents/.base/ which is then copied into each temp dir.
	cfg, err = config.Resolve(cfg, projectRoot)
	if err != nil {
		return err
	}

	// Determine which tools to diff (same logic as sync)
	var targets []translators.Syncer
	if len(args) == 1 {
		t, err := translators.Get(args[0])
		if err != nil {
			return err
		}
		targets = []translators.Syncer{t}
	} else {
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

	bold := color.New(color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	totalChanged, totalNew := 0, 0

	for _, t := range targets {
		fmt.Printf("\n%s\n", bold(t.Name()))

		// Generate into a temp directory so we never touch real files.
		// Copy .agents/ into the temp dir first so Generate can read commands,
		// personas, and other source files via the standard projectRoot paths.
		tmpDir, err := os.MkdirTemp("", "ajolote-diff-*")
		if err != nil {
			return fmt.Errorf("creating temp dir: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		agentsDir := filepath.Join(projectRoot, ".agents")
		if _, err := os.Stat(agentsDir); err == nil {
			if err := copyDir(agentsDir, filepath.Join(tmpDir, ".agents")); err != nil {
				return fmt.Errorf("preparing temp dir: %w", err)
			}
		}

		if err := t.Generate(cfg, tmpDir); err != nil {
			return fmt.Errorf("generating %s config: %w", t.Name(), err)
		}

		// Walk every file written into the temp dir
		err = filepath.Walk(tmpDir, func(tmpPath string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(tmpDir, tmpPath)
			relPath = filepath.ToSlash(relPath) // normalise to forward slashes on Windows
			realPath := filepath.Join(projectRoot, relPath)

			generated, err := os.ReadFile(tmpPath)
			if err != nil {
				return err
			}

			existing, err := os.ReadFile(realPath)
			if os.IsNotExist(err) {
				// New file
				fmt.Printf("  %s %s\n", cyan("+"), relPath)
				printDiffLines(nil, generated)
				totalNew++
				return nil
			}
			if err != nil {
				return err
			}

			if string(existing) == string(generated) {
				fmt.Printf("  %s %s\n", green("✔"), relPath)
				return nil
			}

			// Changed file — show unified diff
			fmt.Printf("  %s %s\n", red("~"), relPath)
			d := unifiedDiff(string(existing), string(generated), relPath)
			for _, line := range strings.Split(strings.TrimRight(d, "\n"), "\n") {
				printDiffLine(line)
			}
			totalChanged++
			return nil
		})
		if err != nil {
			return fmt.Errorf("comparing %s output: %w", t.Name(), err)
		}
	}

	fmt.Println()
	if totalChanged == 0 && totalNew == 0 {
		color.Green("Nothing would change.")
		return nil
	}

	parts := []string{}
	if totalChanged > 0 {
		parts = append(parts, fmt.Sprintf("%d file(s) would change", totalChanged))
	}
	if totalNew > 0 {
		parts = append(parts, fmt.Sprintf("%d file(s) would be created", totalNew))
	}
	color.Yellow(strings.Join(parts, ", "))
	return fmt.Errorf("diff: run ajolote sync to apply")
}

// printDiffLine prints a unified diff line with appropriate color.
func printDiffLine(line string) {
	switch {
	case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
		fmt.Println("     " + color.New(color.Bold).Sprint(line))
	case strings.HasPrefix(line, "@@"):
		fmt.Println("     " + color.New(color.FgCyan).Sprint(line))
	case strings.HasPrefix(line, "+"):
		fmt.Println("     " + color.New(color.FgGreen).Sprint(line))
	case strings.HasPrefix(line, "-"):
		fmt.Println("     " + color.New(color.FgRed).Sprint(line))
	default:
		fmt.Println("     " + line)
	}
}

// printDiffLines prints a new file's content as pure additions.
func printDiffLines(_, content []byte) {
	for _, line := range strings.Split(strings.TrimRight(string(content), "\n"), "\n") {
		fmt.Println("     " + color.New(color.FgGreen).Sprint("+"+line))
	}
}

// copyDir recursively copies src into dst, creating dst and all subdirectories as needed.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

// ── Unified diff ──────────────────────────────────────────────────────────────

const diffContext = 3

type lineOp int

const (
	opKeep lineOp = iota
	opDelete
	opInsert
)

type lineEdit struct {
	op   lineOp
	text string
}

// unifiedDiff returns a unified diff between oldContent and newContent.
// Returns an empty string when the content is identical.
func unifiedDiff(oldContent, newContent, path string) string {
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	edits := lcsEdits(oldLines, newLines)

	// Mark which edit indices fall within diffContext lines of a real change
	inContext := make([]bool, len(edits))
	for i, e := range edits {
		if e.op != opKeep {
			for j := max(0, i-diffContext); j <= min(len(edits)-1, i+diffContext); j++ {
				inContext[j] = true
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("--- " + path + "\n")
	sb.WriteString("+++ " + path + " (generated)\n")

	oldLine, newLine := 1, 1
	inHunk := false
	var hunkBuf strings.Builder
	var hunkOldStart, hunkNewStart, hunkOldCount, hunkNewCount int

	flushHunk := func() {
		if inHunk {
			sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
				hunkOldStart, hunkOldCount, hunkNewStart, hunkNewCount))
			sb.WriteString(hunkBuf.String())
			hunkBuf.Reset()
			inHunk = false
		}
	}

	for i, e := range edits {
		if !inContext[i] {
			flushHunk()
			if e.op != opInsert {
				oldLine++
			}
			if e.op != opDelete {
				newLine++
			}
			continue
		}
		if !inHunk {
			inHunk = true
			hunkOldStart, hunkNewStart = oldLine, newLine
			hunkOldCount, hunkNewCount = 0, 0
		}
		switch e.op {
		case opKeep:
			hunkBuf.WriteString(" " + e.text + "\n")
			hunkOldCount++
			hunkNewCount++
			oldLine++
			newLine++
		case opDelete:
			hunkBuf.WriteString("-" + e.text + "\n")
			hunkOldCount++
			oldLine++
		case opInsert:
			hunkBuf.WriteString("+" + e.text + "\n")
			hunkNewCount++
			newLine++
		}
	}
	flushHunk()

	return sb.String()
}

// lcsEdits computes a minimal edit script via LCS. Works well for small files
// (the typical case for generated config files).
func lcsEdits(a, b []string) []lineEdit {
	m, n := len(a), len(b)

	// dp[i][j] = length of LCS of a[:i] and b[:j]
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to produce edits (reverse order, then flip)
	edits := make([]lineEdit, 0, m+n)
	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && a[i-1] == b[j-1]:
			edits = append(edits, lineEdit{opKeep, a[i-1]})
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			edits = append(edits, lineEdit{opInsert, b[j-1]})
			j--
		default:
			edits = append(edits, lineEdit{opDelete, a[i-1]})
			i--
		}
	}
	for l, r := 0, len(edits)-1; l < r; l, r = l+1, r-1 {
		edits[l], edits[r] = edits[r], edits[l]
	}
	return edits
}

// splitLines splits content into lines, preserving the final empty-line behavior.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	// strings.Split always produces a trailing empty element for content ending
	// in "\n" — drop it so our line counts match the file's actual line count.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
