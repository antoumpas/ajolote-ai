# 14 — Gemini CLI Translator

Tests for `ajolote use gemini` (generate) and Gemini import.

**Note:** Gemini uses the same @file reference format as Claude. It supports user-scoped MCP servers but does not support scoped rules or project-scoped MCP.

---

## Generate

### GEMINI-001 — GEMINI.md uses @file references

**Prerequisites:** Initialized project with rules, skills, personas, context.  
**Steps:**
1. `ajolote use gemini`
2. `cat GEMINI.md`

**Expected result:** `GEMINI.md` contains `@.agents/rules/general.md` (and other @file references). Content NOT inlined. Generation header present.  
**Pass / Fail:** ☐

---

### GEMINI-002 — No scoped rule output

**Prerequisites:** Config has a scoped rule.  
**Steps:**
1. `ajolote use gemini`
2. `ls -la | grep -i gemini`

**Expected result:** Only `GEMINI.md` is created. No scoped rule files generated for Gemini.  
**Pass / Fail:** ☐

---

### GEMINI-003 — User-scoped server written to ~/.gemini/settings.json

**Prerequisites:** Config has a server with `"scope": "user"`.  
**Steps:**
1. `cp ~/.gemini/settings.json ~/.gemini/settings.json.bak 2>/dev/null || true`
2. `ajolote use gemini`
3. `cat ~/.gemini/settings.json | grep "userserver"`

**Expected result:** User-scoped server written to `~/.gemini/settings.json`. Merged (not overwritten).  
**Pass / Fail:** ☐

---

### GEMINI-004 — User-scoped merge is idempotent

**Prerequisites:** A user-scoped server already in `~/.gemini/settings.json`.  
**Steps:**
1. `ajolote use gemini`
2. `ajolote use gemini` (again)
3. `cat ~/.gemini/settings.json | grep -c "userserver"`

**Expected result:** Server appears exactly once in the file. Not duplicated on second run.  
**Pass / Fail:** ☐

---

## Import

### GEMINI-005 — Import GEMINI.md as rule file (if not generated)

**Prerequisites:** `GEMINI.md` exists with hand-written content (no ajolote header).  
**Steps:**
1. `ajolote init` (fresh project)
2. `cat .agents/rules/general.md`

**Expected result:** Content of `GEMINI.md` imported into `general.md`.  
**Pass / Fail:** ☐

---

### GEMINI-006 — Skip ajolote-generated GEMINI.md

**Prerequisites:** `GEMINI.md` starts with the ajolote generation header.  
**Steps:**
1. `ajolote init` (fresh project)
2. `cat .agents/rules/general.md`

**Expected result:** `general.md` contains default boilerplate, not the generated `GEMINI.md` content.  
**Pass / Fail:** ☐
