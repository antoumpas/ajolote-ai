// Package scanning detects leaked secrets and prompt-injection payloads in
// .agents/rules/ and .agents/skills/ markdown files.
package scanning

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Finding describes one detected issue in a scanned file.
type Finding struct {
	File    string // project-relative path
	Line    int    // 1-based line number (0 = whole-file match)
	Kind    string // "secret" or "injection"
	Rule    string // rule identifier, e.g. "aws-access-key"
	Snippet string // safe excerpt — secrets are redacted
}

// pattern is an internal rule that the Detector applies to file content.
type pattern struct {
	rule     string
	kind     string // "secret" | "injection"
	re       *regexp.Regexp
	perLine  bool // true → apply per-line; false → apply to full content
	redact   bool // true → redact matched value in snippet
}

// Detector holds compiled detection patterns and runs them against files.
type Detector struct {
	patterns []pattern
}

// NewDetector builds a Detector pre-loaded with all built-in rules.
func NewDetector() *Detector {
	d := &Detector{}
	d.addSecrets()
	d.addInjections()
	return d
}

// addSecrets registers secret-detection patterns (applied per line).
func (d *Detector) addSecrets() {
	secrets := []struct {
		rule string
		expr string
	}{
		// AWS Access Key IDs always start with AKIA and are 20 chars total.
		{"aws-access-key", `\bAKIA[0-9A-Z]{16}\b`},
		// AWS secret key assignments.
		{"aws-secret-key", `(?i)aws[_-]?secret[_-]?access[_-]?key\s*[:=]\s*\S{20,}`},
		// GitHub Personal Access Tokens (classic and fine-grained).
		// Classic PATs: ghp_ + 36 alphanumeric chars.
		{"github-pat", `\bghp_[A-Za-z0-9]{36}\b`},
		// Fine-grained PATs: github_pat_ + 82 chars.
		{"github-pat-fine", `\bgithub_pat_[A-Za-z0-9_]{82}\b`},
		// Slack tokens.
		{"slack-token", `\bxox[baprs]-[A-Za-z0-9][A-Za-z0-9_-]{8,}\b`},
		// PEM private key headers.
		{"private-key", `-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`},
		// Database URIs with embedded credentials.
		{"database-uri", `(?i)(postgres|postgresql|mysql|mongodb|redis|mssql):\/\/[^:\s]+:[^@\s]+@`},
		// Generic API key / token assignments (value ≥ 20 chars).
		{"generic-api-key", `(?i)(api[_-]?key|apikey|access[_-]?token|auth[_-]?token|secret[_-]?key|client[_-]?secret)\s*[:=]\s*['"]?[A-Za-z0-9+/=_\-\.]{20,}`},
		// Generic Bearer / token header values (value must be ≥ 20 chars to avoid matching
		// short placeholder strings like "Bearer mytoken").
		{"bearer-token", `(?i)\bBearer\s+[A-Za-z0-9\-\._~\+\/]{20,}=*\b`},
	}
	for _, s := range secrets {
		d.patterns = append(d.patterns, pattern{
			rule:    s.rule,
			kind:    "secret",
			re:      regexp.MustCompile(s.expr),
			perLine: true,
			redact:  true,
		})
	}
}

// addInjections registers prompt-injection detection patterns.
func (d *Detector) addInjections() {
	// Per-line injection patterns (case-insensitive).
	linePatterns := []struct {
		rule string
		expr string
	}{
		{"ignore-instructions", `(?i)ignore\s+(all\s+)?(previous|above|prior|earlier)\s+instructions?`},
		{"disregard-rules", `(?i)disregard\s+(your|all|previous|prior)(\s+\w+)?\s+(instructions?|rules?|guidelines?|constraints?)`},
		{"persona-hijack", `(?i)you\s+are\s+now\s+(a\s+|an\s+)?[a-zA-Z]`},
		{"override-instructions", `(?i)(override|bypass|ignore|forget)\s+(all\s+)?(your\s+)?(previous\s+)?(instructions?|rules?|guidelines?|training)`},
		{"do-not-follow", `(?i)do\s+not\s+follow\s+(the\s+)?(above|previous|prior|these)\s+(instructions?|rules?|guidelines?)`},
		{"fake-system-delimiter", `(?i)(system|user|assistant|human)\s*:\s*(you\s+are|ignore|your\s+new)`},
		{"special-token-injection", `<\|[a-z_]+\|>`},
		{"jailbreak-prefix", `(?i)(DAN|STAN|DUDE|AIM|KEVIN)\s*(mode|prompt|jailbreak|persona)`},
	}
	for _, p := range linePatterns {
		d.patterns = append(d.patterns, pattern{
			rule:    p.rule,
			kind:    "injection",
			re:      regexp.MustCompile(p.expr),
			perLine: true,
			redact:  false,
		})
	}

	// Whole-file patterns (control characters, etc.).
	wholeFilePatterns := []struct {
		rule string
		expr string
	}{
		{"null-byte", `\x00`},
		{"bare-carriage-return", `\r[^\n]`},
	}
	for _, p := range wholeFilePatterns {
		d.patterns = append(d.patterns, pattern{
			rule:    p.rule,
			kind:    "injection",
			re:      regexp.MustCompile(p.expr),
			perLine: false,
			redact:  false,
		})
	}
}

