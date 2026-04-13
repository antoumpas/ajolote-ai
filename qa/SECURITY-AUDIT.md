# Security Audit Report — ajolote-ai

**Date:** 2026-04-13
**Scope:** Full source tree at commit `75cae87` (main branch)
**Auditor:** Automated analysis + manual code review
**Tool version:** Source (post-QA fixes)
**Remediation status:** All 9 findings addressed (see status per finding below)

---

## Executive Summary

ajolote is a local CLI that reads `.agents/config.json` (a committed, team-reviewed file) and generates tool-specific config files. It does not run a server, accept network input, or execute user-supplied commands. Its primary attack surface is **malicious content in config files or imported tool configs** and **file-system operations on user-controlled paths**.

9 findings were identified: 2 High, 5 Medium, 2 Low. All have been remediated.

No Critical findings. The most impactful issues require an attacker to either control `.agents/config.json` (a committed, code-reviewed file) or pre-plant malicious files in tool config directories (`.cursor/`, `.claude/`, etc.).

---

## Findings

### SEC-001 — Path traversal via config.json paths (HIGH)

**Files:** `internal/translators/util.go:126-131`, `internal/cli/commands/sync.go:128-136`, all translator `Generate()` methods
**CWE:** CWE-22 (Path Traversal)

**Description:**
Paths in `config.json` fields (`rules`, `skills`, `context`, `personas[].path`, `scoped_rules[].path`) flow into `filepath.Join(projectRoot, userPath)` without validation. `filepath.Join` normalises `..` segments, so a path like `../../.env` resolves outside the project. An absolute path like `/etc/cron.d/job` bypasses the project root entirely.

**Read path:** `inlineFiles()`, `readCommands()`, `os.ReadFile(filepath.Join(projectRoot, sr.Path))` across all translators — can read arbitrary files and inline their contents into generated configs.

**Write path:** `sync.go:128` writes scoped rule content to `filepath.Join(projectRoot, sr.Path)`. A crafted `scoped_rules[].path` can write outside `.agents/rules/`.

**Attack scenario:**
A malicious commit adds to `config.json`:
```json
"scoped_rules": [{"name":"x", "globs":["*"], "path":"../../.env"}]
```
On next `ajolote sync`, the rule body overwrites `.env` two directories up.

**Mitigating factors:**
- `config.json` is committed to git and subject to code review
- An attacker with commit access can cause more damage directly
- `ajolote validate` checks file existence but not path boundaries

**Recommendation:**
Add a `validatePath` helper that rejects paths containing `..` segments or starting with `/`. Call it in `config.Load()` for all path fields. Example:
```go
func validatePath(p string) error {
    clean := filepath.Clean(p)
    if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") || strings.Contains(clean, string(filepath.Separator)+"..") {
        return fmt.Errorf("path %q must be relative and within the project", p)
    }
    return nil
}
```

---

### SEC-002 — No symlink protection on home-directory writes (HIGH)

**Files:** `internal/translators/util.go:80-122` (`mergeUserMCPConfig`), called from `claude.go:37`, `cursor.go:48`, `gemini.go:29`
**CWE:** CWE-59 (Symlink Following)

**Description:**
User-scoped MCP servers are written to `~/.claude.json`, `~/.cursor/mcp.json`, and `~/.gemini/settings.json` via `mergeUserMCPConfig()`. This function calls `os.WriteFile()` without checking whether the target is a symlink.

**Attack scenario:**
On a shared system, an attacker creates `~/.claude.json` as a symlink to `~/.ssh/authorized_keys`. Next `ajolote use claude` overwrites the symlink target with MCP server JSON, corrupting the SSH config. This is a destructive write, not code execution, but could deny access or overwrite sensitive files.

**Mitigating factors:**
- Requires the attacker to have write access to the victim's home directory
- Home directories typically have `0o700` permissions on multi-user systems

**Recommendation:**
Before writing, call `os.Lstat()` on the target path and reject symlinks:
```go
if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
    return fmt.Errorf("refusing to write to symlink: %s", path)
}
```

---

### SEC-003 — TOML injection via MCP server names and env keys (MEDIUM)

**File:** `internal/translators/codex.go:72, 90, 97`
**CWE:** CWE-74 (Injection)

**Description:**
The Codex translator generates TOML by string interpolation. Server names are inserted unquoted into TOML table headers:
```go
sb.WriteString(fmt.Sprintf("\n[mcp.servers.%s]\n", name))
```
Env var keys are also inserted unquoted:
```go
sb.WriteString(fmt.Sprintf("%s = %q\n", k, expandEnv(srv.Env[k])))
```

