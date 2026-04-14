# QA 22 — Config Inheritance (`extends`)

## Prerequisites

- `ajolote` binary built (`make build` or `go install`)
- A terminal with `git` available (for git source tests)
- Internet access optional (for HTTPS / git SSH tests)

---

## Setup: Create a base standards project

```sh
mkdir /tmp/ai-standards && cd /tmp/ai-standards
ajolote init
```

Edit `.agents/config.json` to add an org-wide MCP server and confirm default rules exist:

```json
{
  "mcp": {
    "servers": {
      "org-github": {
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-github"],
        "env": { "GITHUB_TOKEN": "${GITHUB_TOKEN}" },
        "scope": "user"
      }
    }
  },
  "rules": [".agents/rules/org-style.md"],
  "skills": [".agents/skills/git.md"],
  "personas": [],
  "context": [],
  "commands": [".agents/commands/review.md"]
}
```

Create `.agents/rules/org-style.md`:

```sh
echo "# Org Style Rules\nAlways write tests." > .agents/rules/org-style.md
```

Rename `general.md` to `org-style.md` or create as above. Commit the base repo:

```sh
git init && git add .agents && git commit -m "init org standards"
```

---

## Test 1 — Local file inheritance

```sh
mkdir /tmp/my-project && cd /tmp/my-project
ajolote init
```

Edit `.agents/config.json` to add `"extends": "/tmp/ai-standards"`:

```json
{
  "extends": "/tmp/ai-standards",
  "mcp": { "servers": {} },
  "rules": [".agents/rules/general.md"],
  "skills": [],
  "personas": [],
  "context": []
}
```

Run:

```sh
AJOLOTE_CACHE_TTL_SECONDS=0 ajolote use claude
```

**Expected:**
- `.agents/.base/` directory is created, containing `config.json`, `rules/org-style.md`, etc.
- `CLAUDE.md` includes `@.agents/.base/rules/org-style.md` (inherited rule)
- `CLAUDE.md` also includes `@.agents/rules/general.md` (local rule)
- `.claude/settings.json` includes `org-github` MCP server

---

## Test 2 — Local overrides base

Still in `/tmp/my-project`, create a local override with the same filename as the base rule:

```sh
echo "# Local Org Style (override)" > .agents/rules/org-style.md
```

Update `config.json` to include it:

```json
{
  "extends": "/tmp/ai-standards",
  "rules": [".agents/rules/general.md", ".agents/rules/org-style.md"],
  ...
}
```

```sh
AJOLOTE_CACHE_TTL_SECONDS=0 ajolote use claude
```

**Expected:**
- `CLAUDE.md` includes `@.agents/rules/org-style.md` (local version)
- `CLAUDE.md` does NOT include `@.agents/.base/rules/org-style.md`
- No duplicate `org-style.md` entries

---

## Test 3 — MCP server conflict: local wins

Add a server with the same name in the local config:

```json
{
  "extends": "/tmp/ai-standards",
  "mcp": {
    "servers": {
      "org-github": { "command": "node", "args": ["./local-github.js"] }
    }
  },
  ...
}
```

```sh
AJOLOTE_CACHE_TTL_SECONDS=0 ajolote use claude
```

**Expected:**
- `.claude/settings.json` (or `~/.claude.json` for user-scoped) has `org-github` pointing to `node ./local-github.js` (local version wins)

---

## Test 4 — Command inheritance (local source)

Add a command to the base project:

```sh
echo "# Deploy\nRun the deployment script." > /tmp/ai-standards/.agents/commands/deploy.md
```

```sh
cd /tmp/my-project
AJOLOTE_CACHE_TTL_SECONDS=0 ajolote use claude
```

**Expected:**
- `.claude/commands/deploy.md` exists and contains the inherited deploy command
- If local project also has a `deploy.md` in `.agents/commands/`, the local one is used

---

## Test 5 — `ajolote validate` shows Extends section

```sh
cd /tmp/my-project
ajolote validate
```

**Expected:**
```
Extends
  ✔ /tmp/ai-standards

Rules
  ✔ .agents/rules/general.md
  ✔ .agents/.base/rules/org-style.md
  ...
```

---

## Test 6 — Invalid extends scheme

