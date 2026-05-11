# Installation

## Requirements

- [Go 1.21+](https://go.dev/dl/)

On Linux, `xclip` or `xsel` is required for clipboard support:

```bash
sudo apt install xclip   # Debian/Ubuntu
```

---

## Option 1 – Windows installer (`.exe`)

The easiest way to install on Windows is the pre-built `screensaver-setup.exe`
installer generated with [Inno Setup](https://jrsoftware.org/isdl.php).

1. Download `screensaver-setup.exe` from the
   [Releases](https://github.com/aung-arata/screensaver/releases) page.
2. Run it and follow the wizard (installs to `Program Files\Screensaver`).
3. A Start Menu shortcut is created automatically; an optional Desktop shortcut
   is offered during setup.
4. To uninstall, use **Apps & Features / Add or Remove Programs**.

> **Note:** The installer requires administrator privileges to write to
> `Program Files`.

### Build this installer locally (Windows)

To generate the same `.exe` installer from source:

1. Install **Inno Setup 6** from <https://jrsoftware.org/isdl.php>.
2. Installer assets:
   - `packaging/windows/screensaver.iss`
   - `packaging/windows/build-installer.ps1`
3. Build `screensaver.exe` first:

```powershell
go build -o screensaver.exe ./cmd/screensaver
```

4. Build installer (choose one):

```powershell
# Manual
ISCC.exe packaging\windows\screensaver.iss
```

```powershell
# Helper script
.\packaging\windows\build-installer.ps1
```

Both commands output `screensaver-setup.exe` in the repository root.

---

## Option 2 – `go install` (recommended for developers)

```bash
go install github.com/aung-arata/screensaver/cmd/screensaver@latest
```

The binary is placed in `$GOPATH/bin` (usually `~/go/bin` on Linux/macOS or
`%USERPROFILE%\go\bin` on Windows).

> **Windows PATH note:** PowerShell will not find `screensaver` unless
> `%GOPATH%\bin` is in your `PATH`.  Run the helper script below, or add the
> directory manually:
>
> ```powershell
> # One-time setup – add Go's bin directory to your user PATH
> $goBin  = Join-Path (& go env GOPATH) "bin"
> $cur    = [System.Environment]::GetEnvironmentVariable("Path", "User")
> $segs   = if ($cur) { $cur -split ';' | Where-Object { $_ -ne '' } } else { @() }
> $exists = $segs | Where-Object { $_.TrimEnd('\','/') -ieq $goBin.TrimEnd('\','/') }
> if (-not $exists) {
>     $new = ($segs + $goBin) -join ';'
>     [System.Environment]::SetEnvironmentVariable("Path", $new, "User")
> }
> # Then open a new terminal window.
> ```

---

## Option 3 – Windows install script

Clone the repo and run the provided PowerShell helper, which builds the binary,
copies it to `%GOPATH%\bin`, and updates your `PATH` automatically:

```powershell
git clone https://github.com/aung-arata/screensaver.git
cd screensaver
.\install.ps1
```

Open a new terminal window after the script finishes, then run:

```powershell
screensaver --version
```

---

## Option 4 – Build locally (all platforms)

```bash
git clone https://github.com/aung-arata/screensaver.git
cd screensaver

# Linux / macOS
make build          # produces ./screensaver
# or: go build -o screensaver ./cmd/screensaver

# Windows (PowerShell / Command Prompt)
go build -o screensaver.exe ./cmd/screensaver
```

Copy the resulting binary (`screensaver` / `screensaver.exe`) to any directory
that is already on your `PATH`, or run it from its current location using the
full path (e.g. `.\screensaver.exe` on Windows).
