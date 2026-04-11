# ajolote

One shared AI agent config for teams where developers use different tools.

You commit `.agents/` — a single directory of rules, skills, personas, and MCP server definitions. Every developer runs `ajolote use <their-tool>` and gets a freshly generated config in their tool's native format. No more per-tool config files committed to git, no more "it works on my Cursor but not your Claude Code."

Works with any language or framework. Supports Claude Code, Cursor, Windsurf, GitHub Copilot, Cline / Roo Code, and Aider.

**When it helps most:**
- Your team uses different tools (Alice is on Claude Code, Bob is on Cursor, CI runs Copilot)
- You want a single place to evolve agent rules without touching per-tool config files
- You need MCP servers shared across tools, with personal tokens staying local and out of git
- You want `ajolote diff` in CI to catch config drift before it reaches `main`

## The mental model

| Committed to git ✅ | Gitignored — generated locally ❌ |
|---|---|
| `.agents/config.json` — the canonical config | `CLAUDE.md`, `.claude/settings.json` |
| `.agents/rules/*.md` — your ground rules | `.cursor/rules/agents.mdc`, `.cursor/mcp.json` |
| `.agents/skills/*.md` — reusable instructions | `.windsurf/rules/agents.md` |
| `.agents/personas/*.md` — role behaviours | `.github/copilot-instructions.md` |
| `.agents/context/*.md` — project knowledge | `.clinerules`, `.roo/mcp.json`, `.roomodes` |
| `.agents/commands/*.md` — shared slash commands | `.aider.conf.yml` |

Everyone edits the `.agents/` files. Each developer runs `ajolote use <tool>` once to generate their own tool config from the shared source of truth.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh
```

## Usage

```sh
# Maintainer — one-time setup in any project (Python, Go, Java, Ruby — anything)
ajolote init

# Developer — after cloning, generate config for your tool of choice
ajolote use claude   # or cursor, windsurf, copilot, cline, aider

# Two-way sync — pull in changes from your tool, push updated config back out
ajolote sync           # syncs all tools whose files are present
ajolote sync cursor    # syncs only cursor

# Check that all referenced files and MCP servers are valid
ajolote validate

# Preview what sync would change without writing anything (useful in CI)
ajolote diff
ajolote diff cursor

