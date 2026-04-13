# 20 — Gitignore Management

Tests for the managed `.gitignore` block that ajolote maintains. The block is delimited by `# <ajolote-ai>` and `# </ajolote-ai>` markers and is idempotent.

---

### GITIGNORE-001 — .gitignore created when it does not exist

**Prerequisites:** Fresh git repo; no `.gitignore` file.  
**Steps:**
1. `ajolote init`
2. `ls .gitignore`
3. `cat .gitignore`

**Expected result:** `.gitignore` created. Contains the managed block with all translator output paths (except `AGENTS.md`).  
**Pass / Fail:** ☐

---

### GITIGNORE-002 — Block appended to existing .gitignore without block

**Prerequisites:** `.gitignore` exists with entries (`node_modules/`, `dist/`) but NO ajolote block.  
**Steps:**
1. `ajolote init`
2. `cat .gitignore`

**Expected result:** `node_modules/` and `dist/` still present. Managed block appended after existing content. Existing content untouched.  
**Pass / Fail:** ☐

---

### GITIGNORE-003 — Block replaced in-place on subsequent init/ignore

**Prerequisites:** `.gitignore` with a complete ajolote block.  
**Steps:**
1. `ajolote ignore`
2. `cat .gitignore | grep -c "ajolote-ai"`

**Expected result:** Exactly 2 occurrences of `ajolote-ai` (open + close markers). Block not duplicated.  
**Pass / Fail:** ☐

---

### GITIGNORE-004 — External entries outside block never modified

**Prerequisites:** `.gitignore` with entries before AND after the managed block.  
**Steps:**
1. `echo "# custom before" >> /tmp/gi_header.txt`
2. Set up file with entries before block, block, entries after block.
3. `ajolote ignore`
4. Inspect the file.

**Expected result:** Entries before and after the block are unchanged. Only block content is replaced.  
**Pass / Fail:** ☐

---

### GITIGNORE-005 — AGENTS.md not in managed block

**Prerequisites:** Fresh git repo.  
**Steps:**
1. `ajolote init`
2. `grep "AGENTS.md" .gitignore`

**Expected result:** No output. `AGENTS.md` must not appear in the managed block.  
**Pass / Fail:** ☐

---

### GITIGNORE-006 — All 9 non-committed translator output paths in block

**Prerequisites:** Fresh git repo after `ajolote init`.  
**Steps:**
1. `cat .gitignore`

**Expected result:** Block contains paths for all translators except agents-md. At minimum:
- `CLAUDE.md`
- `.claude/settings.json`
- `.claude/commands/`
- `.cursor/mcp.json`
- `.cursor/rules/`
- `.windsurf/`
- `.github/copilot-instructions.md`
- `.github/instructions/`
- `.clinerules`
- `.roo/`
- `.roomodes`
- `.aider.conf.yml`
- `GEMINI.md`
- `.codex/`  

**Pass / Fail:** ☐

---

### GITIGNORE-007 — `ajolote ignore` refreshes stale block

**Prerequisites:** Initialized project; managed block is present but missing a translator path (manually deleted).  
**Steps:**
1. Remove one line from the block.
2. `ajolote ignore`
3. `cat .gitignore`

**Expected result:** Removed line restored. Block is fully rebuilt with all current paths.  
**Pass / Fail:** ☐

---

### GITIGNORE-008 — `ajolote ignore` is idempotent

**Prerequisites:** Initialized project with a valid, complete managed block.  
**Steps:**
1. `cat .gitignore > /tmp/gi_before.txt`
2. `ajolote ignore`
3. `diff /tmp/gi_before.txt .gitignore`

**Expected result:** `diff` produces no output.  
**Pass / Fail:** ☐

---

### GITIGNORE-009 — Block markers are exact (not fuzzy)

**Prerequisites:** `.gitignore` with a slightly modified open marker (e.g., `# <ajolote-ai> ` with trailing space).  
**Steps:**
1. `ajolote ignore`
2. `cat .gitignore`

**Expected result:** A NEW block is appended (marker not recognized as existing block). Behavior documents the strict marker requirement. This is expected — the modified marker is treated as a comment, not an ajolote block.  
**Pass / Fail:** ☐
