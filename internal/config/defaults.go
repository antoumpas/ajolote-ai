package config

// DefaultConfig returns a scaffold Config for a new project.
func DefaultConfig() *Config {
	return &Config{
		MCP: MCP{
			Servers: map[string]MCPServer{},
		},
		Rules: []string{
			".agents/rules/general.md",
			".agents/rules/code-style.md",
		},
		Skills: []string{
			".agents/skills/git.md",
			".agents/skills/testing.md",
		},
		Personas: []Persona{
			{Path: ".agents/personas/reviewer.md"},
			{Path: ".agents/personas/architect.md"},
		},
		Context: []string{
			".agents/context/architecture.md",
			".agents/context/data-model.md",
			".agents/context/glossary.md",
		},
	}
}

// GeneralRulesContent is seeded into .agents/rules/general.md.
const GeneralRulesContent = `# General Rules

- Read before writing. Always read the relevant files before modifying them.
- Minimal diffs. Change only what is necessary to satisfy the task.
- No unsolicited refactors. Do not rename or restructure code unless explicitly asked.
- No hallucinated APIs. Check installed versions before using a library function.

## Testing

- Every new function or module must have at least one accompanying test.
- Do not delete or weaken existing tests.

## Security

- Never hard-code secrets, tokens, or credentials — use environment variables.
- Never log sensitive user data.

## Commits

- Write commit messages in the imperative mood: Add X, Fix Y, Remove Z.
- One logical change per commit.
`

// CodeStyleRulesContent is seeded into .agents/rules/code-style.md.
const CodeStyleRulesContent = `# Code Style

- Follow the existing conventions in each file. When in doubt, run the linter.
- Do not introduce new dependencies without a comment explaining why.
- Match the surrounding indentation, quote style, and semicolon usage exactly.
`

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

// ReviewerPersonaContent is seeded into .agents/personas/reviewer.md.
const ReviewerPersonaContent = `# Persona: Code Reviewer

When acting as a code reviewer, adopt the following mindset and priorities.

## Mindset
- Be constructive and specific — point to the exact line and explain why
- Distinguish between blocking issues and suggestions
- Assume good intent; ask questions before concluding something is wrong

## Review Checklist
- [ ] Does the change do what the PR description says?
- [ ] Are edge cases handled?
- [ ] Are new functions covered by tests?
- [ ] Are there any security implications (secrets, input validation, auth)?
- [ ] Does the code follow existing conventions (style, naming, structure)?
- [ ] Is the change minimal — no unrelated edits snuck in?

## Tone
- Prefer "Consider X" or "What do you think about Y?" over "You must" or "This is wrong"
- Acknowledge good decisions, not just problems
`

// ArchitectPersonaContent is seeded into .agents/personas/architect.md.
const ArchitectPersonaContent = `# Persona: Architect

When acting as a software architect, adopt the following mindset and priorities.

## Mindset
- Think in trade-offs, not absolutes — every decision has a cost
- Optimise for the team's ability to change the system safely over time
- Prefer proven patterns over clever solutions

## Design Principles
- Keep services and modules loosely coupled, highly cohesive
- Define clear boundaries: who owns what data, who calls whom
- Make failure modes explicit — what happens when a dependency is down?
- Avoid premature optimisation; measure before changing for performance

## When Proposing Changes
- State the problem clearly before proposing a solution
- List at least two alternatives with their trade-offs
- Flag any irreversible decisions explicitly
- Consider operational concerns: deployment, rollback, monitoring
`

// ArchitectureContextContent is seeded into .agents/context/architecture.md.
const ArchitectureContextContent = `# Architecture

> Keep this file up to date as the system evolves. Agents use it as their primary
> source of truth about how the system is structured.

## Overview

<!-- Describe the high-level shape of the system: monolith, microservices, etc. -->

## Components

<!-- List the main services, packages, or layers and their responsibilities. -->

| Component | Responsibility |
|---|---|
| _example_ | _what it does_ |

## Key Data Flows

<!-- Describe the most important request/event flows through the system. -->

## External Dependencies

<!-- List third-party services, APIs, and databases the system relies on. -->

## Known Constraints

<!-- Performance limits, compliance requirements, legacy decisions to be aware of. -->
`

// DataModelContextContent is seeded into .agents/context/data-model.md.
const DataModelContextContent = `# Data Model

> Describe the core entities and their relationships. Keep this in sync with the schema.

## Entities

<!-- For each core entity, describe its purpose and key fields. -->

### Example Entity

- **Purpose:** What this entity represents
- **Key fields:** id, name, created_at, ...
- **Relationships:** belongs to X, has many Y

## Conventions

<!-- Naming conventions, soft-delete patterns, timestamps, etc. -->
`

// ReviewCommandContent is seeded into .agents/commands/review.md.
const ReviewCommandContent = `---
description: Review recent changes for correctness, tests, and style
---

Review the current changes (staged files, recent commits, or the diff since main).
Apply the reviewer persona from ` + "`.agents/personas/reviewer.md`" + `.

Focus on:
- Does the change do what it claims?
- Are edge cases handled?
- Is test coverage adequate?
- Are there security implications?
- Does the code follow the conventions in ` + "`.agents/skills/`" + `?

Provide feedback with specific file and line references.
Separate blocking issues from suggestions.
`

// GlossaryContextContent is seeded into .agents/context/glossary.md.
const GlossaryContextContent = `# Glossary

Domain terms and abbreviations used in this project. Agents should use these
definitions consistently when reading and writing code, comments, and docs.

| Term | Definition |
|---|---|
| _Add terms here_ | _And their definitions_ |
`
