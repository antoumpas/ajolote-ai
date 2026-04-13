# 04 — `ajolote sync`

Tests for `ajolote sync`: the two-phase import ↑ (read existing tool configs into `.agents/config.json`) and export ↓ (regenerate tool configs from updated config).

**Start each test** in a fresh initialized project unless stated otherwise.

---

### SYNC-001 — Sync with no tool output files present

**Prerequisites:** Initialized project; no `ajolote use` has been run (no `.claude/`, no `.cursor/`, etc.).  
**Steps:**
1. `ajolote sync`

**Expected result:** Command prints a message indicating no tool configs are found on disk. Exits with code 0 (not an error). `.agents/config.json` is unchanged.  
**Pass / Fail:** ☐

---

### SYNC-002 — Import new MCP server from Claude settings

**Prerequisites:** Initialized project; ran `ajolote use claude`; then manually add a new server to `.claude/settings.json`:
```json
{"mcpServers":{"newserver":{"command":"npx","args":["-y","some-package"]}}}
```
**Steps:**
1. `ajolote sync claude`
2. `cat .agents/config.json | grep "newserver"`

**Expected result:** `"newserver"` appears in `.agents/config.json` under `mcp.servers`. Sync output shows `↑ Imported 1 MCP server from claude`. Config file is saved.  
**Pass / Fail:** ☐

---

### SYNC-003 — Import new command from Claude commands directory

**Prerequisites:** Initialized project; ran `ajolote use claude`; then create `.claude/commands/release.md`:
```
---
description: Cut a release
---

Run ./scripts/release.sh
```
**Steps:**
1. `ajolote sync claude`
2. `ls .agents/commands/release.md`
3. `cat .agents/commands/release.md`

**Expected result:** `.agents/commands/release.md` exists with the imported content. Sync output shows the command was imported.  
**Pass / Fail:** ☐

---

### SYNC-004 — Import scoped rule from Cursor

**Prerequisites:** Initialized project; ran `ajolote use cursor`; then create `.cursor/rules/frontend.mdc`:
```
---
description: Frontend rules
globs: **/*.tsx, **/*.css
---

Use TypeScript strict mode.
```
**Steps:**
1. `ajolote sync cursor`
2. `cat .agents/config.json | grep -A5 "scoped_rules"`
3. `cat .agents/rules/frontend.md`

**Expected result:** `scoped_rules` array in config contains a `frontend` entry with `globs: ["**/*.tsx", "**/*.css"]`. `.agents/rules/frontend.md` exists with the rule content.  
**Pass / Fail:** ☐

---

### SYNC-005 — Auto-detect all tools with output files present

**Prerequisites:** Initialized project; ran `ajolote use claude` and `ajolote use cursor`.  
**Steps:**
1. `ajolote sync` (no tool argument)

**Expected result:** Both Claude and Cursor are processed (import + export). Output shows both tool names. No error for Windsurf, Copilot, etc. (they have no files on disk, so they are skipped).  
**Pass / Fail:** ☐

---

### SYNC-006 — Same MCP server not imported twice

**Prerequisites:** Initialized project with `"github"` server already in config; ran `ajolote use claude` (which wrote it to `.claude/settings.json`).  
**Steps:**
1. `ajolote sync claude`
2. `cat .agents/config.json | grep -c '"github"'`

**Expected result:** `"github"` appears exactly once in the config. Sync output reports nothing new was imported (0 new servers). Config is NOT dirtied/re-saved.  
**Pass / Fail:** ☐

---

### SYNC-007 — Export phase regenerates files after import

**Prerequisites:** Initialized project; ran `ajolote use claude`; then ran SYNC-002 (imported a new server).  
**Steps:**
1. After SYNC-002 completes, inspect `.claude/settings.json`.

**Expected result:** `.claude/settings.json` now contains the newly imported `"newserver"` entry — the export phase regenerated it from the updated config automatically within the same `ajolote sync` run.  
**Pass / Fail:** ☐

---

### SYNC-008 — Config not saved when nothing was imported

**Prerequisites:** Initialized project; ran `ajolote use claude`; no changes to tool outputs since last sync.  
**Steps:**
1. Note the modification time of `.agents/config.json`.
2. `ajolote sync claude`
3. Check modification time again.

**Expected result:** `.agents/config.json` modification time is unchanged (file not rewritten when content is unchanged).  
**Pass / Fail:** ☐

---

### SYNC-009 — Sync with Copilot (scoped rules imported, no MCP)

**Prerequisites:** Initialized project; ran `ajolote use copilot`; then create `.github/instructions/backend.instructions.md`:
```
---
applyTo: "**/*.go"
---

Follow Go idioms and error handling conventions.
```
**Steps:**
1. `ajolote sync copilot`
2. `cat .agents/config.json | grep -A5 "scoped_rules"`

**Expected result:** A `backend` scoped rule with `globs: ["**/*.go"]` is added to config. No MCP servers are imported (Copilot has no MCP config). No error.  
**Pass / Fail:** ☐

---

### SYNC-010 — Sync with Aider (export only, no import)

**Prerequisites:** Initialized project; ran `ajolote use aider`.  
**Steps:**
1. Modify `.agents/config.json` to add a new rule path (and create the file).
2. `ajolote sync aider`
3. `cat .aider.conf.yml`

**Expected result:** `.aider.conf.yml` is regenerated with the new rule path in the `read:` list. No import phase output for Aider. No error.  
**Pass / Fail:** ☐
