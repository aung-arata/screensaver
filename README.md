# Screensaver 🖼️

A **lightweight, portable screenshot tool** for Windows inspired by [Lightshot](https://app.prntscr.com/).

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
| Post-capture annotation editor | ✅ |
| Pen / freehand drawing | ✅ |
| Rectangle annotation | ✅ |
| Arrow annotation | ✅ |
| Text annotation | ✅ |
| Undo | ✅ |
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
| Image annotation/drawing | [fogleman/gg](https://github.com/fogleman/gg) |
| System tray | [getlantern/systray](https://github.com/getlantern/systray) (planned) |
| GUI | [Fyne](https://fyne.io/) or [Walk](https://github.com/lxn/walk) or raw Win32 (planned) |
| Image encoding | stdlib `image/png`, `image/jpeg` |

---

## Browser Extension (Chrome)

A GoFullPage-style Chrome extension for capturing the full scrollable page of a browser. Use this when the main app can't capture browser content (e.g. full-page websites).

- Source: [browser-extension/](./browser-extension/)
- Load unpacked: `chrome://extensions` → Developer mode → Load unpacked → select `browser-extension/dist`

**Build:**

```bash
cd browser-extension
npm install
npm run build   # or npm run watch
```

Captures viewport screenshots by step-scroll + `chrome.tabs.captureVisibleTab`, stitches them with DOM offsets (deterministically), and opens a viewer where you can download PNG/JPG/PDF.

---

## Windows Installer

A ready-to-use `.exe` installer for Windows is generated with
[Inno Setup 6](https://jrsoftware.org/isdl.php).

| File | Purpose |
|---|---|
| `packaging/windows/screensaver.iss` | Inno Setup script |
| `packaging/windows/build-installer.ps1` | Helper — builds the binary **and** compiles the installer |

**Quick build (PowerShell):**

```powershell
# One-step: builds screensaver.exe then produces screensaver-setup.exe
.\packaging\windows\build-installer.ps1
```

Or compile manually after building `screensaver.exe`:

```powershell
ISCC.exe packaging\windows\screensaver.iss
```

Output: `screensaver-setup.exe` in the repository root.

See [docs/development.md](./docs/development.md#building-the-windows-installer) for full details.

---

## Documentation

- [Installation](./docs/installation.md) — requirements and install options
- [Usage](./docs/usage.md) — CLI flags and workflows
- [Architecture](./docs/architecture.md) — directory structure and how it works
- [Development](./docs/development.md) — running tests, cross-compilation, and building the installer
- [Roadmap](./docs/roadmap.md) — what's in progress, planned, and deferred

---

## License

MIT
