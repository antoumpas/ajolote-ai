# ajolote installer for Windows
# Usage: irm https://raw.githubusercontent.com/antoumpas/ajolote-ai/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo   = "antoumpas/ajolote-ai"
$Binary = "ajolote.exe"

# ── detect architecture ───────────────────────────────────────────────────────

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default {
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }
}

# ── resolve latest version ────────────────────────────────────────────────────

$Version = $env:VERSION
if (-not $Version) {
    try {
        $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
        $Version = $Release.tag_name
    } catch {
        Write-Error "Could not determine latest release. Set the VERSION environment variable and retry:`n  `$env:VERSION = 'v0.1.0'; iex (irm ...)"
        exit 1
    }
}

# ── download and install ──────────────────────────────────────────────────────

$Archive = "ajolote_windows_${Arch}.zip"
$Url     = "https://github.com/$Repo/releases/download/$Version/$Archive"

$InstallDir = Join-Path $env:USERPROFILE ".local\bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

$Tmp = Join-Path $env:TEMP "ajolote-install-$([System.IO.Path]::GetRandomFileName())"
New-Item -ItemType Directory -Path $Tmp | Out-Null

try {
    $ChecksumsUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"

    Write-Host "Downloading ajolote $Version (windows/$Arch)..."
    $ZipPath = Join-Path $Tmp $Archive
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath -UseBasicParsing

    # Verify download integrity (SEC-004)
    Write-Host "Verifying checksum..."
    $ChecksumsPath = Join-Path $Tmp "checksums.txt"
    Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $ChecksumsPath -UseBasicParsing
    $ExpectedLine = (Get-Content $ChecksumsPath | Where-Object { $_ -match $Archive })
    if ($ExpectedLine) {
        $ExpectedHash = ($ExpectedLine -split '\s+')[0]
        $ActualHash = (Get-FileHash -Algorithm SHA256 $ZipPath).Hash.ToLower()
        if ($ActualHash -ne $ExpectedHash) {
            Write-Error "Checksum verification failed — download may be corrupted or tampered with.`nExpected: $ExpectedHash`nActual:   $ActualHash"
            exit 1
        }
    } else {
        Write-Warning "Archive not found in checksums.txt — skipping verification."
    }

    Expand-Archive -Path $ZipPath -DestinationPath $Tmp -Force

    $Dest = Join-Path $InstallDir $Binary
    Move-Item -Path (Join-Path $Tmp $Binary) -Destination $Dest -Force
} finally {
    Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "ajolote $Version installed to $Dest"
Write-Host ""

# ── PATH hint ─────────────────────────────────────────────────────────────────

$CurrentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
if ($CurrentPath -notlike "*$InstallDir*") {
    Write-Host "Add the install directory to your PATH by running:"
    Write-Host "  [System.Environment]::SetEnvironmentVariable('PATH', `"`$env:PATH;$InstallDir`", 'User')"
    Write-Host ""
}

Write-Host "Get started:"
Write-Host "  cd your-project"
Write-Host "  ajolote init"
