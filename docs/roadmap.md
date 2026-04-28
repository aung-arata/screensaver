# Roadmap

> **Focus:** Windows-first. Cross-platform support is explicitly deferred until the Windows experience is complete.

---

## 🚧 In Progress (Windows)

- **Global hotkey daemon** — complete `hotkey_windows.go` with `RegisterHotKey` + Win32 message loop
- **System tray icon** — finish `tray/tray.go` with context menu (Capture, Region, Open Last, Settings, Quit) + capture notification
- **Interactive annotation editor GUI** — Fyne or raw Win32 toolbar with Pen, Rectangle, Arrow, Text, Color picker, Undo/Redo

---

## 🗓 Planned (Windows)

- **Config file support** (`%APPDATA%\screensaver\config.yaml`) — hotkey, save path, format, upload target
- **Upload / share integration** — Imgur or S3 upload, returns shareable URL to clipboard
- **Capture history** — local JSON index, `--history` flag, `--open last`
- **Output format expansion** — WebP support, JPEG quality flag, short GIF/screen recording
- **Scrolling / long-page capture** — stitch multiple captures of a scrollable region
- **Post-capture plugin hooks** — `--after-capture "cmd"` for external workflow integration
- **Structured logging / debug mode** — `--debug` flag using `log/slog`

---

## 🔮 Future / Deferred

- Cross-platform overlay on Linux/macOS (currently stubbed)
- Linux/macOS hotkey support

---

## ✅ Completed

- Full-screen capture
- Region capture (Windows)
- Copy to clipboard
- Save to PNG / JPEG
- Multi-monitor support
- Full-screen selection overlay (Windows)
- Rubber-band region selection (Windows)
- Post-capture annotation editor (render layer)
- Pen, Rectangle, Arrow, Text annotation
- Undo

---

## CI/CD & Release Pipeline (ongoing)

- GitHub Actions: `go test`, `go vet`, `staticcheck`
- Cross-compile matrix: windows/amd64 binary release
- Auto-attach `.exe` to GitHub Releases on tag push
