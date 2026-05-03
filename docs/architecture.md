# Architecture

## Directory Structure

```
cmd/
└── screensaver/
    ├── main.go                       # CLI entry point, flag parsing, mode dispatch
    ├── singleinstance_windows.go     # MessageBox shown when daemon already running
    └── singleinstance_stub.go        # no-op stub for non-Windows

internal/
├── appicon/
│   └── icon.go                       # Embedded 16×16 ICO bytes (camera icon)
├── autostart/
│   ├── autostart_windows.go          # HKCU Run registry install/uninstall
│   └── autostart_stub.go             # no-op stub for non-Windows
├── capture/
│   ├── capture.go                    # Full-screen and region capture (kbinani/screenshot)
│   └── capture_test.go
├── clipboard/
│   └── clipboard.go                  # Copy image.Image to system clipboard
├── config/
│   ├── config.go                     # YAML config loader, defaults, merge with CLI flags
│   └── config_test.go
├── console/
│   ├── console_windows.go            # FreeConsole() — detach from terminal in daemon mode
│   └── console_stub.go               # no-op stub for non-Windows
├── editor/
│   ├── editor.go                     # Annotation engine: Pen, Rect, Arrow, Text; Render/Undo
│   ├── editor_windows.go             # Native Win32 editor GUI: toolbar, zoom/pan, status bar
│   ├── editor_stub.go                # Fallback: saves annotated PNG to disk
│   └── editor_test.go
├── history/
│   ├── history.go                    # JSON capture history index (max 200 entries)
│   └── history_test.go
├── hotkey/
│   ├── hotkey.go                     # Listener interface
│   ├── hotkey_windows.go             # Win32 RegisterHotKey + message loop
│   └── hotkey_stub.go                # no-op stub for non-Windows
├── overlay/
│   ├── overlay.go                    # Selection overlay interface
│   ├── overlay_windows.go            # Win32 fullscreen dimmed overlay + rubber-band select
│   └── overlay_stub.go               # no-op stub for non-Windows
├── savedialog/
│   ├── savedialog_windows.go         # GetSaveFileNameW via comdlg32.dll
│   └── savedialog_stub.go            # no-op stub for non-Windows
├── singleinstance/
│   ├── singleinstance_windows.go     # Named Windows mutex (Global\ScreensaverDaemonMutex)
│   └── singleinstance_stub.go        # Always returns true on non-Windows
├── tray/
│   ├── tray.go                       # Config struct + Run() entry point
│   ├── tray_windows.go               # Shell_NotifyIconW, context menu, Recent submenu
│   └── tray_stub.go                  # no-op stub for non-Windows
└── utils/
    ├── utils.go                      # SaveImage, GenerateFilename, DefaultSaveDirectory
    └── utils_test.go
```

---

## How it works

### Daemon mode (default)

1. `singleinstance.Acquire()` — creates a named Windows mutex; exits with MessageBox if already running.
2. `console.Detach()` — calls `FreeConsole()` so no black terminal appears on the taskbar.
3. `hotkey.NewListener` registers the global hotkey (default **Ctrl+Shift+S**) via `RegisterHotKey`.
4. `tray.Run` creates a hidden Win32 message-only window, adds a `Shell_NotifyIconW` tray icon with the embedded camera icon, and runs the message loop.
5. On hotkey or tray **Take Screenshot**: `captureAndEdit()` → `capture.FullScreen` → `editor.Run` (Win32 GUI).
6. On tray **Select Region**: `selectAndEdit()` → `overlay.Show` → `capture.CaptureRegion` → `editor.Run`.
7. On editor **Save**: `savedialog` shows `GetSaveFileNameW`; after save `history.Add` records the entry.
8. On tray **Recent Screenshots ▶**: submenu shows last 5 entries from `history.Recent(5)`; click opens file with `ShellExecuteW`.
9. On tray **Quit**: mutex released, daemon exits.

### One-shot modes

```
screensaver --once            capture full screen → clipboard
screensaver --once --output   capture full screen → save to path
screensaver --once --edit     capture full screen → editor
screensaver --select          region select → clipboard
screensaver --select --output region select → save to path
screensaver --select --edit   region select → editor
```

### Config load order

```
Defaults → config.yaml (APPDATA\screensaver\config.yaml) → CLI flags
```

CLI flags always win. Use `screensaver config show` to print the merged result.

---

## Package dependency graph

```
main
 ├── capture
 ├── clipboard
 ├── config
 ├── console
 ├── editor ──────── savedialog
 │                   appicon
 │                   clipboard
 │                   utils
 ├── history
 ├── hotkey
 ├── overlay
 ├── singleinstance
 ├── tray ─────────  appicon
 │                   history
 └── utils
```

No import cycles. All Windows-specific packages have a paired `_stub.go` for non-Windows builds.