// ScanFile applies all patterns to the given content and returns findings.
// path should be a project-relative path used in Finding.File.
func (d *Detector) ScanFile(path string, content []byte) []Finding {
	var findings []Finding

	// Per-line patterns.
	lines := bytes.Split(content, []byte("\n"))
	for _, p := range d.patterns {
		if !p.perLine {
			continue
		}
		for lineNum, line := range lines {
			if loc := p.re.FindIndex(line); loc != nil {
				snippet := buildSnippet(line, loc, p.redact)
				findings = append(findings, Finding{
					File:    path,
					Line:    lineNum + 1,
					Kind:    p.kind,
					Rule:    p.rule,
					Snippet: snippet,
				})
			}
		}
	}

	// Whole-file patterns.
	for _, p := range d.patterns {
		if p.perLine {
			continue
		}
		if p.re.Match(content) {
			findings = append(findings, Finding{
				File:    path,
				Kind:    p.kind,
				Rule:    p.rule,
				Snippet: fmt.Sprintf("control character detected (%s)", p.rule),
			})
		}
	}

	return findings
}

// ScanDir scans all .md files under agentsDir/rules and agentsDir/skills,
// returning all findings. agentsDir should be the absolute path to .agents/.
func (d *Detector) ScanDir(agentsDir string) ([]Finding, error) {
	var all []Finding

	dirsToScan := []string{
		filepath.Join(agentsDir, "rules"),
		filepath.Join(agentsDir, "skills"),
		filepath.Join(agentsDir, "personas"),
		filepath.Join(agentsDir, "context"),
		filepath.Join(agentsDir, "commands"),
	}

	for _, dir := range dirsToScan {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(path), ".md") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", path, err)
			}
			// Use path relative to the parent of agentsDir so findings show ".agents/rules/..."
			rel, err := filepath.Rel(filepath.Dir(agentsDir), path)
			if err != nil {
				rel = path
			}
			all = append(all, d.ScanFile(rel, content)...)
			return nil
		})
		if err != nil {
			return all, err
		}
	}

	return all, nil
}

// ScanFiles scans a list of project-relative file paths from the given root.
func (d *Detector) ScanFiles(projectRoot string, paths []string) ([]Finding, error) {
	var all []Finding
	seen := map[string]bool{}
	for _, rel := range paths {
		if seen[rel] {
			continue
		}
		seen[rel] = true
		full := filepath.Join(projectRoot, rel)
		content, err := os.ReadFile(full)
		if err != nil {
			// Validate handles missing files; skip silently here.
			continue
		}
		all = append(all, d.ScanFile(rel, content)...)
	}
	return all, nil
}

// buildSnippet produces a display-safe excerpt from a matched line.
// When redact is true, the matched segment is partially hidden.
func buildSnippet(line []byte, loc []int, redact bool) string {
	matched := string(line[loc[0]:loc[1]])
	if !redact {
		// Trim to 80 chars for readability.
		display := strings.TrimSpace(string(line))
		if len(display) > 80 {
			display = display[:77] + "..."
		}
		return display
	}
	// Redact: keep up to 4 characters of the matched value, then mask the rest.
	show := matched
	if len(show) > 4 {
		show = show[:4] + strings.Repeat("*", min(len(show)-4, 8))
	}
	prefix := strings.TrimSpace(string(line[:loc[0]]))
	if len(prefix) > 40 {
		prefix = "..." + prefix[len(prefix)-37:]
	}
	return fmt.Sprintf("%s%s", prefix, show)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