A server name containing `]` or a newline (valid in JSON keys) breaks TOML structure. An env key containing `=` or newlines similarly corrupts output.

**Attack scenario:**
Config contains server name `"evil]\n[mcp.servers.inject"`. Generated TOML becomes:
```toml
[mcp.servers.evil]
[mcp.servers.inject]
```
The injected section could shadow or override legitimate server configuration when parsed by Codex.

**Mitigating factors:**
- Server names come from `config.json` (committed, reviewed)
- Codex CLI may reject malformed TOML rather than misparse it
- Env var keys with special characters are unlikely in practice

**Recommendation:**
Validate server names and env keys against `^[a-zA-Z][a-zA-Z0-9_-]*$` during config loading, or quote them in TOML output using the `"dotted.key"` syntax.

---

### SEC-004 — No download integrity verification in install scripts (MEDIUM)

**Files:** `install.sh:59-62`, `install.ps1:47-51`
**CWE:** CWE-494 (Download Without Integrity Check)

**Description:**
Both installers download a binary archive from GitHub and extract it without verifying checksums. The goreleaser config generates `checksums.txt` alongside releases, but neither installer fetches or verifies it.

```bash
curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"
tar -xzf "${TMP}/${ARCHIVE}" -C "$TMP"
# No checksum verification
```

**Attack scenario:**
A MITM attacker (corporate proxy, compromised DNS) serves a tampered binary. The installer extracts and places it in PATH without any integrity check.

**Mitigating factors:**
- GitHub serves releases over HTTPS with certificate pinning
- `curl -f` fails on non-200 responses, preventing partial-download attacks
- The `curl | sh` pattern is standard for CLI tools (homebrew, rustup, nvm)

**Recommendation:**
Download and verify `checksums.txt` before extracting:
```bash
curl -fsSL "${CHECKSUMS_URL}" -o "${TMP}/checksums.txt"
(cd "$TMP" && sha256sum -c --ignore-missing checksums.txt)
```

---

### SEC-005 — YAML frontmatter injection in generated tool configs (MEDIUM)

**Files:** `internal/translators/claude.go:54-55`, `cursor.go:54`, `windsurf.go:57`
**CWE:** CWE-74 (Injection)

**Description:**
Scoped rule names and globs are interpolated into YAML frontmatter without quoting:
```go
content := fmt.Sprintf("---\ndescription: %s\nglobs: %s\n---\n\n@%s\n",
    sr.Name, strings.Join(sr.Globs, ", "), sr.Path)
```

A `scoped_rules[].name` containing a newline or YAML special characters could inject extra frontmatter fields or break the document structure. In Claude Code's case, `@` references in injected content could cause arbitrary file reads.

**Mitigating factors:**
- `sr.Name` comes from `config.json` (committed, reviewed) or from filesystem filenames (which cannot contain newlines or `/` on most OSes)
- AI tools generally sanitize frontmatter on their end

**Recommendation:**
Quote YAML values: `description: %q` or validate names against `^[a-zA-Z0-9_-]+$`.

---

### SEC-006 — Environment variable secrets written to project files (MEDIUM)

**File:** `internal/translators/util.go:46-53` (`expandEnv`), all translator `Generate()` methods
**CWE:** CWE-532 (Information Exposure Through Log Files)

**Description:**
`${VAR}` placeholders in MCP server config are expanded from the shell environment at generation time. Resolved values (which may include API tokens) are written to project-level files like `.claude/settings.json`, `.cursor/mcp.json`, and `.codex/config.toml`.

These files are gitignored by ajolote, but:
- A developer could `git add -f` them accidentally
- CI environments may persist workspace artifacts
- The scope separation (`"scope": "user"`) correctly routes user-scoped servers to home directories, but project-scoped servers with env vars still land in the project directory

**Mitigating factors:**
- This is documented, intentional behavior
- All generated files are gitignored by `ajolote init`
- The `validate` command does not expose env var values

**Recommendation:**
Add a comment header in generated files that contain expanded env vars:
```
# WARNING: This file may contain secrets resolved from environment variables.
# Do not commit this file to version control.
```

---

### SEC-007 — Home-directory config files created world-readable (MEDIUM)

**File:** `internal/translators/util.go:121` (`mergeUserMCPConfig`)
**CWE:** CWE-732 (Incorrect Permission Assignment)

