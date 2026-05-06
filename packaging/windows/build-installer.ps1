# build-installer.ps1
# Builds screensaver.exe then compiles the Inno Setup installer.
#
# Usage (from the repo root):
#   .\packaging\windows\build-installer.ps1
#
# Requirements:
#   - Go 1.21+
#   - Inno Setup 6 (ISCC.exe on PATH, or installed at the default location)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
$IssScript = Join-Path $PSScriptRoot 'screensaver.iss'

# --- 1. Build the Go binary ---
Write-Host "Building screensaver.exe..." -ForegroundColor Cyan
Push-Location $RepoRoot
try {
    go build -o screensaver.exe ./cmd/screensaver
} finally {
    Pop-Location
}
Write-Host "screensaver.exe built successfully." -ForegroundColor Green

# --- 2. Extract version from main.go ---
$versionLine = Select-String -Path (Join-Path $RepoRoot 'cmd\screensaver\main.go') `
    -Pattern 'Version\s*=\s*"([^"]+)"' | Select-Object -First 1
if ($versionLine -and $versionLine.Matches.Count -gt 0) {
    $appVersion = $versionLine.Matches[0].Groups[1].Value
} else {
    $appVersion = '0.0.0'
    Write-Warning "Could not read version from main.go; using '$appVersion'"
}
Write-Host "App version: $appVersion" -ForegroundColor Cyan

# --- 3. Locate ISCC.exe ---
$iscc = $null

# Check PATH first
$iscc = Get-Command 'ISCC.exe' -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source

if (-not $iscc) {
    # Fall back to the default Inno Setup 6 install location
    $defaultPath = Join-Path $env:ProgramFiles 'Inno Setup 6\ISCC.exe'
    if (Test-Path $defaultPath) {
        $iscc = $defaultPath
    }
}

if (-not $iscc) {
    Write-Error @"
ISCC.exe not found.
Install Inno Setup 6 from https://jrsoftware.org/isdl.php and ensure ISCC.exe is on your PATH,
or install it to the default location: $($env:ProgramFiles)\Inno Setup 6\
"@
    exit 1
}

Write-Host "Using Inno Setup compiler: $iscc" -ForegroundColor Cyan

# --- 4. Compile the installer ---
Write-Host "Compiling installer..." -ForegroundColor Cyan
& $iscc /DMyAppVersion=$appVersion $IssScript

if ($LASTEXITCODE -ne 0) {
    Write-Error "ISCC.exe failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}

Write-Host ""
Write-Host "Installer built successfully: $RepoRoot\screensaver-setup.exe" -ForegroundColor Green
