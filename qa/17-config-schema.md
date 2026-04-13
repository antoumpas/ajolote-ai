# 17 — Config JSON Schema

Tests for `.agents/config.json` parsing, serialization, persona formats, model aliases, and round-trip correctness.

---

### CONFIG-001 — Old-style personas (string list) parsed correctly

**Prerequisites:** `config.json` has `"personas": [".agents/personas/reviewer.md"]` (plain string, not object).  
**Steps:**
1. `ajolote use claude`
2. `cat CLAUDE.md | grep "personas"`

**Expected result:** Persona file reference appears in `CLAUDE.md`. No parse error.  
**Pass / Fail:** ☐

---

### CONFIG-002 — New-style personas (object with path + claude) parsed correctly

**Prerequisites:** `config.json` has:
```json
"personas": [{"path": ".agents/personas/reviewer.md", "claude": {"model": "sonnet", "tools": ["Read"]}}]
```
**Steps:**
1. `ajolote use claude`
2. `ls .claude/agents/`

**Expected result:** `.claude/agents/reviewer.md` created with correct frontmatter. No parse error.  
**Pass / Fail:** ☐

---

### CONFIG-003 — Mixed persona list (string + object) parsed correctly

**Prerequisites:** `config.json` has:
```json
"personas": [
  ".agents/personas/simple.md",
  {"path": ".agents/personas/rich.md", "claude": {"model": "opus", "tools": []}}
]
```
**Steps:**
1. `ajolote use claude`

**Expected result:** Both personas handled. Only `rich.md` gets a `.claude/agents/` file. `simple.md` reference appears in `CLAUDE.md`. No error.  
**Pass / Fail:** ☐

---

### CONFIG-004 — MCP server with all optional fields omitted

**Prerequisites:** `config.json` has `"minimal": {}` as a server (no command, transport, args, env, url, scope).  
**Steps:**
1. `ajolote use claude`
2. `cat .claude/settings.json`

**Expected result:** Server written to settings with empty/default fields. No crash.  
**Pass / Fail:** ☐

---

### CONFIG-005 — MCP server with transport: "http" + url

**Prerequisites:** `config.json` has `"remote": {"transport": "http", "url": "https://mcp.example.com/api"}`.  
**Steps:**
1. `ajolote use cursor`
2. `cat .cursor/mcp.json | grep "remote" -A5`

**Expected result:** Server written with `transport: "http"` and `url` fields.  
**Pass / Fail:** ☐

---

### CONFIG-006 — MCP server with scope: "user" parsed and routed correctly

**Prerequisites:** `config.json` has `"personal": {"command": "npx", "args": ["tool"], "scope": "user"}`.  
**Steps:**
1. `ajolote use claude`
2. `cat .claude/settings.json | grep "personal"`

**Expected result:** `"personal"` not in `.claude/settings.json`. It should only appear in `~/.claude.json`.  
**Pass / Fail:** ☐

---

### CONFIG-007 — Config save uses 2-space indent + trailing newline

**Prerequisites:** Initialized project; then `ajolote sync` to trigger a config re-save.  
**Steps:**
1. `cat -A .agents/config.json | tail -3` (shows trailing chars)

**Expected result:** File uses 2-space indentation. Final character is a newline. No trailing spaces.  
**Pass / Fail:** ☐

---

### CONFIG-008 — Config load → save round-trip preserves all fields

**Prerequisites:** Initialized project with a full config (rules, scoped_rules, skills, personas, context, MCP servers).  
**Steps:**
1. `cp .agents/config.json /tmp/before.json`
2. `ajolote sync` (triggers a reload + save)
3. `diff /tmp/before.json .agents/config.json`

**Expected result:** Files are identical (or differ only in whitespace normalization). No field lost or reordered unexpectedly.  
**Pass / Fail:** ☐

---

### CONFIG-009 — Missing config.json gives helpful error

**Prerequisites:** Directory with no `.agents/` folder.  
**Steps:**
1. `ajolote use claude`

**Expected result:** Error: `no ajolote config found at .agents/config.json — run 'ajolote init' first`. Exit non-zero.  
**Pass / Fail:** ☐

---

### CONFIG-010 — Malformed config.json gives parse error

**Prerequisites:** `.agents/config.json` contains invalid JSON (e.g., trailing comma).  
**Steps:**
1. `echo '{"mcp":{"servers":{},},}' > .agents/config.json`
2. `ajolote use claude`

**Expected result:** Error mentioning `parsing config:` with the parse error detail. No panic.  
**Pass / Fail:** ☐

---

### CONFIG-011 — Claude model alias "haiku" resolves correctly

**Prerequisites:** Persona with `"model": "haiku"`.  
**Steps:**
1. `ajolote use claude`
2. `grep "model:" .claude/agents/*.md`

**Expected result:** `model: claude-haiku-4-5-20251001`  
**Pass / Fail:** ☐

---

### CONFIG-012 — Claude model alias "sonnet" resolves correctly

**Prerequisites:** Persona with `"model": "sonnet"`.  
**Steps:**
1. `ajolote use claude`
2. `grep "model:" .claude/agents/*.md`

**Expected result:** `model: claude-sonnet-4-6`  
**Pass / Fail:** ☐

---

### CONFIG-013 — Claude model alias "opus" resolves correctly

**Prerequisites:** Persona with `"model": "opus"`.  
**Steps:**
1. `ajolote use claude`
2. `grep "model:" .claude/agents/*.md`

**Expected result:** `model: claude-opus-4-6`  
**Pass / Fail:** ☐

---

### CONFIG-014 — Unknown model string passed through unchanged

**Prerequisites:** Persona with `"model": "some-future-model-id"`.  
**Steps:**
1. `ajolote use claude`
2. `grep "model:" .claude/agents/*.md`

**Expected result:** `model: some-future-model-id` (unchanged, not an error).  
**Pass / Fail:** ☐
