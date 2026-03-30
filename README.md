# Screensaver 🖼️

A **lightweight, portable screenshot tool** for Windows (with Linux/macOS support) inspired by [Lightshot](https://app.prntscr.com/).

Built in **Go** — compiles to a single binary with no runtime dependencies.

---

## Features

| Feature | Status |
|---|---|
| Full-screen capture | ✅ |
| Region capture | ✅ |
| Copy to clipboard | ✅ |
| Save to PNG / JPEG | ✅ |
| Multi-monitor support | ✅ |
| Full-screen selection overlay | ✅ (Windows only) |
| Rubber-band region selection | ✅ (Windows only) |
| Post-capture annotation editor | 🚧 |
| Pen / freehand drawing | 🚧 |
| Rectangle annotation | 🚧 |
| Arrow annotation | 🚧 |
| Text annotation | 🚧 |
| Undo | 🚧 |
| Global hotkey daemon | 🚧 |
| System tray icon | 🚧 |

> ✅ = implemented, 🚧 = planned / in progress

---

## Tech Stack

| Purpose | Library |
|---|---|
| **Language** | [Go](https://go.dev/) — single binary, no runtime deps, fast startup |
| Screen capture | [kbinani/screenshot](https://github.com/kbinani/screenshot) |
| Win32 hooks (hotkey, cursor) | [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows) |
| Clipboard integration | Platform-native (xclip/xsel on Linux, osascript on macOS, PowerShell on Windows) |
| Image annotation/drawing | [fogleman/gg](https://github.com/fogleman/gg) (planned) |
| System tray | [getlantern/systray](https://github.com/getlantern/systray) (planned) |
| GUI | [Fyne](https://fyne.io/) or [Walk](https://github.com/lxn/walk) or raw Win32 (planned) |
| Image encoding | stdlib `image/png`, `image/jpeg` |

---

## Requirements

- [Go 1.21+](https://go.dev/dl/)

On Linux, `xclip` or `xsel` is required for clipboard support:

```bash
sudo apt install xclip   # Debian/Ubuntu
```

---

## Installation

### From source

```bash
go install github.com/aung-arata/screensaver/cmd/screensaver@latest
```

### Build locally

```bash
git clone https://github.com/aung-arata/screensaver.git
cd screensaver
go build -o screensaver ./cmd/screensaver
```

This produces a single `screensaver` binary (or `screensaver.exe` on Windows).

---

## Usage

### One-shot mode (capture full screen)

```bash
# Copy screenshot to clipboard
screensaver --once

# Save screenshot to a file
screensaver --once --output screenshot.png
```

### Interactive region selection (Windows only)

```bash
# Select a region and copy to clipboard
screensaver --select

# Select a region and save to a file
screensaver --select --output region.png
```

The screen dims and a crosshair cursor appears. Click and drag to select the
region you want to capture. Release the mouse to capture the selection, or
press **Escape** to cancel.

> **Note:** `--select` requires Windows (Win32 APIs). On non-Windows platforms
> the command returns a "not yet implemented" error (see `overlay_stub.go`).
> Cross-platform support via a GUI toolkit (e.g. Fyne) is planned.

### Background daemon (planned)

```bash
screensaver
# Press Ctrl+Shift+S to take a screenshot
# Press Ctrl+C to quit
```

### Custom hotkey (planned)

```bash
screensaver --hotkey "ctrl+shift+p"
```

### Version

```bash
screensaver --version
```

---

## How it works

1. Run `screensaver --once` to capture the full screen.
2. The image is copied to the clipboard or saved to the specified path.

**Planned interactive workflow:**

1. Press the hotkey (default: **Ctrl+Shift+S**).
2. The screen dims and a crosshair cursor appears.
3. Click and drag to select the region you want to capture.
4. Release the mouse — the **Editor** window opens with the captured region.
5. Use the toolbar to annotate the image (pen, rectangle, arrow, text).
6. Click **Copy** to copy to clipboard, or **Save** to write to disk.
7. Press **Escape** at any time during selection to cancel.

---

## Architecture

```
cmd/
└── screensaver/
    └── main.go              # CLI entry point + flag parsing

internal/
├── capture/
│   ├── capture.go           # Screen capture (kbinani/screenshot)
│   └── capture_test.go      # Unit tests
├── clipboard/
│   └── clipboard.go         # Copy image to system clipboard
├── editor/
│   └── editor.go            # Post-capture annotation editor (planned)
├── hotkey/
│   ├── hotkey.go            # Global hotkey listener interface
│   ├── hotkey_windows.go    # Win32 RegisterHotKey implementation
│   └── hotkey_stub.go       # Stub for non-Windows platforms
├── overlay/
│   ├── overlay.go           # Selection overlay interface
│   ├── overlay_windows.go   # Win32 layered window (planned)
│   └── overlay_stub.go      # Stub for non-Windows platforms
├── tray/
│   └── tray.go              # System tray integration (planned)
└── utils/
    ├── utils.go             # Save image & path helpers
    └── utils_test.go        # Unit tests
```

---

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

## License

MIT
