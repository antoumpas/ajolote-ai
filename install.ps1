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
    Write-Host "Downloading ajolote $Version (windows/$Arch)..."
    $ZipPath = Join-Path $Tmp $Archive
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath -UseBasicParsing

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
