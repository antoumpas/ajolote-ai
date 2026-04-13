# 18 — Environment Variable Substitution

Tests for `${VAR}` and `$VAR` expansion in MCP server `url` and `env` fields. The feature allows teams to commit their config without leaking secrets.

**Key behavior:** If a variable is set, it is substituted. If it is NOT set, the placeholder is left as-is (e.g., `${MY_TOKEN}` stays as `${MY_TOKEN}` — a visible reminder, not an empty string).

---

### ENV-001 — ${VAR} in MCP url substituted when env var is set

**Prerequisites:** Config has `"remote": {"transport": "http", "url": "${MY_MCP_URL}"}`.  
**Steps:**
1. `export MY_MCP_URL="https://mcp.example.com/api"`
2. `ajolote use claude`
3. `cat .claude/settings.json | grep "url"`

**Expected result:** `"url": "https://mcp.example.com/api"` — variable expanded.  
**Pass / Fail:** ☐

---

### ENV-002 — Unset ${VAR} in MCP url left as placeholder

**Prerequisites:** Config has `"remote": {"transport": "http", "url": "${MY_MCP_URL}"}`. `MY_MCP_URL` is NOT set.  
**Steps:**
1. `unset MY_MCP_URL`
2. `ajolote use claude`
3. `cat .claude/settings.json | grep "url"`

**Expected result:** `"url": "${MY_MCP_URL}"` — placeholder preserved, not expanded to empty string.  
**Pass / Fail:** ☐

---

### ENV-003 — ${VAR} in MCP env value substituted

**Prerequisites:** Config has `"server": {"command": "npx", "args": ["pkg"], "env": {"API_KEY": "${MY_API_KEY}"}}`.  
**Steps:**
1. `export MY_API_KEY="super-secret-token"`
2. `ajolote use claude`
3. `cat .claude/settings.json | grep "API_KEY"`

**Expected result:** `"API_KEY": "super-secret-token"` — value substituted.  
**Pass / Fail:** ☐

---

### ENV-004 — Unset ${VAR} in MCP env value left as placeholder

**Prerequisites:** Same config as ENV-003. `MY_API_KEY` is NOT set.  
**Steps:**
1. `unset MY_API_KEY`
2. `ajolote use claude`
3. `cat .claude/settings.json | grep "API_KEY"`

**Expected result:** `"API_KEY": "${MY_API_KEY}"` — placeholder preserved.  
**Pass / Fail:** ☐

---

### ENV-005 — Multiple ${VAR} in same string all expanded

**Prerequisites:** Config has `"url": "${PROTOCOL}://${HOST}/api"`. Both vars are set.  
**Steps:**
1. `export PROTOCOL="https"; export HOST="mcp.example.com"`
2. `ajolote use claude`
3. `cat .claude/settings.json | grep "url"`

**Expected result:** `"url": "https://mcp.example.com/api"` — both variables expanded in one string.  
**Pass / Fail:** ☐

---

### ENV-006 — $VAR (no braces) also substituted

**Prerequisites:** Config has `"url": "$MY_MCP_URL"` (no curly braces). Var is set.  
**Steps:**
1. `export MY_MCP_URL="https://example.com"`
2. `ajolote use claude`
3. `cat .claude/settings.json | grep "url"`

**Expected result:** `"url": "https://example.com"` — bare `$VAR` format also expanded.  
**Pass / Fail:** ☐

---

### ENV-007 — Env var set to empty string substitutes as empty

**Prerequisites:** Config has `"url": "${EMPTY_VAR}"`. Var is set to `""`.  
**Steps:**
1. `export EMPTY_VAR=""`
2. `ajolote use claude`
3. `cat .claude/settings.json | grep "url"`

**Expected result:** `"url": ""` — empty string substituted.  
**Pass / Fail:** ☐

---

### ENV-008 — Substitution applied in Cursor mcp.json

**Prerequisites:** Config has a server with `"env": {"TOKEN": "${MY_TOKEN}"}`. Var is set.  
**Steps:**
1. `export MY_TOKEN="cursor-token"`
2. `ajolote use cursor`
3. `cat .cursor/mcp.json | grep "TOKEN"`

**Expected result:** `"TOKEN": "cursor-token"` in `.cursor/mcp.json`.  
**Pass / Fail:** ☐

---

### ENV-009 — Substitution applied in Codex config.toml

**Prerequisites:** Config has `"url": "${CODEX_URL}"` and `"env": {"KEY": "${CODEX_KEY}"}`. Both vars set.  
**Steps:**
1. `export CODEX_URL="https://codex.example.com"; export CODEX_KEY="key123"`
2. `ajolote use codex`
3. `cat .codex/config.toml`

**Expected result:** `url = "https://codex.example.com"` and `KEY = "key123"` in TOML. Vars substituted.  
**Pass / Fail:** ☐

---

### ENV-010 — Substitution applied in Cline .roo/mcp.json

**Prerequisites:** Config has a server with an env var in its `env` field. Var is set.  
**Steps:**
1. Set the env var.
2. `ajolote use cline`
3. `cat .roo/mcp.json | grep "KEY"`

**Expected result:** Value substituted in `.roo/mcp.json`.  
**Pass / Fail:** ☐

---

### ENV-011 — Env key itself is never expanded

**Prerequisites:** Config has `"env": {"${SHOULD_NOT_EXPAND}": "value"}`.  
**Steps:**
1. `export SHOULD_NOT_EXPAND="REAL_KEY"`
2. `ajolote use claude`
3. `cat .claude/settings.json`

**Expected result:** The key remains literally `${SHOULD_NOT_EXPAND}` — only values are expanded, not keys.  
**Pass / Fail:** ☐