# Check what's generated locally
ajolote status
```

## How sync works

`ajolote sync` runs in two directions:

**↑ Import** — reads MCP server definitions and commands from the tool's own config files and merges any new ones into `.agents/config.json`. This catches servers a developer added directly in their tool's UI, and commands written in the tool's native format.

**↓ Export** — regenerates the tool's files from the (now updated) canonical config, so rules, skills, and context are always current.

| Tool | Importable (↑) | Exported (↓) |
|---|---|---|
| claude | `.claude/settings.json`, `.claude/commands/` | `CLAUDE.md`, `.claude/settings.json`, `.claude/commands/` |
| cursor | `.cursor/mcp.json`, `.cursor/rules/` | `.cursor/rules/agents.mdc`, `.cursor/mcp.json`, `.cursor/rules/` |
| cline | `.roo/mcp.json`, `.roo/rules/` | `.clinerules`, `.roo/mcp.json`, `.roo/rules/`, `.roomodes` |
| windsurf | — | `.windsurf/rules/agents.md`, `.windsurf/workflows/` |
| copilot | — | `.github/copilot-instructions.md` |
| aider | — | `.aider.conf.yml` |

Rules always flow **config → tool** only. `.agents/config.json` is the authority for rules; the tool files are never trusted for rules.

---

## Why not just symlink?

The obvious alternative — symlinking a shared `AGENTS.md` into `CLAUDE.md`, `.cursorrules`, etc. — breaks down as soon as the tools diverge:

- **Different formats** — Claude Code uses `@file` imports; Cursor uses `.mdc` files with YAML frontmatter; Windsurf, Copilot, and Cline need the content inlined because they can't follow file references; Aider uses a `read:` YAML list; Roo Code needs a `.roomodes` JSON file. A single symlink target can't satisfy all of these.

- **MCP servers need per-tool JSON** — `.claude/settings.json` and `.cursor/mcp.json` have different schemas. Personal tokens (`scope: "user"`) should go in `~/.claude.json`, not a committed project file. A Markdown file can't express any of this.

- **Symlinks break on Windows** — many CI runners and Windows developers can't create symlinks without elevated permissions.

- **No two-way sync** — if a developer adds an MCP server directly in Cursor's UI, a symlink won't propagate it back to the shared config. `ajolote sync` does.

---

## config.json reference

`.agents/config.json` is the only file you edit. Everything else is generated from it.

```jsonc
{
  // MCP servers shared across all tools that support MCP
  "mcp": {
    "servers": {
      // Committed to git — all team members get this server
      "filesystem": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
        "description": "Read/write access to this repo",
        "scope": "project"  // default — written to .claude/settings.json, .cursor/mcp.json, etc.
      },
      // HTTP-transport server (no command/args)
      "remote-api": {
        "transport": "http",
        "url": "https://mcp.example.com/api",
        "description": "Remote MCP server over HTTP"
      },
      // User-scoped — written to ~/.claude.json and ~/.cursor/mcp.json, never committed
      "personal-figma": {
        "command": "npx",
        "args": ["-y", "figma-mcp"],
        "env": { "FIGMA_TOKEN": "${FIGMA_TOKEN}" },
        "description": "Figma MCP with personal access token",
        "scope": "user"
      },
      "github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": { "GITHUB_TOKEN": "${GITHUB_TOKEN}" },
        "description": "GitHub API — issues, PRs, branches"
      }
    }
  },

  // Rule files — paths to .md files agents read as ground rules (always applied)
  "rules": [
    ".agents/rules/general.md",
    ".agents/rules/code-style.md"
  ],

  // Scoped rules — only applied when the agent is working on matching files
  "scoped_rules": [
    {
      "name": "frontend",
      "globs": ["**/*.tsx", "**/*.css"],
      "path": ".agents/rules/frontend.md"
    },
    {
      "name": "api",
      "globs": ["src/app/api/**/*.ts"],
      "path": ".agents/rules/api.md"
    }
  ],

  // Skill files — reusable task instructions in .agents/skills/
  "skills": [
    ".agents/skills/git.md",
    ".agents/skills/testing.md"
  ],

  // Personas — role-based behaviours agents should adopt for specific tasks.
  // Simple string form (backward compatible) or object form with optional Claude subagent metadata.
  "personas": [
    ".agents/personas/reviewer.md",  // simple form — rendered as @file import in CLAUDE.md
    {
      "path": ".agents/personas/architect.md",
      "claude": {
        // Generates .claude/agents/architect.md — a proper Claude Code subagent file
        "model": "haiku",               // "haiku" | "sonnet" | "opus" | full model ID
        "tools": ["Read", "Grep", "Glob"],
        "description": "Software architect. Invoke for design decisions and trade-off analysis."
      }
    }
  ],

  // Context — background knowledge about the project (keep these up to date)
  "context": [
    ".agents/context/architecture.md",
    ".agents/context/data-model.md",
    ".agents/context/glossary.md"
  ]
}
```

### Fields

| Field | Type | Description |
|---|---|---|
| `mcp.servers` | object | MCP server definitions — translated to each tool's MCP config format |
| `mcp.servers.<name>.command` | string | Binary to run (e.g. `npx`, `node`) — omit for HTTP/SSE servers |
| `mcp.servers.<name>.args` | string[] | Arguments passed to the command |
| `mcp.servers.<name>.env` | object | Environment variables (use `${VAR}` to reference shell env) |
| `mcp.servers.<name>.description` | string | Human-readable description (optional) |
| `mcp.servers.<name>.transport` | string | `"stdio"` (default) \| `"http"` \| `"sse"` |
| `mcp.servers.<name>.url` | string | Server URL — required for `http`/`sse` transport |
| `mcp.servers.<name>.scope` | string | `"project"` (default) — written to project config files and committed; `"user"` — written to `~/.claude.json` / `~/.cursor/mcp.json` and never committed |
| `rules` | string[] | Paths to rule files — always applied to every file and context |
| `scoped_rules` | object[] | Rules that only activate for files matching specific glob patterns |
| `scoped_rules[].name` | string | Identifier used as the output filename (e.g. `frontend`) |
| `scoped_rules[].globs` | string[] | Glob patterns that trigger this rule (e.g. `**/*.tsx`) |
| `scoped_rules[].path` | string | Path to the `.md` rule file in `.agents/rules/` |
| `skills` | string[] | Paths to skill files — reusable task instructions |
| `personas` | string[] \| object[] | Persona entries — plain path string or object with optional `claude` block |
| `personas[].path` | string | Path to the `.md` persona file |
| `personas[].claude` | object | Optional — generates a `.claude/agents/<name>.md` subagent file for Claude Code |
| `personas[].claude.model` | string | `"haiku"` \| `"sonnet"` \| `"opus"` or full model ID |
| `personas[].claude.tools` | string[] | Claude Code tool names the subagent is allowed to use |
| `personas[].claude.description` | string | Auto-invocation trigger text; derived from first paragraph of persona file if omitted |
| `context` | string[] | Paths to context files — background knowledge about the project |

---

## What each tool gets

Each tool gets a rendering strategy appropriate for its capabilities:

| Tool | Strategy |
|---|---|
| Claude Code | `@file` imports — Claude Code follows references natively |
| Cursor | File path references — Cursor follows them in `.mdc` files |
| Copilot, Windsurf, Cline | Inline — actual file content embedded directly |
| Aider | `read:` list — Aider includes files as context natively |

### Claude Code — `ajolote use claude`

Generates two files:

**`CLAUDE.md`** — read automatically by Claude Code at the start of every session. Uses `@file` import syntax so Claude Code loads each file as additional context:

```markdown
# CLAUDE.md

> Generated by ajolote-ai from `.agents/config.json` — do not edit manually.

## Rules

@.agents/rules/general.md
@.agents/rules/code-style.md

## Skills

@.agents/skills/git.md
@.agents/skills/testing.md

## Personas

@.agents/personas/reviewer.md

## Context

@.agents/context/architecture.md
```

**`.claude/settings.json`** — registers MCP servers with Claude Code:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "description": "Read/write access to this repo"
    }
  }
}
```

**`.claude/agents/<name>.md`** — generated for each persona that has a `claude:` block. This creates a proper Claude Code subagent with YAML frontmatter:

```markdown
---
name: Architect
description: Software architect. Invoke for design decisions and trade-off analysis.
model: claude-haiku-4-5-20251001
tools:
  - Read
  - Grep
  - Glob
---

@.agents/personas/architect.md
```

Personas without a `claude:` block are not affected — they continue to appear as `@file` imports in `CLAUDE.md`.

---

### Cursor — `ajolote use cursor`

**`.cursor/rules/agents.mdc`** — applied as a persistent rule in Cursor's AI panel:

```markdown
---
description: Shared agent rules — generated by ajolote-ai, do not edit manually
alwaysApply: true
---

## Rules

- `.agents/rules/general.md`
- `.agents/rules/code-style.md`

## Skills

- `.agents/skills/git.md`
- `.agents/skills/testing.md`

## Personas

- `.agents/personas/reviewer.md`

## Context

- `.agents/context/architecture.md`
```

**`.cursor/mcp.json`** — registers MCP servers with Cursor:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}
```

---

### Windsurf — `ajolote use windsurf`

**`.windsurf/rules/agents.md`** — loaded as a persistent rule. Content from each referenced file is embedded directly:

```markdown
> Generated by ajolote-ai from `.agents/config.json` — do not edit manually.

## Rules

<!-- .agents/rules/general.md -->

# General Rules

Always read before writing. Always read the relevant files before modifying them.
...

<!-- .agents/rules/code-style.md -->

# Code Style

Follow the existing conventions in each file...
```

---

### GitHub Copilot — `ajolote use copilot`

**`.github/copilot-instructions.md`** — injected into every Copilot Chat session in VS Code. Content is inlined directly:

```markdown
<!-- Generated by ajolote-ai from .agents/config.json — do not edit manually -->

## Rules

<!-- .agents/rules/general.md -->

# General Rules

Always read before writing...

## Skills

<!-- .agents/skills/git.md -->

# Git

Use feature branches...
```

---

### Cline / Roo — `ajolote use cline`

**`.clinerules`** — loaded as system context by Cline and Roo. Content is inlined directly:

```
# Agent Rules
# Generated by ajolote-ai from .agents/config.json — do not edit manually

## Rules

<!-- .agents/rules/general.md -->

# General Rules

Always read before writing...
```

**`.roo/mcp.json`** — registers MCP servers with Roo:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}
```

**`.roomodes`** — Roo Code custom modes, one per persona. Each persona becomes an invocable mode with its full content as the role definition:

```json
{
  "customModes": [
    {
      "slug": "reviewer",
      "name": "Reviewer",
      "roleDefinition": "# Persona: Code Reviewer\n\nWhen acting as a code reviewer...",
      "groups": ["read", "edit", "browser", "command", "mcp"],
      "source": "project"
    },
    {
      "slug": "architect",
      "name": "Architect",
      "roleDefinition": "# Persona: Architect\n\nWhen acting as a software architect...",
      "groups": ["read", "edit", "browser", "command", "mcp"],
      "source": "project"
    }
  ]
}
```

Roo Code users can switch to these modes with `/mode reviewer` or by selecting them from the mode picker. Modes receive full tool access by default — restrict `groups` manually in `.roomodes` if needed.

---

### Aider — `ajolote use aider`

**`.aider.conf.yml`** — Aider project configuration. Uses Aider's native `read:` option so Aider loads the files as context directly:

```yaml
# Generated by ajolote-ai from .agents/config.json — do not edit manually

# aider configuration
auto-commits: false
gitignore: false

# Files to include as context
read:
  - .agents/rules/general.md
  - .agents/rules/code-style.md
  - .agents/skills/git.md
  - .agents/skills/testing.md
  - .agents/personas/reviewer.md
  - .agents/context/architecture.md
```

---

## The .agents/ directory

All files in `.agents/` are committed to git and shared across the team. Agents receive these as context when developers run `ajolote use <tool>`.

```
.agents/
  config.json              ← canonical config (the only file ajolote reads)
  rules/
    general.md             ← General agent behaviour rules
    code-style.md          ← Code style and formatting conventions
  skills/
    git.md                 ← Git workflows: branch naming, PR hygiene
    testing.md             ← How to write and run tests
    api-design.md          ← REST / GraphQL conventions (add your own)
    migrations.md          ← DB migration workflow (add your own)
    debugging.md           ← Step-by-step debugging protocol (add your own)
  personas/
    reviewer.md            ← How to behave during code review
    architect.md           ← How to approach design and architecture tasks
  context/
    architecture.md        ← High-level system design (keep up to date!)
    data-model.md          ← Core entities and relationships
    glossary.md            ← Domain terms and abbreviations
  commands/
    deploy.md              ← Custom slash commands shared across all tools
```

`ajolote init` seeds all directories with starter templates. Fill them in with your project's real information and commit — that's where the value is.

### Scoped Rules

Rules that only activate for files matching specific glob patterns. Each tool renders them in its native format:

| Tool | Output | Format |
|---|---|---|
| Cursor | `.cursor/rules/<name>.mdc` | `globs:` frontmatter — Cursor applies the rule only to matching files |
| Copilot | `.github/instructions/<name>.instructions.md` | `applyTo:` frontmatter — VS Code Copilot's per-file instruction system |
| Claude Code | `.claude/rules/<name>.md` | `globs:` frontmatter + `@file` import |
| Windsurf | `.windsurf/rules/<name>.md` | `globs:` frontmatter |
| Cline | `.roo/rules/<name>.md` | Inlined content |
| Aider | `.aider.conf.yml` `read:` list | No glob support — always included |

**Two-way sync** — `ajolote sync cursor` or `ajolote sync copilot` imports user-authored scoped rules from `.cursor/rules/*.mdc` (by detecting a `globs:` field) or `.github/instructions/*.instructions.md` (by detecting `applyTo:`), writes the body to `.agents/rules/<name>.md`, and adds the entry to `config.json`.

### Rules

Agent ground rules and conventions — how agents should behave, write code, and interact with the project. Reference under `"rules"` in `config.json`:

```json
"rules": [
  ".agents/rules/general.md",
  ".agents/rules/code-style.md"
]
```

`ajolote init` creates `general.md` and `code-style.md` as starter templates. Edit them to match your project's real conventions. Add new rule files for testing, security, commits, or any convention your team wants to enforce.

### Skills

Reusable task instructions. Reference them under `"skills"` in `config.json`:

```json
"skills": [
  ".agents/skills/git.md",
  ".agents/skills/testing.md",
  ".agents/skills/api-design.md"
]
```

### Personas

Role-based behaviours agents adopt for specific task types. Reference under `"personas"`:

```json
"personas": [
  ".agents/personas/reviewer.md",
  ".agents/personas/architect.md"
]
```

Example usage — a developer tells their AI tool:
> "Act as the architect persona and review this proposed change to the auth service."

The agent reads `.agents/personas/architect.md` and applies its mindset and principles.

### Context

Background knowledge about the project — architecture, data model, glossary. Reference under `"context"`:

```json
"context": [
  ".agents/context/architecture.md",
  ".agents/context/data-model.md",
  ".agents/context/glossary.md"
]
```

Keep these files accurate. They are the most important inputs for agents working on non-trivial tasks.

### Commands

Custom slash commands shared across the whole team. Place `.md` files in `.agents/commands/` and they are automatically translated into each tool's native command format on `ajolote use <tool>` or `ajolote sync`.

| Tool | Command format |
|---|---|
| Claude Code | `.claude/commands/<name>.md` |
| Cursor | `.cursor/rules/<name>.mdc` |
| Windsurf | `.windsurf/workflows/<name>.yaml` |
| Cline / Roo | `.roo/rules/<name>.md` |

Each `.agents/commands/<name>.md` file should start with a one-line description followed by the command body:

```markdown
Run the full test suite and report coverage.

```sh
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```
```

---

## Supported tools

| Tool | Config files generated |
|---|---|
| [Claude Code](https://claude.ai/code) | `CLAUDE.md`, `.claude/settings.json`, `.claude/commands/` |
| [Cursor](https://cursor.com) | `.cursor/rules/agents.mdc`, `.cursor/mcp.json`, `.cursor/rules/` |
| [Windsurf](https://windsurf.com) | `.windsurf/rules/agents.md`, `.windsurf/workflows/` |
| [GitHub Copilot](https://github.com/features/copilot) | `.github/copilot-instructions.md` |
| [Cline / Roo Code](https://github.com/cline/cline) | `.clinerules`, `.roo/mcp.json`, `.roo/rules/`, `.roomodes` |
| [Aider](https://aider.chat) | `.aider.conf.yml` |
