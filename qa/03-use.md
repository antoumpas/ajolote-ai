# 03 — `ajolote use <tool>`

Tests for the `ajolote use` command: generating local tool-specific config files from `.agents/config.json`.

**Start each test** in a project that has already been initialized (`ajolote init`), unless the test explicitly requires otherwise.

---

### USE-001 — Error when config does not exist

**Prerequisites:** Empty directory; no `.agents/config.json`.  
**Steps:**
1. `ajolote use claude`

**Expected result:** Error message: `no ajolote config found at .agents/config.json — run 'ajolote init' first`. Exit code non-zero.  
**Pass / Fail:** ☐

---

### USE-002 — Error for unknown tool name

**Prerequisites:** Initialized project.  
**Steps:**
1. `ajolote use notarealtool`

**Expected result:** Error message listing valid tool names (claude, cursor, windsurf, copilot, cline, aider, gemini, codex, agents-md). Exit code non-zero.  
**Pass / Fail:** ☐

---

### USE-003 — `ajolote use claude` creates output files

**Prerequisites:** Initialized project with at least one rule file.  
**Steps:**
1. `ajolote use claude`
2. `ls CLAUDE.md .claude/settings.json .claude/commands/`

**Expected result:** `CLAUDE.md`, `.claude/settings.json`, and `.claude/commands/ajolote-sync.md` all exist. Command prints the list of generated files.  
**Pass / Fail:** ☐

---

### USE-004 — `ajolote use` is idempotent

**Prerequisites:** Initialized project.  
**Steps:**
1. `ajolote use claude`
2. Record the content of `CLAUDE.md`.
3. `ajolote use claude` (run again)
4. Compare content of `CLAUDE.md`.

**Expected result:** Second run produces identical output. No error. No duplicate content.  
**Pass / Fail:** ☐

---

### USE-005 — `ajolote use claude` regenerates after config change

**Prerequisites:** Initialized project; already ran `ajolote use claude`.  
**Steps:**
1. Add a new rule path to `.agents/config.json`: `".agents/rules/code-style.md"`.
2. Create `.agents/rules/code-style.md` with content `# Code Style`.
3. `ajolote use claude`
4. `cat CLAUDE.md`

**Expected result:** `CLAUDE.md` now contains `@.agents/rules/code-style.md`. The new rule path appears.  
**Pass / Fail:** ☐

---

### USE-006 — `ajolote use` for all 9 supported tools

**Prerequisites:** Initialized project with at least one rule and one MCP server in config.  
**Steps (repeat for each tool name):**
1. `ajolote use claude` → check `CLAUDE.md` and `.claude/settings.json` exist
2. `ajolote use cursor` → check `.cursor/rules/agents.mdc` and `.cursor/mcp.json` exist
3. `ajolote use windsurf` → check `.windsurf/rules/agents.md` exists
4. `ajolote use copilot` → check `.github/copilot-instructions.md` exists
5. `ajolote use cline` → check `.clinerules` and `.roo/mcp.json` exist
6. `ajolote use aider` → check `.aider.conf.yml` exists
7. `ajolote use gemini` → check `GEMINI.md` exists
8. `ajolote use codex` → check `.codex/config.toml` exists
9. `ajolote use agents-md` → check `AGENTS.md` exists

**Expected result:** Each command succeeds and creates the expected files. No errors for any tool.  
**Pass / Fail:** ☐

---

### USE-007 — Minimal config (no MCP servers) generates valid files

**Prerequisites:** Initialized project; config has rules but `"servers": {}` (empty MCP).  
**Steps:**
1. `ajolote use claude`
2. `cat .claude/settings.json`
3. `ajolote use cline`
4. `cat .roo/mcp.json`

**Expected result:** `.claude/settings.json` contains `{"mcpServers":{}}`. `.roo/mcp.json` contains `{"mcpServers":{}}`. No errors. Files are valid JSON.  
**Pass / Fail:** ☐
