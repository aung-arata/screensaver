# Roadmap

> **Focus:** Windows-first. Cross-platform support is explicitly deferred until the Windows experience is complete.

---

## 🚧 In Progress (Windows)

*(All Windows-first items listed below have been completed — see ✅ Completed.)*

---

## 🗓 Planned (Windows)

- **Upload / share integration** — Imgur or S3 upload, returns shareable URL to clipboard *(requires Config — now done)*
- **Capture history** — local JSON index, `--history` flag, `--open last` *(requires Config — now done)*
- **Output format expansion** — WebP support, short GIF/screen recording
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
- **Global hotkey daemon** (Windows) — `RegisterHotKey` + Win32 message loop
- **System tray icon** (Windows) — Shell_NotifyIcon with context menu
- **Interactive annotation editor GUI** (Windows) — native Win32 toolbar (Pen, Rect, Arrow, Text, Color, Undo, Copy, Save)
- **Native save-location picker dialog** (Windows) — `GetSaveFileNameW` via `comdlg32.dll`
- **"Open Last Screenshot" from tray** (Windows) — `ShellExecuteW` opens last saved file in default viewer
- **Config file support** — `%APPDATA%\screensaver\config.yaml` on Windows (`~/.screensaver/config.yaml` fallback); hotkey, save path, format, quality; `--format`, `--quality`, `--config` flags; `config show/init/path` subcommand

---

## CI/CD & Release Pipeline (ongoing)

- GitHub Actions: `go test`, `go vet`, `staticcheck`
- Cross-compile matrix: windows/amd64 binary release
- Auto-attach `.exe` to GitHub Releases on tag push
