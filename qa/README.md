# ajolote-ai QA Manual Testing Suite — v1.0.0

This directory contains the complete manual QA test suite for ajolote-ai. Each file covers one logical test area. Work through each file top-to-bottom, following the steps exactly and recording the result.

---

## How to Use This Suite

1. **Copy the file** you are testing to a dated working copy (e.g., `02-init_2026-04-12_AT.md`) so you can check boxes without modifying the originals.
2. **Set up the environment** (see below) before starting any file.
3. Work through each test case in order. Mark `☐` as `☑ PASS` or `☒ FAIL`.
4. For failures, note the actual output beneath the expected result.
5. File a bug report referencing the **Test ID** (e.g., `INIT-003`).

---

## Environment Setup

Before running any test:

```sh
# 1. Install ajolote (or build from source)
go install github.com/ajolote-ai/ajolote/cmd/ajolote@latest
# or: build locally
go build -o ajolote ./cmd/ajolote && mv ajolote /usr/local/bin/

# 2. Confirm version
ajolote --help

# 3. Create a clean scratch directory for each test (do not reuse)
mkdir ~/ajolote-test && cd ~/ajolote-test && git init

# 4. For tests that write to home directory (~/.claude.json, ~/.cursor/mcp.json, ~/.gemini/settings.json)
# Backup any existing files first:
cp ~/.claude.json ~/.claude.json.bak 2>/dev/null || true
```

**Required tools on PATH:**
- `git` (for init to detect git repo)
- `go` (if building from source)
- `curl` (for installer tests on macOS/Linux)
- PowerShell 5.1+ (for installer tests on Windows)

---

## Test ID Convention

| Prefix | File | Area |
|--------|------|------|
| `SETUP` | 01-setup-and-installation.md | Installation |
| `INIT` | 02-init.md | `ajolote init` |
| `USE` | 03-use.md | `ajolote use` |
| `SYNC` | 04-sync.md | `ajolote sync` |
| `DIFF` | 05-diff.md | `ajolote diff` |
| `VALIDATE` | 06-validate.md | `ajolote validate` |
| `STATUS` | 07-status-and-ignore.md | `ajolote status` |
| `IGNORE` | 07-status-and-ignore.md | `ajolote ignore` |
| `CLAUDE` | 08-translator-claude.md | Claude translator |
| `CURSOR` | 09-translator-cursor.md | Cursor translator |
| `WINDSURF` | 10-translator-windsurf.md | Windsurf translator |
| `COPILOT` | 11-translator-copilot.md | Copilot translator |
| `CLINE` | 12-translator-cline.md | Cline translator |
| `AIDER` | 13-translator-aider.md | Aider translator |
| `GEMINI` | 14-translator-gemini.md | Gemini translator |
| `CODEX` | 15-translator-codex.md | Codex translator |
| `AGENTSMD` | 16-translator-agents-md.md | agents-md translator |
| `CONFIG` | 17-config-schema.md | Config JSON schema |
| `ENV` | 18-env-var-substitution.md | Env var substitution |
| `SCOPE` | 19-mcp-server-scopes.md | MCP server scopes |
| `GITIGNORE` | 20-gitignore-management.md | .gitignore management |
| `XPLAT` | 21-cross-platform.md | Cross-platform |

---

## File Index

| File | Area | # Tests |
|------|------|---------|
| [01-setup-and-installation.md](01-setup-and-installation.md) | Installer scripts (macOS, Linux, Windows) | 9 |
| [02-init.md](02-init.md) | `ajolote init` — scaffolding and import | 10 |
| [03-use.md](03-use.md) | `ajolote use <tool>` — file generation | 7 |
| [04-sync.md](04-sync.md) | `ajolote sync` — import ↑ and export ↓ | 10 |
| [05-diff.md](05-diff.md) | `ajolote diff` — change preview and CI | 10 |
| [06-validate.md](06-validate.md) | `ajolote validate` — pre-sync checks | 15 |
| [07-status-and-ignore.md](07-status-and-ignore.md) | `ajolote status` + `ajolote ignore` | 6 |
| [08-translator-claude.md](08-translator-claude.md) | Claude — generate + import | 14 |
| [09-translator-cursor.md](09-translator-cursor.md) | Cursor — generate + import | 9 |
| [10-translator-windsurf.md](10-translator-windsurf.md) | Windsurf — generate + import | 6 |
| [11-translator-copilot.md](11-translator-copilot.md) | Copilot — generate + import | 8 |
| [12-translator-cline.md](12-translator-cline.md) | Cline/Roo — generate + import | 9 |
| [13-translator-aider.md](13-translator-aider.md) | Aider — generate | 4 |
| [14-translator-gemini.md](14-translator-gemini.md) | Gemini CLI — generate + import | 6 |
| [15-translator-codex.md](15-translator-codex.md) | Codex CLI — generate + import | 7 |
| [16-translator-agents-md.md](16-translator-agents-md.md) | agents-md — committed output | 7 |
| [17-config-schema.md](17-config-schema.md) | config.json schema and round-trips | 14 |
| [18-env-var-substitution.md](18-env-var-substitution.md) | `${VAR}` expansion in MCP fields | 11 |
| [19-mcp-server-scopes.md](19-mcp-server-scopes.md) | Project vs user-scoped MCP servers | 9 |
| [20-gitignore-management.md](20-gitignore-management.md) | Managed `.gitignore` block | 9 |
| [21-cross-platform.md](21-cross-platform.md) | macOS / Linux / Windows parity | 11 |
