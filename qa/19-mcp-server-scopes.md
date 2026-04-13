# 19 — MCP Server Scopes

Tests for project-scoped vs user-scoped MCP server routing. Project-scoped servers go into committed tool configs. User-scoped servers go into the developer's home directory and are never committed or diffed.

**Setup note:** Back up home-directory MCP files before these tests:
```sh
cp ~/.claude.json ~/.claude.json.bak 2>/dev/null || true
cp ~/.cursor/mcp.json ~/.cursor/mcp.json.bak 2>/dev/null || true
cp ~/.gemini/settings.json ~/.gemini/settings.json.bak 2>/dev/null || true
```

---

### SCOPE-001 — Project-scoped server (default) written to .claude/settings.json

**Prerequisites:** Config has `"github": {"command": "npx", "args": ["pkg"]}` (no scope field — defaults to project).  
**Steps:**
1. `ajolote use claude`
2. `cat .claude/settings.json | grep "github"`

**Expected result:** `"github"` server present in `.claude/settings.json`.  
**Pass / Fail:** ☐

---

### SCOPE-002 — Project-scoped server written to .cursor/mcp.json

**Prerequisites:** Same config as SCOPE-001.  
**Steps:**
1. `ajolote use cursor`
2. `cat .cursor/mcp.json | grep "github"`

**Expected result:** `"github"` server present in `.cursor/mcp.json`.  
**Pass / Fail:** ☐

---

### SCOPE-003 — User-scoped server NOT in .claude/settings.json

**Prerequisites:** Config has `"personal": {"command": "npx", "args": ["my-tool"], "scope": "user"}`.  
**Steps:**
1. `ajolote use claude`
2. `cat .claude/settings.json | grep "personal"`

**Expected result:** No output. `"personal"` not in `.claude/settings.json`.  
**Pass / Fail:** ☐

---

### SCOPE-004 — User-scoped server written to ~/.claude.json

**Prerequisites:** Same config as SCOPE-003.  
**Steps:**
1. `ajolote use claude`
2. `cat ~/.claude.json | grep "personal"`

**Expected result:** `"personal"` server appears in `~/.claude.json`.  
**Pass / Fail:** ☐

---

### SCOPE-005 — User-scoped server written to ~/.cursor/mcp.json

**Prerequisites:** Config has a user-scoped server.  
**Steps:**
1. `ajolote use cursor`
2. `cat ~/.cursor/mcp.json | grep "personal"`

**Expected result:** `"personal"` server appears in `~/.cursor/mcp.json`.  
**Pass / Fail:** ☐

---

### SCOPE-006 — User-scoped server written to ~/.gemini/settings.json

**Prerequisites:** Config has a user-scoped server.  
**Steps:**
1. `ajolote use gemini`
2. `cat ~/.gemini/settings.json | grep "personal"`

**Expected result:** `"personal"` server appears in `~/.gemini/settings.json`.  
**Pass / Fail:** ☐

---

### SCOPE-007 — Home file merge preserves existing entries

**Prerequisites:** `~/.claude.json` already has `"existing-server"` entry (from another project or tool).  
**Steps:**
1. `ajolote use claude` (config has a user-scoped server)
2. `cat ~/.claude.json | grep "existing-server"`

**Expected result:** `"existing-server"` still present in `~/.claude.json`. Existing entries not overwritten or removed.  
**Pass / Fail:** ☐

---

### SCOPE-008 — User-scoped merge is idempotent

**Prerequisites:** Config has a user-scoped server. Run `ajolote use claude` once.  
**Steps:**
1. `ajolote use claude` (second run)
2. `cat ~/.claude.json | grep -c '"personal"'`

**Expected result:** Server appears exactly once. Not duplicated on second run.  
**Pass / Fail:** ☐

---

### SCOPE-009 — `ajolote diff` excludes user-scoped servers

**Prerequisites:** Config has both a project-scoped and a user-scoped server. `ajolote use claude` has been run.  
**Steps:**
1. Change the user-scoped server's command in config.
2. `ajolote diff claude`

**Expected result:** The diff does NOT include changes to `~/.claude.json` (outside project). Diff output only references project files. Exit code reflects only project-file changes.  
**Pass / Fail:** ☐
