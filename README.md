# ajolote

Shared AI agent configuration for multi-tool teams.

`ajolote` keeps a single canonical config (`.agents/config.json`) and translates it into the native format of each developer's AI coding assistant — Claude Code, Cursor, Windsurf, GitHub Copilot, Cline, and Aider. Tool-specific files are gitignored; only the shared config is committed.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh
```

## Usage

```sh
# One-time setup in any project (Python, Go, Java, Ruby — anything)
ajolote init

# After cloning a repo that already uses ajolote
ajolote sync

# Add a new tool mid-project
ajolote add cursor

# Check status
ajolote status
```

## How it works

| File | Committed? |
|---|---|
| `.agents/config.json` | ✅ Yes — canonical config |
| `.agents/skills/*.md` | ✅ Yes — human-authored |
| `CLAUDE.md`, `.claude/settings.json` | ❌ Gitignored — generated |
| `.cursor/rules/agents.mdc`, `.cursor/mcp.json` | ❌ Gitignored — generated |
| `.windsurf/rules/agents.md` | ❌ Gitignored — generated |
| `.github/copilot-instructions.md` | ❌ Gitignored — generated |
| `.clinerules`, `.roo/mcp.json` | ❌ Gitignored — generated |
| `.aider.conf.yml` | ❌ Gitignored — generated |

## Supported tools

- [Claude Code](https://claude.ai/code)
- [Cursor](https://cursor.com)
- [Windsurf](https://windsurf.com)
- [GitHub Copilot](https://github.com/features/copilot)
- [Cline / Roo](https://github.com/cline/cline)
- [Aider](https://aider.chat)
