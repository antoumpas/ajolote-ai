# 12 — Cline / Roo Code Translator

Tests for `ajolote use cline` (generate) and Cline import.

---

## Generate

### CLINE-001 — .clinerules created with inlined content

**Prerequisites:** Initialized project with rules, skills, personas, context.  
**Steps:**
1. `ajolote use cline`
2. `cat .clinerules`

**Expected result:** `.clinerules` exists with the ajolote generation header. All rule/skill/persona/context file contents inlined directly (not @file references).  
**Pass / Fail:** ☐

---

### CLINE-002 — .roo/mcp.json contains project-scoped servers

**Prerequisites:** Config has a project-scoped MCP server.  
**Steps:**
1. `ajolote use cline`
2. `cat .roo/mcp.json`

**Expected result:** Valid JSON with `mcpServers` key. Project-scoped server present. User-scoped server absent.  
**Pass / Fail:** ☐

---

### CLINE-003 — Commands written to .roo/rules/ with # heading format

**Prerequisites:** `.agents/commands/deploy.md` exists.  
**Steps:**
1. `ajolote use cline`
2. `cat .roo/rules/deploy.md`

**Expected result:** File starts with `# deploy` heading. Command content follows. `ajolote-sync.md` also present in `.roo/rules/`.  
**Pass / Fail:** ☐

---

### CLINE-004 — .roomodes JSON created with one mode per persona

**Prerequisites:** Config has two personas.  
**Steps:**
1. `ajolote use cline`
2. `cat .roomodes`

**Expected result:** Valid JSON with `customModes` array. One entry per persona, each with `slug`, `name`, `roleDefinition` (inlined persona content), `groups`, and `source: "project"` fields.  
**Pass / Fail:** ☐

---

### CLINE-005 — No personas → no .roomodes file

**Prerequisites:** Initialized project with empty `personas` array in config.  
**Steps:**
1. `ajolote use cline`
2. `ls .roomodes 2>/dev/null || echo "not found"`

**Expected result:** `.roomodes` file does NOT exist.  
**Pass / Fail:** ☐

---

## Import

### CLINE-006 — Import MCP servers from .roo/mcp.json

**Prerequisites:** `.roo/mcp.json` contains a server not in config.  
**Steps:**
1. `ajolote sync cline`
2. `cat .agents/config.json | grep "clineserver"`

**Expected result:** Server appears in `config.json`.  
**Pass / Fail:** ☐

---

### CLINE-007 — Import commands from .roo/rules/ (skips ajolote-sync)

**Prerequisites:** `.roo/rules/` contains `build.md` (content: `# build\nRun make.`) and `ajolote-sync.md`.  
**Steps:**
1. `ajolote sync cline`
2. `ls .agents/commands/`

**Expected result:** `.agents/commands/build.md` created. `ajolote-sync` not imported.  
**Pass / Fail:** ☐

---

### CLINE-008 — Import .clinerules as rule file (if not ajolote-generated)

**Prerequisites:** `.clinerules` exists with hand-written content (no ajolote header).  
**Steps:**
1. `ajolote init` (fresh project)
2. `cat .agents/rules/general.md`

**Expected result:** Content from `.clinerules` imported into `general.md`.  
**Pass / Fail:** ☐

---

### CLINE-009 — Persona roleDefinition inlines the persona file content

**Prerequisites:** `.agents/personas/reviewer.md` contains multi-paragraph content.  
**Steps:**
1. `ajolote use cline`
2. `cat .roomodes | python3 -m json.tool | grep -A5 '"roleDefinition"'`

**Expected result:** The `roleDefinition` field contains the full text of the persona file, not a file reference.  
**Pass / Fail:** ☐
