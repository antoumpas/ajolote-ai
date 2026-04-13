# 07 — `ajolote status` and `ajolote ignore`

Tests for the `ajolote status` command (which tools are ready locally) and `ajolote ignore` (rebuild the managed `.gitignore` block).

---

## `ajolote status`

### STATUS-001 — No tools generated yet

**Prerequisites:** Initialized project; `ajolote use` has NOT been run for any tool.  
**Steps:**
1. `ajolote status`

**Expected result:** All 9 tools listed. Each shows `○` (yellow) with a hint like `run ajolote use <tool>`. None show `✔`.  
**Pass / Fail:** ☐

---

### STATUS-002 — One tool generated

**Prerequisites:** Initialized project; `ajolote use claude` run.  
**Steps:**
1. `ajolote status`

**Expected result:** `claude` shows `✔` (green). All other tools show `○`. Exactly one tool is active.  
**Pass / Fail:** ☐

---

### STATUS-003 — All tools generated

**Prerequisites:** Initialized project; `ajolote use <tool>` run for all 9 tools.  
**Steps:**
1. `ajolote status`

**Expected result:** All 9 tools show `✔` (green). No `○` markers.  
**Pass / Fail:** ☐

---

## `ajolote ignore`

### IGNORE-001 — Rebuilds .gitignore block

**Prerequisites:** Initialized project with `.gitignore`; block may be stale (e.g., missing some entries).  
**Steps:**
1. Manually delete a line from the `# <ajolote-ai>` block in `.gitignore`.
2. `ajolote ignore`
3. `cat .gitignore`

**Expected result:** The managed block is fully rebuilt with all current translator output paths. The deleted line is restored.  
**Pass / Fail:** ☐

---

### IGNORE-002 — AGENTS.md not added by `ajolote ignore`

**Prerequisites:** Initialized project.  
**Steps:**
1. `ajolote ignore`
2. `grep "AGENTS.md" .gitignore`

**Expected result:** No output (AGENTS.md not in `.gitignore`). The agents-md translator has `CommittedOutput` and must be excluded.  
**Pass / Fail:** ☐

---

### IGNORE-003 — `ajolote ignore` is idempotent

**Prerequisites:** Initialized project with a properly formed `.gitignore` block.  
**Steps:**
1. `cat .gitignore > /tmp/before.txt`
2. `ajolote ignore`
3. `cat .gitignore > /tmp/after.txt`
4. `diff /tmp/before.txt /tmp/after.txt`

**Expected result:** `diff` produces no output — the file is byte-for-byte identical after a redundant `ajolote ignore` call.  
**Pass / Fail:** ☐
