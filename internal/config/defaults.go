package config

// DefaultConfig returns a scaffold Config for a new project.
// projectName is typically the base name of the project directory.
func DefaultConfig(projectName string) *Config {
	return &Config{
		Project: Project{
			Name:       projectName,
			Stack:      "",
			RepoType:   "",
			Language:   "",
			TestRunner: "",
		},
		MCP: MCP{
			Servers: map[string]MCPServer{},
		},
		Rules: Rules{
			General: []string{
				"Read before writing. Always read the relevant files before modifying them.",
				"Minimal diffs. Change only what is necessary to satisfy the task.",
				"No unsolicited refactors. Do not rename or restructure code unless explicitly asked.",
				"No hallucinated APIs. Check installed versions before using a library function.",
			},
			CodeStyle: []string{
				"Follow the existing conventions in each file. When in doubt, run the linter.",
				"Do not introduce new dependencies without a comment explaining why.",
				"Match the surrounding indentation, quote style, and semicolon usage exactly.",
			},
			Testing: []string{
				"Every new function or module must have at least one accompanying test.",
				"Do not delete or weaken existing tests.",
			},
			Security: []string{
				"Never hard-code secrets, tokens, or credentials — use environment variables.",
				"Never log sensitive user data.",
			},
			Commits: []string{
				"Write commit messages in the imperative mood: Add X, Fix Y, Remove Z.",
				"One logical change per commit.",
			},
		},
		Skills: []string{
			".agents/skills/git.md",
			".agents/skills/testing.md",
		},
	}
}

// GitSkillContent is the default content seeded into .agents/skills/git.md.
const GitSkillContent = `# Git Workflow

## Branch Naming
- Features: ` + "`feature/<ticket>-<short-description>`" + `
- Bug fixes: ` + "`fix/<ticket>-<short-description>`" + `
- Chores: ` + "`chore/<short-description>`" + `

## Commit Messages
- Imperative mood: ` + "`Add X`" + `, ` + "`Fix Y`" + `, ` + "`Remove Z`" + `
- Reference the relevant issue when known: ` + "`Fix login redirect (#42)`" + `
- One logical change per commit

## Pull Requests
- Keep PRs small and focused
- Include a clear description of what changed and why
- Request review before merging to main
`

// TestingSkillContent is the default content seeded into .agents/skills/testing.md.
const TestingSkillContent = `# Testing Guidelines

## General
- Every new function or module must have at least one test
- Tests should be deterministic and not depend on external state
- Do not delete or weaken existing tests

## Test Structure
- Arrange: set up the test inputs and state
- Act: call the function under test
- Assert: verify the output matches expectations

## Running Tests
- Run the full test suite before marking a task complete
- Fix failing tests before adding new features
`
