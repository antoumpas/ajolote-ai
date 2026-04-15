# 25 — Local File Protection (`.agents/config.local.json`)

Tests for the developer-local override file that prevents `ajolote use` and `ajolote sync`
from overwriting files the developer has manually customised.

**Key facts:**
- File lives at `.agents/config.local.json` — gitignored, never committed
- Protection is enforced silently: the write is skipped, original content is preserved
- Protected files show `⊘ <file> (protected)` (yellow) in output instead of `✔` / `↓`
- Absent file = no protection (default behaviour, zero regression risk)

---

### PROTECT-001 — No local config file → normal generation (baseline)

**Prerequisites:** Initialized project. `.agents/config.local.json` does NOT exist.  
**Steps:**
1. `ajolote use claude`
2. Verify `CLAUDE.md` is created and contains generated content.

**Expected result:** `CLAUDE.md` generated normally. No `⊘` symbols in output. Exit code `0`.  
**Pass / Fail:** ☐

---

### PROTECT-002 — Exact file match → file not overwritten by `use`

**Prerequisites:** Initialized project with `CLAUDE.md` already present with custom content.  
**Steps:**
1. `echo "# My personal notes" > CLAUDE.md`
2. `echo '{"protect":["CLAUDE.md"]}' > .agents/config.local.json`
3. `ajolote use claude`
4. `cat CLAUDE.md`

**Expected result:** `CLAUDE.md` still contains `# My personal notes`. The file was not overwritten. Exit code `0`.  
**Pass / Fail:** ☐

---

### PROTECT-003 — Exact file match → file not overwritten by `sync`

**Prerequisites:** Initialized project with cursor config present. Custom `.cursor/mcp.json` content.  
**Steps:**
1. `ajolote use cursor`
2. `echo '{"mcpServers":{"my-local-server":{"command":"my-tool"}}}' > .cursor/mcp.json`
3. `echo '{"protect":[".cursor/mcp.json"]}' > .agents/config.local.json`
4. `ajolote sync cursor`
5. `cat .cursor/mcp.json`

**Expected result:** `.cursor/mcp.json` still contains the custom local server entry. Exit code `0`.  
**Pass / Fail:** ☐

---

### PROTECT-004 — Protected file shows `⊘` in `use` output

**Prerequisites:** `.agents/config.local.json` with `"protect": ["CLAUDE.md"]`.  
**Steps:**
1. `ajolote use claude`
2. Observe the output.

**Expected result:** Output contains `⊘ CLAUDE.md (protected)` in yellow. Other files (`.claude/settings.json`, etc.) show `✔` normally.  
**Pass / Fail:** ☐

---

### PROTECT-005 — Protected file shows `⊘` in `sync` output

**Prerequisites:** `.agents/config.local.json` with `"protect": ["CLAUDE.md"]`. Claude config present on disk.  
**Steps:**
1. `ajolote sync claude`
2. Observe the `↓` / `⊘` lines in the output.

**Expected result:** `⊘ CLAUDE.md (protected)` shown under the `claude` section. Other files show `↓` normally.  
**Pass / Fail:** ☐

---

### PROTECT-006 — Unprotected files are still regenerated

**Prerequisites:** `.agents/config.local.json` with `"protect": ["CLAUDE.md"]`.  
**Steps:**
1. Delete `.claude/settings.json` if it exists.
2. `ajolote use claude`
3. Check that `.claude/settings.json` was created.

**Expected result:** `.claude/settings.json` is created/updated normally. Only `CLAUDE.md` is skipped.  
**Pass / Fail:** ☐

---

### PROTECT-007 — Glob pattern protects matching files

**Prerequisites:** Initialized project. Custom command file exists at `.claude/commands/my-cmd.md`.  
**Steps:**
1. `ajolote use claude`
2. `echo "# My custom command" > .claude/commands/my-cmd.md`
3. `echo '{"protect":[".claude/commands/*.md"]}' > .agents/config.local.json`
4. `ajolote use claude`
5. `cat .claude/commands/my-cmd.md`

**Expected result:** `.claude/commands/my-cmd.md` still contains `# My custom command`. The generated `ajolote-sync.md` is also protected by the glob and not regenerated from the canonical template.  
**Pass / Fail:** ☐

---

### PROTECT-008 — Glob does NOT match files in subdirectories (correct `*` semantics)

**Prerequisites:** `.agents/config.local.json` with `"protect": [".claude/commands/*.md"]`.  
**Steps:**
1. Create `.claude/commands/sub/nested.md` with custom content.
2. `ajolote use claude`
3. Verify `.claude/commands/sub/nested.md` content.

**Expected result:** The `*` glob does not cross directory boundaries. `.claude/commands/sub/nested.md` is NOT protected by `*.md` (it would need `**/*.md` or the directory prefix `".claude/commands/"`). File is overwritten/not protected.  
**Pass / Fail:** ☐

---

### PROTECT-009 — Directory prefix protects everything underneath

**Prerequisites:** Initialized project with Claude config.  
**Steps:**
1. `ajolote use claude`
2. `echo "custom content" > .claude/settings.json`
3. `echo '{"protect":[".claude/"]}' > .agents/config.local.json`
4. `ajolote use claude`
5. `cat .claude/settings.json`

