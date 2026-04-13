# 10 — Windsurf Translator

Tests for `ajolote use windsurf` (generate) and Windsurf import.

**Note:** Windsurf has no MCP support and does not import commands — export only for those features.

---

## Generate

### WINDSURF-001 — agents.md uses inlined content (not @file references)

**Prerequisites:** Initialized project; `.agents/rules/general.md` contains `# General\nAlways read before writing.`  
**Steps:**
1. `ajolote use windsurf`
2. `cat .windsurf/rules/agents.md`

**Expected result:** The actual text `Always read before writing.` appears directly in `agents.md`. File does NOT contain `@.agents/rules/general.md` references. Generation header present.  
**Pass / Fail:** ☐

---

### WINDSURF-002 — Commands written as YAML workflow files

**Prerequisites:** `.agents/commands/deploy.md` exists with content.  
**Steps:**
1. `ajolote use windsurf`
2. `cat .windsurf/workflows/deploy.yaml`

**Expected result:** YAML file exists for the `deploy` command. Content follows Windsurf workflow format with a steps array. `ajolote-sync.yaml` also present.  
**Pass / Fail:** ☐

---

### WINDSURF-003 — Scoped rule written under .windsurf/rules/

**Prerequisites:** Config has a scoped rule with a path and globs.  
**Steps:**
1. `ajolote use windsurf`
2. `ls .windsurf/rules/`

**Expected result:** A `<name>.md` file exists under `.windsurf/rules/` for the scoped rule (in addition to `agents.md`).  
**Pass / Fail:** ☐

---

## Import

### WINDSURF-004 — Import agents.md as rule file (if not ajolote-generated)

**Prerequisites:** `.windsurf/rules/agents.md` exists with hand-written content (no ajolote header).  
**Steps:**
1. `ajolote init` (fresh project)
2. `cat .agents/rules/general.md`

**Expected result:** Content from `agents.md` imported as the general rule file.  
**Pass / Fail:** ☐

---

### WINDSURF-005 — No MCP import from Windsurf

**Prerequisites:** Initialized project with `.windsurf/rules/agents.md` present.  
**Steps:**
1. `ajolote sync windsurf`
2. `cat .agents/config.json | grep '"servers"'`

**Expected result:** No new MCP servers added to config from Windsurf (Windsurf has no MCP config). Servers section unchanged.  
**Pass / Fail:** ☐

---

### WINDSURF-006 — No command import from Windsurf

**Prerequisites:** Initialized project; `.windsurf/workflows/` contains workflow files.  
**Steps:**
1. `ajolote sync windsurf`
2. `ls .agents/commands/`

**Expected result:** No new command files added from Windsurf workflows (Windsurf commands are not imported).  
**Pass / Fail:** ☐
