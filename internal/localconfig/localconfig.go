// Package localconfig manages the developer-local, gitignored override file
// .agents/config.local.json. It is never committed and never affects teammates.
package localconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Filename is the project-relative path of the local config file.
const Filename = ".agents/config.local.json"

// LocalConfig holds developer-local overrides. Currently supports a single
// "protect" list — files (or glob patterns) that ajolote must never overwrite.
type LocalConfig struct {
	Protect []string `json:"protect,omitempty"`
}

// Load reads .agents/config.local.json under projectRoot.
// Returns an empty (non-nil) config and no error when the file is absent —
// the common case for developers who have not set up any protection.
// Returns nil and an error only when the file exists but cannot be parsed.
func Load(projectRoot string) (*LocalConfig, error) {
	path := filepath.Join(projectRoot, Filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &LocalConfig{}, nil
		}
		return nil, err
	}
	var lc LocalConfig
	if err := json.Unmarshal(data, &lc); err != nil {
		return nil, err
	}
	return &lc, nil
}

// IsProtected reports whether relPath matches any pattern in the protect list.
//
// Pattern semantics (filepath.Match rules apply):
//   - Exact match:     "CLAUDE.md"        → protects only CLAUDE.md
//   - Glob:            ".claude/commands/*.md" → protects any .md file in that dir
//   - Directory:       ".claude/commands/"     → protects everything under that dir
//
// Calling IsProtected on a nil *LocalConfig always returns false — safe zero value.
func (lc *LocalConfig) IsProtected(relPath string) bool {
	if lc == nil {
		return false
	}
	// Normalise to forward slashes so patterns work cross-platform.
	relPath = filepath.ToSlash(relPath)
	for _, pattern := range lc.Protect {
		pattern = filepath.ToSlash(pattern)

		// Directory prefix — pattern ending with "/" protects everything inside.
		if strings.HasSuffix(pattern, "/") {
			if strings.HasPrefix(relPath, pattern) {
				return true
			}
			continue
		}

		matched, err := filepath.Match(pattern, relPath)
		if err == nil && matched {
			return true
		}
	}
	return false
}
