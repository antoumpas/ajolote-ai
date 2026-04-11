package translators

import "github.com/ajolote-ai/ajolote/internal/config"

// Translator generates tool-specific config files from the canonical config.
type Translator interface {
	// Name returns the tool identifier (e.g. "claude", "cursor").
	Name() string
	// OutputFiles returns the file paths this translator will write,
	// relative to the project root. Used to populate .gitignore.
	OutputFiles() []string
	// Generate writes tool-specific files under projectRoot.
	Generate(cfg *config.Config, projectRoot string) error
}

// CommittedOutput is implemented by translators whose output files belong in
// version control rather than in .gitignore. The canonical example is the
// agents-md translator, which generates AGENTS.md — the Linux Foundation /
// AAIF standard that any AI tool can read without ajolote installed.
type CommittedOutput interface {
	Committed() bool
}
