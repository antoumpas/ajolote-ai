# 05 — `ajolote diff`

Tests for `ajolote diff`: previewing what `ajolote sync` would change without writing anything. Critical for CI pipelines.

**Exit codes:** `0` = nothing would change; `1` = at least one file would be created or modified.

---

### DIFF-001 — Diff with no tool output files on disk

**Prerequisites:** Initialized project; `ajolote use` has NOT been run.  
**Steps:**
1. `ajolote diff`
2. `echo "Exit: $?"`

**Expected result:** Output: `No tool configs found on disk. Run 'ajolote use <tool>' first.` Exit code `0`.  
**Pass / Fail:** ☐

---

### DIFF-002 — Diff immediately after `ajolote use` (no changes)

**Prerequisites:** Initialized project; `ajolote use claude` just completed successfully.  
**Steps:**
1. `ajolote diff claude`
2. `echo "Exit: $?"`

**Expected result:** Output shows `✔ CLAUDE.md`, `✔ .claude/settings.json`, etc. Final line: `Nothing would change.` Exit code `0`.  
**Pass / Fail:** ☐

---

### DIFF-003 — Diff detects added rule path in config

**Prerequisites:** Initialized project; `ajolote use claude` run. Then:
- Add `.agents/rules/code-style.md` to `rules` array in config.
- Create the file with content `# Code Style`.

**Steps:**
1. `ajolote diff claude`
2. `echo "Exit: $?"`

**Expected result:** Output shows `~ CLAUDE.md` with a unified diff. The diff shows `+@.agents/rules/code-style.md`. Final line indicates 1 file would change. Exit code `1`.  
**Pass / Fail:** ☐

---

### DIFF-004 — Diff shows unified diff with 3-line context

**Prerequisites:** Initialized project; `ajolote use claude` run; then modify config to add a new rule path so CLAUDE.md would change.  
**Steps:**
1. `ajolote diff claude`
2. Inspect the diff output.

**Expected result:** Changed lines are shown with `+`/`-` prefix. Lines immediately before and after the change are shown as context (space prefix). At most 3 context lines on each side. Hunk header `@@ -N,M +N,M @@` is present.  
**Pass / Fail:** ☐

---

### DIFF-005 — Diff for new file (not yet on disk)

**Prerequisites:** Initialized project; ran `ajolote use claude` (so Claude is active). Then add a command file `.agents/commands/review.md` to the project.  
**Steps:**
1. `ajolote diff claude`

**Expected result:** `.claude/commands/review.md` shown with `+` (cyan) marker. Full file content displayed as additions. Exit code `1`.  
**Pass / Fail:** ☐

---

### DIFF-006 — Diff single tool only

**Prerequisites:** Initialized project; `ajolote use claude` and `ajolote use cursor` both run; then modify config to add a new rule path.  
**Steps:**
1. `ajolote diff claude`

**Expected result:** Only Claude files are checked and diffed. Cursor files are not mentioned in the output. Exit code reflects only Claude changes.  
**Pass / Fail:** ☐

---

### DIFF-007 — Diff with user-scoped MCP server excluded

**Prerequisites:** Initialized project; config has one project-scoped and one user-scoped (`"scope": "user"`) MCP server; `ajolote use claude` run.  
**Steps:**
1. `ajolote diff claude`

**Expected result:** The user-scoped server does NOT appear in the diff output (it writes to `~/.claude.json`, which is outside the project). Only project-scoped content is compared.  
**Pass / Fail:** ☐

---

### DIFF-008 — Diff output uses forward slashes on all platforms

**Prerequisites:** Any OS; initialized project; `ajolote use claude` run.  
**Steps:**
1. `ajolote diff claude`
2. Inspect all file path references in the output (e.g., `.claude/settings.json`).

**Expected result:** All paths shown with forward slashes (`/`), even on Windows. No backslashes in file paths in the output.  
**Pass / Fail:** ☐

---

### DIFF-009 — Exit code 1 fails CI pipeline

**Prerequisites:** GitHub Actions workflow with the CI config from `.github/workflows/ci.yml`; a branch where config has drifted (tool output not committed, config changed since last `ajolote use`).  
**Steps:**
1. Push a commit where `.agents/config.json` was changed but generated files were not updated.
2. Observe the CI run.

**Expected result:** The CI step running `ajolote diff` (or equivalent) exits 1, causing the job to fail. The failure message references the diff output.  
**Pass / Fail:** ☐

---

### DIFF-010 — Diff all tools (no argument)

**Prerequisites:** Initialized project; `ajolote use claude` and `ajolote use cursor` run; nothing changed in config.  
**Steps:**
1. `ajolote diff` (no argument)
2. `echo "Exit: $?"`

**Expected result:** Both Claude and Cursor tool sections shown with `✔` markers for all files. Final line: `Nothing would change.` Exit code `0`.  
**Pass / Fail:** ☐
