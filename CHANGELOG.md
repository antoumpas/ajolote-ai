## v1.4.1 — 2026-04-16

- feat: write skills to .claude/skills/<name>/SKILL.md on generate (fe97c40)
- feat: import .claude/skills/ into .agents/skills/ on ajolote init/sync (3e55897)
- chore: update CHANGELOG for v1.4.0 (493cb08)

## v1.4.0 — 2026-04-15

- fix: use path.Match instead of filepath.Match in IsProtected for Windows compat (54e8ff0)
- fix: warn when config.local.json is malformed instead of silently ignoring (6bee182)
- qa: mark all 18 protection scenarios as PASS after manual execution (a737e95)
- qa: add scenarios for local file protection feature (d6b477b)
- feat: add .agents/config.local.json for local file protection (7eed888)
- chore: update CHANGELOG for v1.3.0 (26a383e)

## v1.3.0 — 2026-04-15

- feat: add ajolote scan command for secret and prompt-injection detection (2cf30be)
- chore: update CHANGELOG for v1.2.0 (9ee4051)

## v1.2.0 — 2026-04-15

- feat: add bidirectional import for windsurf scoped rules, commands, and codex MCP servers (bf0833c)
- chore: update CHANGELOG for v1.1.0 (61dc982)

## v1.1.0 — 2026-04-14

- feat: replace env var cache bypass with --refresh flag (d86dbf3)
- feat: add org-level config inheritance via extends field (b6d670f)
- chote: update gitignore (c489a93)
- chore: update CHANGELOG for v1.0.0 (fe7fa53)

## v1.0.0 — 2026-04-13

- fix: failing tests (fd0c81f)
- chore: update CHANGELOG for v0.7.3 (0047ff2)

## v0.7.3 — 2026-04-13

- security: fix 8 of 9 findings from security audit (SEC-001 through SEC-009) (fd0f0a3)
- chore: update CHANGELOG for v0.7.2 (2b5af9f)

## v0.7.2 — 2026-04-13

- fix: resolve 6 QA failures across cursor import, diff, status, env, and cline (75cae87)
- fix: use 4-backtick fence to fix unclosed code block breaking Supported tools section (20d7290)
- chore: update CHANGELOG for v0.7.1 (6a9d89b)

## v0.7.1 — 2026-04-13

- docs: add qa scenarios list (e8a1b40)
- feat: add --from flag to ajolote init for explicit tool import selection (4de2032)
- docs: document INSTALL_DIR override for install.sh in README (f9e31bb)
- fix: inject version at build time via ldflags instead of hardcoding (e992185)
- chore: update CHANGELOG for v0.7.0 (627c0c9)

## v0.7.0 — 2026-04-12

- fix: restore cwd after setupClaudeProject to prevent Windows CI failures (8054c6f)
- docs: document validate, diff commands and add per-tool sections for Gemini, Codex, AGENTS.md (47cdb6c)
- docs: update README to reflect all supported tools (9d137b8)
- feat: Windows support (94d359b)
- docs: document MCP env var substitution feature (7bb7a62)
- feat: expand ${ENV_VAR} placeholders in MCP server secrets at generation time (2ffbc82)
- feat: add agents-md translator for AGENTS.md standard (Linux Foundation / AAIF) (1a399e9)
- feat: add Gemini CLI and Codex CLI translators (68938e1)
- docs: restructure README for better adoption (8de2d29)
- feat: add ajolote diff command (fc6c939)
- feat: add ajolote validate command (6fad004)
- feat: generate .roomodes from personas for Roo Code custom modes (67c1745)
- feat: generate .claude/agents/ from personas with claude: block (baa7b2b)
- feat: add MCP server scoping (project vs user) and HTTP transport support (5e8bbc1)
- feat: add glob-scoped rules to config.json (50bf565)
- feat: per-tool content rendering strategy (inline vs @import) (7460c0e)
- chore: update CHANGELOG for v0.6.0 (670c99e)

## v0.6.0 — 2026-04-11

- Add license file (fe4fc0b)
- Replace structured project/rules objects with markdown file references (9dce954)
- chore: update CHANGELOG for v0.4.0 (fe48c68)

## v0.4.0 — 2026-04-10

- Add regression tests for Import-without-MCP-file bug (9c6c385)
- fix: import commands even when MCP config file is absent (071364a)
- chore: update CHANGELOG for v0.3.0 (a12c61f)

## v0.3.0 — 2026-04-10

- fix: run CHANGELOG update after GoReleaser, not before (fbf9616)
- chore: update CHANGELOG for v0.3.0 (4ffb020)
- init: detect existing tool configs and import MCP servers and commands (cb408fe)
- Add CHANGELOG.md auto-updated on every release (7af728d)

## v0.3.0 — 2026-04-10

- init: detect existing tool configs and import MCP servers and commands (cb408fe)
- Add CHANGELOG.md auto-updated on every release (7af728d)

## v0.2.0 — 2026-04-10

- Add personas and context to .agents/ scaffold (049c93d)
- Simplify init and add `ajolote use <tool>` command (c416f66)
- Add `ajolote sync` two-way sync command (06d983d)
- Generate slash command files per tool on `ajolote use <tool>` (8b990bd)
- Add .agents/commands/ support — translate team commands to each tool's native format (0de0168)
- Make sync two-way for commands — import tool command files back into .agents/commands/ (6b5a167)
- Add tests for commands translation and two-way import (f3cae14)

## v0.1.0 — 2026-04-10

- Initial release (5b7711c)
- Add .gitignore (ff5f414)
