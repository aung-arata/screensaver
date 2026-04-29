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

## Documentation

- [Installation](./docs/installation.md) — requirements and install options
- [Usage](./docs/usage.md) — CLI flags and workflows
- [Architecture](./docs/architecture.md) — directory structure and how it works
- [Development](./docs/development.md) — running tests and cross-compilation
- [Roadmap](./docs/roadmap.md) — what's in progress, planned, and deferred

---

## License

MIT
