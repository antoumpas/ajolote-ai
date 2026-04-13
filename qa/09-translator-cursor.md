# 09 — Cursor Translator

Tests for `ajolote use cursor` (generate) and Cursor import during `ajolote sync`.

---

## Generate

### CURSOR-001 — agents.mdc created with alwaysApply: true

**Prerequisites:** Initialized project with rules, skills, personas, context.  
**Steps:**
1. `ajolote use cursor`
2. `cat .cursor/rules/agents.mdc`

**Expected result:** File starts with YAML frontmatter containing `alwaysApply: true`. Body contains markdown bullet points referencing rule/skill/persona/context files.  
**Pass / Fail:** ☐

---

### CURSOR-002 — .cursor/mcp.json contains project-scoped servers

**Prerequisites:** Config has a project-scoped MCP server.  
**Steps:**
1. `ajolote use cursor`
2. `cat .cursor/mcp.json`

**Expected result:** Valid JSON with `mcpServers` key containing the project-scoped server. User-scoped servers absent.  
**Pass / Fail:** ☐

---

### CURSOR-003 — Commands written as .mdc files with alwaysApply: false

**Prerequisites:** `.agents/commands/deploy.md` exists.  
**Steps:**
1. `ajolote use cursor`
2. `cat .cursor/rules/deploy.mdc`

**Expected result:** File has YAML frontmatter with `alwaysApply: false` and `description:` field. Command content follows the frontmatter.  
**Pass / Fail:** ☐

---

### CURSOR-004 — Scoped rule written as .mdc with globs frontmatter

**Prerequisites:** Config has scoped rule `{"name": "frontend", "globs": ["**/*.tsx", "**/*.css"], "path": ".agents/rules/frontend.md"}`.  
**Steps:**
1. `ajolote use cursor`
2. `cat .cursor/rules/frontend.mdc`

**Expected result:** File has `globs: **/*.tsx, **/*.css` in frontmatter. Rule content from `.agents/rules/frontend.md` is in the body.  
**Pass / Fail:** ☐

---

### CURSOR-005 — User-scoped server written to ~/.cursor/mcp.json

**Prerequisites:** Config has a server with `"scope": "user"`.  
**Steps:**
1. `cp ~/.cursor/mcp.json ~/.cursor/mcp.json.bak 2>/dev/null || true`
2. `ajolote use cursor`
3. `cat ~/.cursor/mcp.json | grep "userserver"`

**Expected result:** User-scoped server appears in `~/.cursor/mcp.json`. NOT in `.cursor/mcp.json`.  
**Pass / Fail:** ☐

---

## Import

### CURSOR-006 — Import MCP servers from .cursor/mcp.json

**Prerequisites:** `.cursor/mcp.json` contains a server not in `.agents/config.json`.  
**Steps:**
1. `ajolote sync cursor`
2. `cat .agents/config.json | grep "newserver"`

**Expected result:** Server now in `config.json`.  
**Pass / Fail:** ☐

---

### CURSOR-007 — Import commands from .cursor/rules/ (skips agents.mdc and ajolote-sync.mdc)

**Prerequisites:** `.cursor/rules/` contains `deploy.mdc`, `agents.mdc`, and `ajolote-sync.mdc`.  
**Steps:**
1. `ajolote sync cursor`
2. `ls .agents/commands/`

**Expected result:** `.agents/commands/deploy.md` created. `agents` and `ajolote-sync` not imported.  
**Pass / Fail:** ☐

---

### CURSOR-008 — Import scoped rules from .mdc files with globs

**Prerequisites:** `.cursor/rules/backend.mdc` has `globs: **/*.go` in frontmatter.  
**Steps:**
1. `ajolote sync cursor`
2. `cat .agents/config.json | grep -A5 "scoped_rules"`

**Expected result:** `backend` scoped rule with `globs: ["**/*.go"]` added to config.  
**Pass / Fail:** ☐

---

### CURSOR-009 — Import agents.mdc as rule file (if not ajolote-generated)

**Prerequisites:** `.cursor/rules/agents.mdc` exists with hand-written content (no ajolote generation header).  
**Steps:**
1. `ajolote init` (fresh project with this file present)
2. `cat .agents/rules/general.md`

**Expected result:** Rule content from `agents.mdc` imported into `general.md`.  
**Pass / Fail:** ☐
