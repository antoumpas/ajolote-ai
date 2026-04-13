# 21 — Cross-Platform

Tests verifying consistent behavior across macOS (arm64 + amd64), Linux (arm64 + amd64), and Windows (amd64). Each test should be run on each platform where applicable.

**Platform key:** 🍎 macOS · 🐧 Linux · 🪟 Windows

---

### XPLAT-001 — Full flow on macOS arm64 🍎

**Prerequisites:** macOS arm64; `ajolote` installed.  
**Steps:**
1. `mkdir ~/xplat-test && cd ~/xplat-test && git init`
2. `ajolote init`
3. `ajolote use claude`
4. `ajolote diff claude`
5. `ajolote validate`

**Expected result:** All commands complete without error. `ajolote diff` exits 0. `ajolote validate` passes.  
**Pass / Fail:** ☐

---

### XPLAT-002 — Full flow on Linux amd64 🐧

**Prerequisites:** Linux amd64 (or container).  
**Steps:** Same as XPLAT-001.

**Expected result:** Same as XPLAT-001.  
**Pass / Fail:** ☐

---

### XPLAT-003 — Full flow on Windows amd64 🪟

**Prerequisites:** Windows 10/11 amd64; `ajolote.exe` installed.  
**Steps:**
1. Create a temp directory, `cd` into it, `git init`.
2. `ajolote init`
3. `ajolote use claude`
4. `ajolote diff claude`
5. `ajolote validate`

**Expected result:** All commands complete without error. `ajolote diff` exits 0. No backslash paths in output.  
**Pass / Fail:** ☐

---

### XPLAT-004 — `ajolote diff` output uses forward slashes on Windows 🪟

**Prerequisites:** Windows; initialized project; `ajolote use claude` run; then modify config to trigger a diff.  
**Steps:**
1. Add a rule path to config.
2. `ajolote diff claude`
3. Inspect all file path references in the output.

**Expected result:** Paths shown as `.claude/settings.json` (forward slash), NOT `.claude\settings.json`. No backslashes in file path strings.  
**Pass / Fail:** ☐

---

### XPLAT-005 — User home directory resolved correctly on Windows 🪟

**Prerequisites:** Windows; config has a user-scoped MCP server.  
**Steps:**
1. `ajolote use claude`
2. `Get-Item "$env:USERPROFILE\.claude.json"`

**Expected result:** `~/.claude.json` is created at `%USERPROFILE%\.claude.json` (e.g., `C:\Users\alex\.claude.json`).  
**Pass / Fail:** ☐

---

### XPLAT-006 — User-scoped MCP written to correct Windows home path 🪟

**Prerequisites:** Windows; config has a server with `"scope": "user"`.  
**Steps:**
1. `ajolote use cursor`
2. Check `$env:USERPROFILE\.cursor\mcp.json`.

**Expected result:** File exists at the Windows user home path and contains the user-scoped server.  
**Pass / Fail:** ☐

---

### XPLAT-007 — Path separators in generated config content are forward slashes (all platforms) 🍎🐧🪟

**Prerequisites:** Any OS; initialized project with rules.  
**Steps:**
1. `ajolote use claude`
2. `cat CLAUDE.md`
3. `cat .claude/settings.json`

**Expected result:** All paths in file content use forward slashes (e.g., `@.agents/rules/general.md`). No OS-native backslashes in generated file content.  
**Pass / Fail:** ☐

---

### XPLAT-008 — Git Bash on Windows redirects to PowerShell 🪟

**Prerequisites:** Windows with Git Bash installed.  
**Steps:**
1. Open Git Bash.
2. `curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh`

**Expected result:** Script detects Windows and prints: `Windows detected. Use the PowerShell installer instead:` with the `irm ... | iex` command. Exits 0. Does not attempt to download or install.  
**Pass / Fail:** ☐

---

### XPLAT-009 — Go integration tests pass on Windows 🪟

**Prerequisites:** Windows with Go installed; repo cloned.  
**Steps:**
1. `go test ./...`

**Expected result:** All tests pass. No `getwd: no such file or directory` errors. No test failures caused by temp-dir cleanup ordering.  
**Pass / Fail:** ☐

---

### XPLAT-010 — File line endings in generated files (all platforms) 🍎🐧🪟

**Prerequisites:** Any OS; `ajolote use claude` run.  
**Steps:**
1. `file CLAUDE.md` or use a hex editor to check line endings.

**Expected result:** Generated files use LF (`\n`) line endings on all platforms (not CRLF on Windows). This ensures consistent git diffs for committed files like `AGENTS.md`.  
**Pass / Fail:** ☐

---

### XPLAT-011 — .gitignore written with correct line endings 🪟

**Prerequisites:** Windows; `ajolote init` run.  
**Steps:**
1. Open `.gitignore` in a hex editor or `Format-Hex .gitignore | Select-String "0D"`.

**Expected result:** `.gitignore` uses LF line endings (no `0x0D` / `\r`). Git on Windows handles this correctly when `core.autocrlf` is configured.  
**Pass / Fail:** ☐