```sh
echo '{"extends":"s3://bucket/standards","mcp":{"servers":{}},"rules":[],"skills":[],"personas":[],"context":[]}' \
  > /tmp/my-project/.agents/config.json
ajolote validate
```

**Expected:**
```
Extends
  ✘ s3://bucket/standards — unsupported source scheme ...

validation failed
```

Exit code: 1

---

## Test 7 — Cache TTL behaviour

```sh
cd /tmp/my-project
# Restore valid config
ajolote use claude          # populates .agents/.base/ (first fetch)
ls -la .agents/.base/       # note mtime
ajolote use claude          # second call — cache should be used (fast, no network)
AJOLOTE_CACHE_TTL_SECONDS=0 ajolote use claude   # force refresh
```

**Expected:**
- First call: `.agents/.base/.meta.json` is created
- Second call: `.meta.json` mtime unchanged (cache reused)
- Third call: `.meta.json` mtime updated (re-fetched)

---

## Test 8 — `ajolote diff` reflects inheritance

```sh
cd /tmp/my-project
ajolote use claude                          # generate files
ajolote diff claude                         # should show nothing would change
echo "new line" >> .agents/.base/rules/org-style.md   # tamper with cache
ajolote diff claude                         # should show a diff
```

**Expected:**
- After tampering: `ajolote diff` exits 1 and shows the changed lines
- `ajolote sync` restores consistency

---

## Test 9 — Git source inheritance (SSH)

Requires a real git repo accessible via SSH (e.g. a private org repo):

```sh
cd /tmp/git-project
ajolote init
```

Edit config.json:

```json
{
  "extends": "git@github.com:your-org/ai-standards.git",
  "mcp": { "servers": {} },
  "rules": [],
  "skills": [],
  "personas": [],
  "context": []
}
```

```sh
AJOLOTE_CACHE_TTL_SECONDS=0 ajolote use claude
```

**Expected:**
- `git` is called with `--depth 1 --sparse`, `.agents/` is fetched
- `.agents/.base/` is populated with the org's standards
- Generated tool files reflect inherited config

---

## Test 10 — Git source: git not in PATH

```sh
PATH=/usr/bin ajolote use claude   # exclude git directory from PATH if needed
```

**Expected:**
- Clear error: `git not found in PATH (required for git source "...") — install git and retry`

---

## Test 11 — HTTPS source with `commands` array

Set up an HTTP server (or use `python3 -m http.server` in the base project directory):

```sh
cd /tmp/ai-standards
python3 -m http.server 9876 &
```

In a project:

```json
{
  "extends": "http://localhost:9876",
  ...
}
```

```sh
AJOLOTE_CACHE_TTL_SECONDS=0 ajolote use claude
```

**Expected:**
- `config.json` and each file listed in it are fetched via HTTP
- Command files listed in `commands` array are fetched
- Unlisted command files are NOT fetched

---

## Test 12 — Stale cache fallback

```sh
cd /tmp/my-project
ajolote use claude                   # populate cache
rm -rf /tmp/ai-standards              # make base unreachable
ajolote use claude                   # should warn, reuse stale cache
```

**Expected:**
- Warning printed to stderr: `could not refresh base config from "..." (using cached copy): ...`
- Command still succeeds using the stale `.agents/.base/` contents

---

## Test 13 — `.gitignore` entry

```sh
mkdir /tmp/gitignore-test && cd /tmp/gitignore-test
ajolote init
grep ".agents/.base/" .gitignore
```

**Expected:**
- `.agents/.base/` appears in the ajolote-managed gitignore block

---

## Test 14 — `ajolote sync` preserves `extends` field

```sh
cd /tmp/my-project
ajolote sync claude
cat .agents/config.json | grep extends
```

**Expected:**
- `"extends"` field is still present in `.agents/config.json` after sync (not dropped by Save)

---

## Test 15 — Inherited files visible to all tools

```sh
cd /tmp/my-project
ajolote use cursor
cat .cursor/rules/agents.mdc
```

**Expected:**
- Cursor config references or inlines the inherited rules (`org-style.md` from `.agents/.base/rules/`)
- All 9 tools reflect inherited config when run with `use`
