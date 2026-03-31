<#
.SYNOPSIS
    Installs screensaver.exe to %GOPATH%\bin and ensures that directory is on your PATH.

.DESCRIPTION
    1. Builds screensaver.exe from source using `go build`.
    2. Copies the binary to %GOPATH%\bin (creates the directory if it does not exist).
    3. Adds %GOPATH%\bin to the current user's PATH environment variable if it is
       not already present.

    Open a new terminal (or restart your current one) after running this script so
    that the updated PATH takes effect.

.EXAMPLE
    .\install.ps1
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$scriptDir = $PSScriptRoot

# ---------------------------------------------------------------------------
# 1. Build
# ---------------------------------------------------------------------------
Write-Host "Building screensaver.exe ..."
go build -o (Join-Path $scriptDir "screensaver.exe") (Join-Path $scriptDir "cmd\screensaver")
if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed. Make sure Go is installed and that Go is on your PATH."
    exit 1
}

# ---------------------------------------------------------------------------
# 2. Determine install directory (%GOPATH%\bin or %USERPROFILE%\go\bin)
# ---------------------------------------------------------------------------
$goPath = & go env GOPATH
if (-not $goPath) {
    $goPath = Join-Path $HOME "go"
}
$goBin = Join-Path $goPath "bin"

if (-not (Test-Path $goBin)) {
    New-Item -ItemType Directory -Path $goBin | Out-Null
    Write-Host "Created directory: $goBin"
}

# ---------------------------------------------------------------------------
# 3. Copy binary
# ---------------------------------------------------------------------------
$dest = Join-Path $goBin "screensaver.exe"
Copy-Item -Path (Join-Path $scriptDir "screensaver.exe") -Destination $dest -Force
Write-Host "Installed: $dest"

# ---------------------------------------------------------------------------
# 4. Add GOPATH\bin to the user PATH (persistent, current user only)
# ---------------------------------------------------------------------------
$userPath = [System.Environment]::GetEnvironmentVariable("Path", "User")
if ($null -eq $userPath) { $userPath = "" }

# Split on ';', trim trailing slashes/backslashes, compare case-insensitively.
$normalizedBin = $goBin.TrimEnd('\', '/')
$segments = $userPath -split ';' | Where-Object { $_ -ne '' }
$alreadyPresent = $segments | Where-Object { $_.TrimEnd('\', '/') -ieq $normalizedBin }

if (-not $alreadyPresent) {
    $newSegments = $segments + $goBin
    $newPath = $newSegments -join ';'
    [System.Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host ""
    Write-Host "Added '$goBin' to your user PATH."
    Write-Host "Please open a new terminal window for the change to take effect."
} else {
    Write-Host "'$goBin' is already in your PATH."
}

Write-Host ""
Write-Host "Done. You can now run:  screensaver --help"
