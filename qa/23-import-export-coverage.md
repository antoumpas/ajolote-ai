# QA 23 — Import/Export Coverage

Tests for complete bidirectional sync coverage for Windsurf and Codex CLI translators.

## Prerequisites

- `ajolote` binary built (`make build` or `go install`)
- A terminal with a project root set up via `ajolote init`

---

## Windsurf — Scoped Rule Import

### IMPORT-001 — Windsurf scoped rule survives sync round-trip

**Steps:**
1. `mkdir /tmp/ws-test && cd /tmp/ws-test && ajolote init`
2. `ajolote use windsurf`
3. Manually create `.windsurf/rules/typescript.md`:
   ```
   ---
   globs: **/*.ts, **/*.tsx
   ---

   # TypeScript Rules
   Always use strict mode.
   ```
4. `ajolote sync windsurf`
5. `cat .agents/config.json | grep typescript`

**Expected result:**
- `.agents/config.json` gains a `scoped_rules` entry with `name: "typescript"` and `globs: ["**/*.ts", "**/*.tsx"]`
- `.agents/rules/typescript.md` is created with body `# TypeScript Rules\nAlways use strict mode.`

**Pass / Fail:** ✅

---

### IMPORT-002 — Windsurf scoped rule file with no `globs:` field is NOT imported as scoped rule

**Steps:**
1. `cd /tmp/ws-test`
2. Create `.windsurf/rules/readme.md` with plain content (no frontmatter):
   ```
   # Dev Notes
   Some developer context notes.
   ```
3. `ajolote sync windsurf`
4. `cat .agents/config.json`

**Expected result:**
- `scoped_rules` does NOT contain a `readme` entry (no `globs:` frontmatter)
- Command exits 0

**Pass / Fail:** ✅

---

### IMPORT-003 — Generated `agents.md` is not re-imported

**Steps:**
1. `cd /tmp/ws-test && ajolote use windsurf`
2. `ajolote sync windsurf`
3. `cat .agents/config.json`

**Expected result:**
- `.agents/config.json` is unchanged (no new rule files imported)
- `agents.md` is recognized as ajolote-generated and skipped

**Pass / Fail:** ✅

---

## Windsurf — Command Import

### IMPORT-004 — User-authored workflow imported as command

**Steps:**
1. `cd /tmp/ws-test`
2. Create `.windsurf/workflows/deploy.yaml`:
   ```yaml
   name: deploy
   description: Deploy to production
   steps:
     - name: Execute
       say: |
         Run ./scripts/deploy.sh
         Verify health check passes.
   ```
3. `ajolote sync windsurf`
4. `cat .agents/commands/deploy.md`

**Expected result:**
- `.agents/commands/deploy.md` is created
- The file contains the description from the workflow
- File body includes `Run ./scripts/deploy.sh`

**Pass / Fail:** ✅

---

### IMPORT-005 — `ajolote-sync.yaml` is not imported as a command

**Steps:**
1. `cd /tmp/ws-test && ajolote use windsurf`
   (This generates `.windsurf/workflows/ajolote-sync.yaml`)
2. `ajolote sync windsurf`
3. `ls .agents/commands/`

**Expected result:**
- No `ajolote-sync.md` appears in `.agents/commands/`

**Pass / Fail:** ✅

---

### IMPORT-006 — Already-imported command is not duplicated on second sync

**Steps:**
1. From IMPORT-004 setup, run `ajolote sync windsurf` a second time
2. `ls .agents/commands/`

**Expected result:**
- Still only one `deploy.md` in `.agents/commands/`
- No duplicate or error

**Pass / Fail:** ✅

---

## Codex CLI — MCP Server Import

### IMPORT-007 — Codex MCP server added to config.toml appears after sync

**Steps:**
1. `mkdir /tmp/codex-test && cd /tmp/codex-test && ajolote init && ajolote use codex`
2. Append a new server block to `.codex/config.toml`:
   ```toml

   [mcp.servers.shell]
   command = "npx"
   args = ["-y", "@modelcontextprotocol/server-shell"]
   ```
3. `ajolote sync codex`
4. `cat .agents/config.json | grep shell`

**Expected result:**
- `.agents/config.json` gains `shell` under `mcp.servers`
- `command` is `"npx"`, `args` contains `"-y"` and `"@modelcontextprotocol/server-shell"`
- Command exits 0

**Pass / Fail:** ✅

---

### IMPORT-008 — Codex MCP server with env vars correctly imported

**Steps:**
1. `cd /tmp/codex-test`
2. Add a server with env block to `.codex/config.toml`:
   ```toml

   [mcp.servers.gh]
   command = "npx"
   args = ["-y", "@modelcontextprotocol/server-github"]

   [mcp.servers.gh.env]
   GITHUB_TOKEN = "ghp_test123"
   ```
3. `ajolote sync codex`
4. `cat .agents/config.json`

**Expected result:**
- `gh` server appears under `mcp.servers` in config.json
- `env.GITHUB_TOKEN` is `"ghp_test123"`

**Pass / Fail:** ✅

---

### IMPORT-009 — MCP server already in config.json is not duplicated

**Steps:**
1. From IMPORT-007 setup, run `ajolote sync codex` a second time
2. Count occurrences of `"shell"` in `.agents/config.json`

**Expected result:**
- Only one `shell` entry in `mcp.servers` — no duplication

**Pass / Fail:** ✅

---

## Regression

### IMPORT-010 — `ajolote sync` with no windsurf or codex files exits cleanly

**Steps:**
1. `mkdir /tmp/plain-test && cd /tmp/plain-test && ajolote init`
2. `ajolote use claude`
3. `ajolote sync`

**Expected result:**
- Command completes without error
- No mention of windsurf or codex in output (they have no files on disk)

**Pass / Fail:** ✅

---

### IMPORT-011 — Empty `.codex/config.toml` (header only) does not import stale servers

**Steps:**
1. `cd /tmp/codex-test && ajolote use codex`
   (Generates a fresh `.codex/config.toml` with no servers if config.json has none)
2. Remove all MCP servers from `.agents/config.json`
3. `ajolote sync codex`

**Expected result:**
- No servers added back from the TOML (it's empty / header-only)
- `.agents/config.json` stays clean

**Pass / Fail:** ✅
