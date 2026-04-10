package gitignore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	blockOpen  = "# <ajolote-ai> — managed automatically, do not edit manually"
	blockClose = "# </ajolote-ai>"
)

// Update writes (or replaces) the ajolote-ai managed block in .gitignore.
// It never touches lines outside the block.
// entries is the list of paths to ignore (one per line inside the block).
func Update(projectRoot string, entries []string) error {
	path := filepath.Join(projectRoot, ".gitignore")

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading .gitignore: %w", err)
	}

	content := string(existing)
	block := buildBlock(entries)

	if hasBlock(content) {
		content = replaceBlock(content, block)
	} else {
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + block + "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing .gitignore: %w", err)
	}

	return nil
}

// Remove deletes the ajolote-ai managed block from .gitignore entirely.
func Remove(projectRoot string) error {
	path := filepath.Join(projectRoot, ".gitignore")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading .gitignore: %w", err)
	}

	content := string(data)
	if !hasBlock(content) {
		return nil
	}

	content = replaceBlock(content, "")
	content = strings.TrimRight(content, "\n") + "\n"

	return os.WriteFile(path, []byte(content), 0o644)
}

// EntriesInBlock returns the current entries inside the managed block, or nil if absent.
func EntriesInBlock(projectRoot string) ([]string, bool) {
	path := filepath.Join(projectRoot, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	content := string(data)
	if !hasBlock(content) {
		return nil, false
	}

	start := strings.Index(content, blockOpen)
	end := strings.Index(content, blockClose)
	if start == -1 || end == -1 {
		return nil, false
	}

	inner := content[start+len(blockOpen) : end]
	var entries []string
	for _, line := range strings.Split(inner, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			entries = append(entries, line)
		}
	}
	return entries, true
}

func buildBlock(entries []string) string {
	var sb strings.Builder
	sb.WriteString(blockOpen)
	sb.WriteString("\n")
	for _, e := range entries {
		sb.WriteString(e)
		sb.WriteString("\n")
	}
	sb.WriteString(blockClose)
	return sb.String()
}

func hasBlock(content string) bool {
	return strings.Contains(content, blockOpen) && strings.Contains(content, blockClose)
}

func replaceBlock(content, newBlock string) string {
	start := strings.Index(content, blockOpen)
	end := strings.Index(content, blockClose)
	if start == -1 || end == -1 {
		return content
	}
	end += len(blockClose)

	// Include surrounding newlines to avoid leaving blank lines
	before := content[:start]
	after := content[end:]

	before = strings.TrimRight(before, "\n")
	after = strings.TrimLeft(after, "\n")

	if newBlock == "" {
		if before == "" {
			return after
		}
		if after == "" {
			return before + "\n"
		}
		return before + "\n" + after
	}

	if before == "" && after == "" {
		return newBlock + "\n"
	}
	if before == "" {
		return newBlock + "\n" + after
	}
	if after == "" {
		return before + "\n\n" + newBlock + "\n"
	}
	return before + "\n\n" + newBlock + "\n\n" + after
}
