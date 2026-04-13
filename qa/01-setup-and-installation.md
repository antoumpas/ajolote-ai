# 01 — Setup & Installation

Tests for `install.sh` (macOS / Linux) and `install.ps1` (Windows PowerShell).

**Scope:** Binary download, extraction, placement, and post-install verification.  
**Not covered here:** Feature behaviour — that starts at [02-init.md](02-init.md).

---

### SETUP-001 — install.sh on macOS (arm64)

**Prerequisites:** macOS arm64 machine; `curl` available; `/usr/local/bin` writable or `sudo` available; ajolote NOT currently installed.  
**Steps:**
1. `curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh`
2. After completion: `which ajolote`
3. `ajolote --help`

**Expected result:** Script prints `Downloading ajolote vX.Y.Z (darwin/arm64)...` and `ajolote vX.Y.Z installed to /usr/local/bin/ajolote`. `which ajolote` returns `/usr/local/bin/ajolote`. `--help` prints usage.  
**Pass / Fail:** ☐

---

### SETUP-002 — install.sh on macOS (amd64)

**Prerequisites:** macOS amd64 machine; same as SETUP-001.  
**Steps:**
1. `curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh`
2. `ajolote --help`

**Expected result:** Script prints `(darwin/amd64)` in the download message. Binary installed and functional.  
**Pass / Fail:** ☐

---

### SETUP-003 — install.sh on Linux (amd64)

**Prerequisites:** Linux amd64 machine or container; `curl` available.  
**Steps:**
1. `curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh`
2. `ajolote --help`

**Expected result:** Script prints `(linux/amd64)`. Binary installed at `/usr/local/bin/ajolote`.  
**Pass / Fail:** ☐

---

### SETUP-004 — install.sh VERSION override

**Prerequisites:** macOS or Linux; a known older release tag (e.g., `v0.1.0`) exists.  
**Steps:**
1. `VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh`
2. `ajolote --help`

**Expected result:** Script uses `v0.1.0` — download URL contains `v0.1.0`. Binary at that version installed.  
**Pass / Fail:** ☐

---

### SETUP-005 — install.sh INSTALL_DIR override

**Prerequisites:** macOS or Linux; a writable custom directory (e.g., `~/bin`).  
**Steps:**
1. `mkdir -p ~/bin`
2. `INSTALL_DIR=~/bin curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh`
3. `ls ~/bin/ajolote`
4. `~/bin/ajolote --help`

**Expected result:** Script installs to `~/bin/ajolote`. Final message shows the custom path. Binary is executable.  
**Pass / Fail:** ☐

---

### SETUP-006 — install.sh on Git Bash / MSYS2 (Windows redirect)

**Prerequisites:** Windows machine with Git Bash or MSYS2.  
**Steps:**
1. Open Git Bash.
2. `curl -fsSL https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.sh | sh`

**Expected result:** Script prints `Windows detected. Use the PowerShell installer instead:` followed by the `irm ... | iex` command. Exits with code 0 (no error). Does NOT attempt to download a binary.  
**Pass / Fail:** ☐

---

### SETUP-007 — install.ps1 on Windows (amd64)

**Prerequisites:** Windows 10/11 amd64; PowerShell 5.1+; internet access; `$env:USERPROFILE\.local\bin` does not exist yet.  
**Steps:**
1. Open PowerShell.
2. `irm https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.ps1 | iex`
3. After completion: `Get-Item "$env:USERPROFILE\.local\bin\ajolote.exe"`
4. `& "$env:USERPROFILE\.local\bin\ajolote.exe" --help`

**Expected result:** Script prints `Downloading ajolote vX.Y.Z (windows/amd64)...` and `ajolote vX.Y.Z installed to ...ajolote.exe`. The file exists and `--help` prints usage.  
**Pass / Fail:** ☐

---

### SETUP-008 — install.ps1 PATH hint when dir not in PATH

**Prerequisites:** Windows; `$env:USERPROFILE\.local\bin` not in the current user's PATH.  
**Steps:**
1. Run the PowerShell installer (as in SETUP-007).
2. Observe console output after installation completes.

**Expected result:** Script prints `Add the install directory to your PATH by running:` followed by the `[System.Environment]::SetEnvironmentVariable(...)` command. Hint is only shown when the directory is absent from PATH.  
**Pass / Fail:** ☐

---

### SETUP-009 — install.ps1 VERSION override

**Prerequisites:** Windows PowerShell; a known older release tag exists.  
**Steps:**
1. `$env:VERSION = 'v0.1.0'`
2. `irm https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.ps1 | iex`

**Expected result:** Download URL contains `v0.1.0`. Binary at that version installed successfully.  
**Pass / Fail:** ☐

---

### SETUP-010 — --version reports the installed release tag

**Prerequisites:** ajolote installed via `install.sh` or `install.ps1` (any platform); note the release tag shown during installation (e.g. `v0.7.0`).  
**Steps:**
1. `ajolote --version`

**Expected result:** Output is `ajolote version X.Y.Z` where `X.Y.Z` matches the release tag downloaded (without the leading `v`). Must NOT show `dev`, `0.0.1`, or any other placeholder.  
**Pass / Fail:** ☐
