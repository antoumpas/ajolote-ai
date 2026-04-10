# AGENTS.md

> **Purpose:** This file establishes a shared contract for all AI coding agents operating in this repository — including Claude Code, GitHub Copilot, Cursor, Windsurf, Codeium, and any other LLM-powered tool. Every agent reads this file first and adheres to it throughout any session.

---

## 1. Project Overview

<!-- Replace this section with a description of your project -->

```
Project:     <your-project-name>
Stack:       <e.g. TypeScript / Node.js / React>
Repo type:   <monorepo | polyrepo>
Primary lang: <language>
Test runner: <e.g. Vitest, Jest, Pytest>
```

---

## 2. Agent Ground Rules

The following rules apply to **all agents** regardless of the tool being used.

### 2.1 General Behaviour

- **Read before writing.** Always read the relevant files before modifying them.
- **Minimal diffs.** Change only what is necessary to satisfy the task. Do not reformat unrelated code.
- **No unsolicited refactors.** Do not rename, reorganise, or restructure code unless explicitly asked.
- **Preserve intent.** If you are unsure what existing code does, ask or add a `TODO` comment — never silently delete it.
- **No hallucinated APIs.** If you are not certain a library function exists, check the installed version in `package.json` / `pyproject.toml` / `go.mod` before using it.

### 2.2 Code Style

- Follow the existing conventions in each file. When in doubt, run the linter.
- Do not introduce new dependencies without a comment explaining why.
- Match the surrounding indentation, quote style, and semicolon usage exactly.

### 2.3 Testing

- Every new function or module must have at least one accompanying test.
- Do not delete or weaken existing tests.
- Run tests locally before marking a task complete: `<insert test command>`.

### 2.4 Security

- Never hard-code secrets, tokens, or credentials — use environment variables.
- Never log sensitive user data.
- Do not introduce `eval`, `exec`, or dynamic code execution unless the pattern already exists in the codebase.

### 2.5 Commits

- Write commit messages in the imperative mood: `Add X`, `Fix Y`, `Remove Z`.
- One logical change per commit.
- Reference the relevant issue or ticket when known: `Fix login redirect (#42)`.

---

## 3. MCP (Model Context Protocol) Configuration

All agents that support MCP should use the **shared server registry** defined here. This prevents each tool from spinning up duplicate or conflicting servers.

### 3.1 Canonical MCP Config Path

```
.mcp/config.json          # shared MCP server definitions (committed)
.mcp/config.local.json    # machine-local overrides (git-ignored)
```

### 3.2 Standard Server Registry

```jsonc
// .mcp/config.json  — edit this section to match your actual servers
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "description": "Read/write access to this repo"
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_TOKEN": "${GITHUB_TOKEN}" },
      "description": "GitHub API — issues, PRs, branches"
    },
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": { "DATABASE_URL": "${DATABASE_URL}" },
      "description": "Read-only DB introspection"
    }
    // Add project-specific servers below
  }
}
```

### 3.3 Per-tool MCP Wiring

| Tool | Config location | Notes |
|---|---|---|
| **Claude Code** | `.claude/settings.json` → `mcpServers` key | Supports full MCP; symlink or import `.mcp/config.json` |
| **Cursor** | `.cursor/mcp.json` | Paste server entries from canonical config |
| **Windsurf** | `~/.codeium/windsurf/mcp_config.json` | Global only; document project servers in a comment |
| **Copilot (VS Code)** | MCP support via extensions; check current extension docs | Limited native support as of mid-2025 |
| **Cline / Roo** | `.roo/mcp.json` or VS Code settings | Per-workspace config supported |

> **Rule:** Do not add a server to a tool-specific config that is not listed in `.mcp/config.json`. Keep the canonical file as the single source of truth.

---

## 4. Skills & Prompt Libraries

"Skills" are reusable instructions that agents can be given to perform common tasks consistently. Store them here so every tool can reference the same source.