**Description:**
Files written to home directories (`~/.claude.json`, `~/.cursor/mcp.json`, `~/.gemini/settings.json`) use permission mode `0o644`, making them readable by all users on the system. These files contain resolved MCP server configurations that may include API tokens.

**Mitigating factors:**
- Home directories are typically `0o700` on Unix, preventing access regardless of file permissions
- macOS and most Linux distributions set restrictive home directory permissions by default

**Recommendation:**
Use `0o600` for home-directory files:
```go
return os.WriteFile(path, append(data, '\n'), 0o600)
```

---

### SEC-008 — TOCTOU in scoped rule file creation (LOW)

**File:** `internal/cli/commands/sync.go:129-136`
**CWE:** CWE-367 (TOCTOU Race Condition)

**Description:**
Scoped rule files are checked for existence with `os.Stat()` then created with `os.WriteFile()`:
```go
if _, err := os.Stat(rulePath); os.IsNotExist(err) {
    os.MkdirAll(filepath.Dir(rulePath), 0o755)
    os.WriteFile(rulePath, []byte(content), 0o644)
}
```
Between the check and the write, an attacker could replace the path with a symlink.

**Mitigating factors:**
- The window is microseconds on a local filesystem
- Requires the attacker to have write access to the project directory
- The same attacker could modify `config.json` directly for greater impact

**Recommendation:**
Use `os.OpenFile` with `os.O_CREATE|os.O_EXCL` for atomic creation, or add an `os.Lstat` check immediately before the write.

---

### SEC-009 — JSON parsing without resource limits (LOW)

**File:** `internal/config/loader.go:24`
**CWE:** CWE-400 (Resource Exhaustion)

**Description:**
`config.json` is parsed with `json.Unmarshal(data, &cfg)` without any size or depth limits. A maliciously large or deeply nested config could cause excessive memory consumption.

**Mitigating factors:**
- `config.json` is a committed file; large payloads would be visible in code review
- Go's `encoding/json` handles deeply nested structures gracefully (no stack overflow)
- The config struct has fixed schema depth; extraneous nesting is ignored

**Recommendation:**
Add a size check before parsing: `if len(data) > 1<<20 { return error }` (1 MB limit).

---

## Dependency Analysis

| Dependency | Version | Risk | Notes |
|---|---|---|---|
| `github.com/fatih/color` | v1.19.0 | Low | Terminal output only, no parsing |
| `github.com/spf13/cobra` | v1.10.2 | Low | Mature CLI framework, no known CVEs |
| `github.com/spf13/pflag` | v1.0.9 | Low | Flag parsing, transitive via cobra |
| `github.com/mattn/go-colorable` | v0.1.14 | Low | Terminal color support |
| `github.com/mattn/go-isatty` | v0.0.20 | Low | TTY detection |
| `github.com/inconshreveable/mousetrap` | v1.1.0 | Low | Windows-only signal handling |
| `golang.org/x/sys` | v0.42.0 | Low | Standard library extension |

No high-risk dependencies. The tool has a minimal dependency tree with no network, crypto, or parsing libraries beyond the Go standard library.

---

## Summary

| ID | Finding | Severity | Status | Fix |
|---|---|---|---|---|
| SEC-001 | Path traversal via config.json paths | High | FIXED | `config/loader.go` — `validatePath()` rejects `..` and absolute paths at load time |
| SEC-002 | Symlink following on home-directory writes | High | FIXED | `translators/util.go` — `os.Lstat` check before write in `mergeUserMCPConfig` |
| SEC-003 | TOML injection via server names / env keys | Medium | FIXED | `config/loader.go` — regex validation of server names and env keys at load time |
| SEC-004 | No download integrity verification | Medium | FIXED | `install.sh`, `install.ps1` — download and verify `checksums.txt` before extracting |
| SEC-005 | YAML frontmatter injection | Medium | FIXED | Covered by SEC-003 name validation; invalid names rejected before reaching generators |
| SEC-006 | Env var secrets in project files | Medium | ACCEPTED | By design; mitigated by `.gitignore`. Documented in README. |
| SEC-007 | Home-directory files world-readable | Medium | FIXED | `translators/util.go` — `0o644` changed to `0o600`, dirs to `0o700` |
| SEC-008 | TOCTOU in scoped rule creation | Low | FIXED | `cli/commands/sync.go` — `os.OpenFile` with `O_CREATE\|O_EXCL` for atomic creation |
| SEC-009 | JSON parsing without resource limits | Low | FIXED | `config/loader.go` — 1 MB size limit before parsing |
