# Development

## Running tests

```bash
go test ./...
```

With verbose output:

```bash
go test -v ./...
```

---

## Cross-compilation

Build for Windows from any platform:

```bash
GOOS=windows GOARCH=amd64 go build -o screensaver.exe ./cmd/screensaver
```

---

## Building the Windows installer

The Windows `.exe` installer is created with
[Inno Setup 6](https://jrsoftware.org/isdl.php).

### Prerequisites

1. Build `screensaver.exe` first (see Cross-compilation above, or run
   `go build -o screensaver.exe ./cmd/screensaver` on Windows).
2. Install **Inno Setup 6** from <https://jrsoftware.org/isdl.php>.

### Compile the installer manually

`screensaver.exe` must already exist in the repository root before running this command (see Prerequisites above).

```powershell
# From the repo root
ISCC.exe packaging\windows\screensaver.iss
```

This produces `screensaver-setup.exe` in the repository root.

### Helper script (Windows only)

A convenience script builds `screensaver.exe` and then compiles the installer
in one step:

```powershell
.\packaging\windows\build-installer.ps1
```

The script automatically locates `ISCC.exe` on `PATH` or at the standard Inno
Setup install paths (`%ProgramFiles(x86)%\Inno Setup 6\ISCC.exe` on most 64-bit
systems, or `%ProgramFiles%\Inno Setup 6\ISCC.exe`).

### Releasing

`screensaver-setup.exe` can be attached as a release asset to a
[GitHub Release](https://github.com/aung-arata/screensaver/releases).
