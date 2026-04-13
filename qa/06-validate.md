# 06 — `ajolote validate`

Tests for `ajolote validate`: pre-sync checks that catch configuration errors before they reach `ajolote sync` or CI.

**Exit codes:** `0` = no errors (warnings are OK); `1` = at least one error.

---

### VALIDATE-001 — All files valid, clean pass

**Prerequisites:** Initialized project; all files referenced in config exist and are non-empty.  
**Steps:**
1. `ajolote validate`
2. `echo "Exit: $?"`

**Expected result:** Output shows `✔` for every file and MCP server. Final line: `All checks passed.` Exit code `0`.  
**Pass / Fail:** ☐

---

### VALIDATE-002 — Missing rule file

**Prerequisites:** Initialized project; `config.json` references `.agents/rules/missing.md`; file does NOT exist.  
**Steps:**
1. `ajolote validate`
2. `echo "Exit: $?"`

**Expected result:** Output shows `✘ .agents/rules/missing.md — file not found`. Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-003 — Empty rule file (0 bytes)

**Prerequisites:** Initialized project; `.agents/rules/general.md` exists but is empty (0 bytes).  
**Steps:**
1. `ajolote validate`
2. `echo "Exit: $?"`

**Expected result:** Output shows `✘ .agents/rules/general.md — file is empty`. Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-004 — Whitespace-only rule file

**Prerequisites:** Initialized project; `.agents/rules/general.md` contains only spaces, tabs, and newlines.  
**Steps:**
1. `printf "   \n\n  \t\n" > .agents/rules/general.md`
2. `ajolote validate`

**Expected result:** Output shows `✘ .agents/rules/general.md — file is empty` (or equivalent whitespace-only message). Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-005 — Missing skill file

**Prerequisites:** Initialized project; `config.json` has a skill path referencing a file that does not exist.  
**Steps:**
1. Remove `.agents/skills/git.md` (or set a non-existent path in config).
2. `ajolote validate`

**Expected result:** Error shown under Skills section. Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-006 — Missing persona file

**Prerequisites:** Initialized project; `config.json` references a persona file that does not exist.  
**Steps:**
1. Set a non-existent persona path in config.
2. `ajolote validate`

**Expected result:** Error shown under Personas section. Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-007 — Missing context file

**Prerequisites:** Initialized project; `config.json` references a context file that does not exist.  
**Steps:**
1. Set a non-existent context path in config.
2. `ajolote validate`

**Expected result:** Error shown under Context section. Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-008 — Scoped rule with no glob patterns

**Prerequisites:** Initialized project; `config.json` contains a scoped rule with `"globs": null` or `"globs": []`.  
**Steps:**
1. `ajolote validate`

**Expected result:** Output shows `✘ <name> — scoped rule has no glob patterns`. Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-009 — Scoped rule with valid globs passes

**Prerequisites:** Initialized project; scoped rule with `"globs": ["**/*.tsx"]` and the rule file exists with content.  
**Steps:**
1. `ajolote validate`

**Expected result:** Scoped rule shows `✔ <name>` in output. Exit code `0` (assuming no other errors).  
**Pass / Fail:** ☐

---

### VALIDATE-010 — MCP stdio server with no command

**Prerequisites:** Config contains `"broken": {"transport": "stdio"}` (no command field).  
**Steps:**
1. `ajolote validate`

**Expected result:** Output shows `✘ broken — stdio server requires a command`. Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-011 — MCP stdio server command in PATH (pass)

**Prerequisites:** Config contains `"shell": {"command": "sh", "args": ["-c", "echo hi"]}`.  
**Steps:**
1. `ajolote validate`

**Expected result:** Output shows `✔ shell (sh)` under MCP Servers. Exit code `0`.  
**Pass / Fail:** ☐

---

### VALIDATE-012 — MCP stdio server command NOT in PATH (warning only)

**Prerequisites:** Config contains `"missing": {"command": "thisdoesnotexist123"}`.  
**Steps:**
1. `ajolote validate`
2. `echo "Exit: $?"`

**Expected result:** Output shows `⚠ missing — command 'thisdoesnotexist123' not found in PATH` (yellow warning). Exit code `0` (warning, not error).  
**Pass / Fail:** ☐

---

### VALIDATE-013 — MCP http server with no URL

**Prerequisites:** Config contains `"remote": {"transport": "http"}` (no url field).  
**Steps:**
1. `ajolote validate`

**Expected result:** Output shows `✘ remote — transport "http" requires a url`. Exit code `1`.  
**Pass / Fail:** ☐

---

### VALIDATE-014 — MCP http server with valid URL

**Prerequisites:** Config contains `"remote": {"transport": "http", "url": "https://mcp.example.com/api"}`.  
**Steps:**
1. `ajolote validate`

**Expected result:** Output shows `✔ remote (http)`. Exit code `0`.  
**Pass / Fail:** ☐

---

### VALIDATE-015 — Multiple errors all listed together

**Prerequisites:** Config with: a missing rule file, a scoped rule with no globs, and an MCP http server with no URL — all three issues present simultaneously.  
**Steps:**
1. `ajolote validate`
2. `echo "Exit: $?"`

**Expected result:** All three errors listed in the output (under their respective sections). Final line: `N error(s)`. Exit code `1`.  
**Pass / Fail:** ☐
