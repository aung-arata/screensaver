# Roadmap

> **Focus:** Windows-first. Cross-platform support is explicitly deferred until the Windows experience is complete.

---

## 🚧 In Progress (Windows)

- **Window-first scroll capture** — two-phase selection: click target window to focus, then draw region inside it; auto-focuses window before each scroll.
- **Enhanced annotation tools** – add ellipse, highlighter, blur/pixelate, crop, resizable shapes with rotation.
- **Floating toolbar after selection** – replace separate editor‑only launch with an immediate toolbar offering Copy, Save, Edit (like Lightshot).
- **Tray‑based Settings UI** – simple HTML/Go‑html dialog for hotkey, format, quality, save folder, upload service, history limits.
- **Optional cloud upload** – integrate with Imgur (anonymous) or custom S3 endpoint; return short URL to clipboard.
- **History viewer GUI** – grid view launched from tray (`--history-gui`) with thumbnails, open‑folder, re‑copy, delete, tagging.
- **Cross‑platform GUI selection** – evaluate Fyne/Winnative for drawing selection rectangle on macOS/Linux; reuse annotation canvas.

---

## 🗓 Planned (Windows)

- **Output format expansion** – WebP support, animated GIF/screen recording (short clips).
- **Post‑capture plugin hooks** – `--after-capture "cmd"` for external workflow integration.
- **Structured logging / debug mode** – `--debug` flag using `log/slog`.
- **Upload / share integration** – Imgur or S3 upload, returns shareable URL to clipboard *(low priority — local‑use tool)*.
- **Accessibility improvements** – optional sound feedback, high‑contrast cursor, toast notifications via systray.

---

## 🔮 Future / Deferred

- Cross‑platform overlay on Linux/macOS (currently stubbed).
- Linux/macOS hotkey support.
- Multi‑monitor aware scrolling capture on non‑Windows.
- Language localisation (i18n) for UI strings.

---

## ✅ Completed

- Full‑screen capture
- Region capture (Windows)
- Copy to clipboard
- Save to PNG / JPEG
- Multi‑monitor support
- Full‑screen selection overlay (Windows)
- Rubber‑band region selection (Windows)
- Post‑capture annotation editor (render layer)
- Pen, Rectangle, Arrow, Text annotation
- Undo
- **Global hotkey daemon** (Windows) – `RegisterHotKey` + Win32 message loop
- **System tray icon** (Windows) – Shell_NotifyIcon with context menu
- **Interactive annotation editor GUI** (Windows) – native Win32 toolbar (Pen, Rect, Arrow, Text, Color, Undo, Copy, Save)
- **Native save‑location picker dialog** (Windows) – `GetSaveFileNameW` via `comdlg32.dll`
- **"Open Last Screenshot" from tray** (Windows) – `ShellExecuteW` opens last saved file in default viewer
- **Config file support** – `%APPDATA%\screensaver\config.yaml` on Windows (`~/.screensaver/config.yaml` fallback); hotkey, save path, format, quality; `--format`, `--quality`, `--config` flags; `config show/init/path` subcommand
- **Custom tray icon** – embedded 16×16 ICO loaded via `CreateIconFromResourceEx`; also set on editor window via `WM_SETICON`
- **Single‑instance guard** – named Windows Mutex (`Global\ScreensaverDaemonMutex`) prevents duplicate daemon; shows MessageBox and exits if already running
- **Console detach in daemon mode** – `FreeConsole()` called when entering daemon mode so no black console window appears on the taskbar
- **Windows autostart** – `--install` / `--uninstall` flags write/remove `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` registry entry; `install.ps1` prompts for autostart registration
- **Editor zoom/pan** – mouse‑wheel zoom (0.1×–32×, zoom‑toward‑cursor), middle/right‑click drag pan, `clampPan()`
- **Editor keyboard shortcuts** – `P/R/A/T` tool switch, `Ctrl+0` fit, `Ctrl+1` 1:1, `+/-` step zoom
- **Editor status bar** – 22px strip: `Tool │ Zoom │ Image │ Cursor`; selective `InvalidateRect` on mouse move
- **Capture history** – `%APPDATA%\screensaver\history.json` (max 200 entries); `--history` / `--history-n` / `--history-clear` CLI flags; "Recent Screenshots" tray submenu (last 5)
- **Scrolling / long‑page capture (Windows-first)** – `--scroll` mode with selected region, auto‑scroll, overlap‑based vertical stitching, stop‑on‑no‑new‑content; graceful full-screen fallback on non-Windows.

---

## CI/CD & Release Pipeline (ongoing)

- GitHub Actions: `go test`, `go vet`, `staticcheck`
- Cross‑compile matrix: windows/amd64 binary release
- Auto‑attach `.exe` to GitHub Releases on tag push