**Expected result:** `.claude/settings.json`, `.claude/commands/ajolote-sync.md`, and any other file under `.claude/` are all preserved. All show `⊘` in output. Exit code `0`.  
**Pass / Fail:** ☐

---

### PROTECT-010 — Multiple files in protect list

**Prerequisites:** Initialized project with Claude config.  
**Steps:**
1. `ajolote use claude`
2. Set custom content in both `CLAUDE.md` and `.claude/settings.json`.
3. Create `.agents/config.local.json`:
   ```json
   {"protect": ["CLAUDE.md", ".claude/settings.json"]}
   ```
4. `ajolote use claude`
5. Verify both files are unchanged.

**Expected result:** Both protected files retain their custom content. Both show `⊘` in output. Exit code `0`.  
**Pass / Fail:** ☐

---

### PROTECT-011 — Protection works for non-Claude tools (cursor)

**Prerequisites:** Initialized project with cursor config.  
**Steps:**
1. `ajolote use cursor`
2. `echo "custom" > .cursor/mcp.json`
3. `echo '{"protect":[".cursor/mcp.json"]}' > .agents/config.local.json`
4. `ajolote use cursor`
5. Verify `.cursor/mcp.json` content.

**Expected result:** `.cursor/mcp.json` still contains `custom`. Demonstrates protection is tool-agnostic. Exit code `0`.  
**Pass / Fail:** ☐

---

### PROTECT-012 — Empty protect list → no files protected (all regenerated)

**Prerequisites:** Initialized project.  
**Steps:**
1. `echo '{"protect":[]}' > .agents/config.local.json`
2. `ajolote use claude`
3. Observe output.

**Expected result:** All files show `✔` normally. No `⊘` symbols. Behaviour identical to having no local config file. Exit code `0`.  
**Pass / Fail:** ☐

---

### PROTECT-013 — Malformed `config.local.json` → graceful failure

**Prerequisites:** Initialized project.  
**Steps:**
1. `echo "not valid json {{{" > .agents/config.local.json`
2. `ajolote use claude`
3. `echo "Exit: $?"`

**Expected result:** Command either exits with a clear error message referencing `config.local.json`, or falls back to no-protection behaviour and completes. Must not panic or produce a cryptic error. Exit code communicates whether generation succeeded.  
**Pass / Fail:** ☐

---

### PROTECT-014 — `ajolote init` adds `config.local.json` to `.gitignore`

**Prerequisites:** Empty project directory with no `.gitignore`.  
**Steps:**
1. `ajolote init`
2. `cat .gitignore`
3. Search for `.agents/config.local.json` in the output.

**Expected result:** `.agents/config.local.json` appears inside the `# <ajolote-ai>` managed block of `.gitignore`. The entry is present even before the file itself is created.  
**Pass / Fail:** ☐

---

### PROTECT-015 — `config.local.json` is not committed by mistake

**Prerequisites:** Git repository with an initialized project. `.agents/config.local.json` created with a protect entry.  
**Steps:**
1. `git status`
2. `git add -A`
3. `git status`

**Expected result:** `.agents/config.local.json` does NOT appear as a file to be committed — it is recognised as gitignored. `git status` shows it as ignored (or absent from the tracked/untracked list).  
**Pass / Fail:** ☐

---

### PROTECT-016 — Re-running `ajolote init` on existing project preserves `config.local.json` in gitignore

**Prerequisites:** Initialized project. `.gitignore` already contains the ajolote block.  
**Steps:**
1. Manually remove `.agents/config.local.json` from `.gitignore`.
2. `ajolote init` (should skip because config already exists).
3. Alternatively, call `ajolote ignore` or note that `init` won't re-run.

**Expected result:** (Documents known behaviour) `init` skips if config exists. Teams should run `ajolote init` fresh or manually ensure the entry is present. This is a known limitation — no auto-repair outside of `init`.  
**Pass / Fail:** ☐

---

### PROTECT-017 — Protected file that doesn't exist yet is created on first run

**Prerequisites:** Initialized project. `CLAUDE.md` does NOT exist yet.  
**Steps:**
1. `echo '{"protect":["CLAUDE.md"]}' > .agents/config.local.json`
2. `ajolote use claude`
3. Check whether `CLAUDE.md` was created.

**Expected result:** Because the file doesn't exist, the write is still skipped (protection applies regardless). `CLAUDE.md` is NOT created. Output shows `⊘ CLAUDE.md (protected)`. The developer must create it manually if they want it.  
**Pass / Fail:** ☐

---

### PROTECT-018 — `sync` import phase is not affected by protection

**Prerequisites:** Project with cursor configured. `.agents/config.local.json` protects `.cursor/mcp.json`.  
**Steps:**
1. Add a new MCP server directly in Cursor's UI (or manually in `.cursor/mcp.json`).
2. `ajolote sync cursor`
3. Check `.agents/config.json`.

**Expected result:** The `↑` import phase still runs — new MCP servers discovered by `sync` are imported into `.agents/config.json` correctly. Protection only blocks the `↓` export/write phase, not the import phase. Exit code `0`.  
**Pass / Fail:** ☐
