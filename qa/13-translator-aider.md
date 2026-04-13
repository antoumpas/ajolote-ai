# 13 — Aider Translator

Tests for `ajolote use aider` (generate). Aider has no import capability — it reads files directly via the `read:` list in its config.

---

### AIDER-001 — .aider.conf.yml created with correct structure

**Prerequisites:** Initialized project with rules, skills, personas, context.  
**Steps:**
1. `ajolote use aider`
2. `cat .aider.conf.yml`

**Expected result:** YAML file with at minimum:
```yaml
auto-commits: false
gitignore: false
read:
  - .agents/rules/general.md
  - .agents/skills/git.md
  - ...
```
All configured file paths appear in the `read:` list.  
**Pass / Fail:** ☐

---

### AIDER-002 — Scoped rule paths included in read list

**Prerequisites:** Config has a scoped rule with `"path": ".agents/rules/frontend.md"`.  
**Steps:**
1. `ajolote use aider`
2. `grep "frontend" .aider.conf.yml`

**Expected result:** `.agents/rules/frontend.md` appears in the `read:` list.  
**Pass / Fail:** ☐

---

### AIDER-003 — Import returns empty (no config to import)

**Prerequisites:** Initialized project; `.aider.conf.yml` exists.  
**Steps:**
1. `ajolote sync aider`
2. Observe output.

**Expected result:** No import output (Aider has nothing to import). Export phase regenerates the file. No error.  
**Pass / Fail:** ☐

---

### AIDER-004 — File paths updated on regeneration

**Prerequisites:** Initialized project; `ajolote use aider` already run. Then add a new skill to config and create the file.  
**Steps:**
1. Add `.agents/skills/testing.md` to config `skills` array and create the file.
2. `ajolote use aider`
3. `grep "testing" .aider.conf.yml`

**Expected result:** `.agents/skills/testing.md` now appears in the `read:` list.  
**Pass / Fail:** ☐