### 4.1 Directory Layout

```
.agents/
  skills/
    git.md              # Git workflows: branch naming, PR hygiene
    testing.md          # How to write tests in this project
    migrations.md       # DB migration conventions
    api-design.md       # REST / GraphQL conventions
    debugging.md        # Step-by-step debugging protocol
  personas/
    reviewer.md         # Persona for code review tasks
    architect.md        # Persona for design / architecture tasks
  context/
    architecture.md     # High-level system design (keep up to date!)
    data-model.md       # Core entities and relationships
    glossary.md         # Domain terms and abbreviations
```

### 4.2 How Agents Should Use Skills

- **Claude Code:** Reference skills with `/skills .agents/skills/testing.md` or include them in the session via `@.agents/skills/testing.md`.
- **Cursor:** Add skills to the **Rules for AI** settings, or drag files into the composer context.
- **Copilot:** Include relevant skill files in the active editor context or reference them in the chat.
- **All agents:** When starting a non-trivial task, state which skill(s) you are following.

### 4.3 Updating Skills

Any agent (or human) that discovers a better approach should open a PR updating the relevant skill file. Skills are versioned with the codebase.

---

## 5. Plugin & Tool Conventions

### 5.1 Approved Tool Integrations

The following integrations are approved for use by agents in this repo. Agents must not invoke external APIs or services outside this list without human confirmation.

| Integration | Purpose | Auth method |
|---|---|---|
| GitHub API | Issues, PRs, Actions status | `GITHUB_TOKEN` env var |
| Internal DB | Schema introspection (read-only) | `DATABASE_URL` env var |
| *(add others)* | | |

### 5.2 Tool Call Etiquette

- Prefer **read operations** before write operations. Inspect before modifying.
- Log the intent of each significant tool call in a comment or reasoning step.
- If a tool call fails, retry **once** with a corrected input, then surface the error to the user rather than silently continuing.
- Never chain more than **5 write operations** without a human checkpoint.

### 5.3 File System Boundaries

```
✅ Agents may freely read/write:
   src/
   tests/
   docs/
   .agents/

⚠️  Agents must ask before modifying:
   package.json / pyproject.toml / go.mod
   *.config.ts / *.config.js
   .github/workflows/
   prisma/schema.prisma (or equivalent)
   Database migration files

🚫 Agents must never touch:
   .env  (or any .env.* file)
   Any file listed in .gitignore that contains credentials
   Compiled / generated output directories (dist/, build/, __pycache__/)
```

---

## 6. Multi-Agent Coordination

When more than one agent is active simultaneously (e.g. Claude Code running a refactor while Copilot handles inline suggestions):

1. **Declare scope.** At the start of a session, the primary agent should create a scratch file at `.agents/session/<timestamp>-<tool>.md` describing the task scope. This prevents collisions.
2. **Don't touch in-flight files.** If a file is open and being edited by another agent session, treat it as locked.
3. **Merge conflicts are human problems.** If a merge conflict results from two agent sessions, do not attempt to auto-resolve it. Surface it to the human.
4. **Shared memory.** Use `.agents/context/` as a shared scratchpad for notes, decisions, and discovered facts that should persist across sessions.

---

## 7. Escalation Policy

An agent **must stop and ask a human** before proceeding in any of the following situations:

- The task requires deleting more than 10 lines of existing logic.
- The task requires a new third-party dependency.
- The task touches authentication, authorisation, or encryption logic.
- The agent is not confident the change is correct (p < 0.8 internal estimate).
- Any tool call returns an unexpected error more than once.
- The instructions in this file conflict with a direct user instruction *(follow the user, but flag the conflict)*.

---

## 8. Versioning This File

| Version | Date | Author | Notes |
|---|---|---|---|
| 1.0 | <!-- date --> | <!-- author --> | Initial version |

> This file should be reviewed and updated whenever a new AI tool is added to the project, or when the MCP server list changes. Treat it as living documentation.
