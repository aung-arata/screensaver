# Installation

## Requirements

- [Go 1.21+](https://go.dev/dl/)

On Linux, `xclip` or `xsel` is required for clipboard support:

```bash
sudo apt install xclip   # Debian/Ubuntu
```

---

## Option 1 – `go install` (recommended)

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

## Option 2 – Windows install script

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

## Option 3 – Build locally (all platforms)

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
