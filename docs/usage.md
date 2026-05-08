# Usage

## Quick reference

| Flag | Mode | Description |
|------|------|-------------|
| *(none)* | daemon | Background daemon with hotkey + tray icon |
| `--once` | one-shot | Capture full screen and exit |
| `--select` | one-shot | Interactive region selection and exit |
| `--scroll` | one-shot | **Experimental** Windows-only long-page capture with auto-scroll + vertical stitching |
| `--scroll-delay <ms>` | modifier | Delay between scroll/capture steps in `--scroll` mode (default 250) |
| `--scroll-step <n>` | modifier | Scroll amount/wheel delta used in `--scroll` mode (default 900) |
| `--scroll-max <n>` | modifier | Maximum frames captured in `--scroll` mode (default 20) |
| `--edit` | modifier | Open annotation editor after capture (combine with `--once`, `--select`, or `--scroll`) |
| `--output <path>` | modifier | Save to explicit file path (combine with `--once`, `--select`, or `--scroll`) |
| `--save-dir <dir>` | modifier | Auto-save to directory with timestamped filename |
| `--hotkey <combo>` | daemon | Override hotkey (default: `ctrl+shift+s`) |
| `--format <fmt>` | any | Output format: `png` or `jpeg` (default from config, fallback `png`) |
| `--quality <n>` | any | JPEG quality 1–100 (default from config, fallback 90; ignored for PNG) |
| `--config <path>` | any | Use a custom config file instead of the default location |
| `--install` | utility | Register for Windows autostart (HKCU Run) and exit |
| `--uninstall` | utility | Remove from Windows autostart and exit |
| `--history` | utility | List recent screenshots and exit |
| `--history-n <n>` | utility | Number of entries to show with `--history` (default 20) |
| `--history-clear` | utility | Clear all screenshot history and exit |
| `--version` | utility | Print version and exit |

---

## One-shot mode (full screen)

```powershell
# Copy screenshot to clipboard
screensaver --once

# Save screenshot to an explicit path
screensaver --once --output screenshot.png

# Auto-save to a directory with a timestamped filename
screensaver --once --save-dir C:\Users\You\Pictures

# Open annotation editor after capture
screensaver --once --edit

# Capture as JPEG at quality 85
screensaver --once --format jpeg --quality 85 --save-dir C:\Screenshots
```

---

## Interactive region selection (Windows only)

```powershell
# Select a region and copy to clipboard
screensaver --select

# Select a region and save to a directory
screensaver --select --save-dir C:\Screenshots

# Select a region and open the annotation editor
screensaver --select --edit
```

The screen dims and a crosshair cursor appears. Click and drag to select the
region you want to capture. Release to capture, or press **Escape** to cancel.

> **Note:** `--select` requires Windows (Win32 APIs). On non-Windows platforms
> the command returns a "not yet implemented" error.

---

## Experimental long-page capture (Windows only)

`--scroll` is an **experimental** Windows-only mode that captures a selected
region repeatedly while auto-scrolling downward, then stitches the frames into
one long image.

```powershell
screensaver --scroll --edit
screensaver --scroll --output page.png
screensaver --scroll --save-dir C:\Screenshots --scroll-delay 300 --scroll-step 900 --scroll-max 15
```

Behavior:
- `--scroll` uses region selection first (draw the scrollable area you want)
- Stops when no meaningful new content is detected or `--scroll-max` is reached
- Output follows normal one-shot flow: editor, explicit output path, auto-save
  directory, or clipboard if no save path is provided

> **Note:** On non-Windows platforms `--scroll` returns:
> `long-page capture is only supported on Windows right now`

---

## Annotation editor (Windows)

The editor opens automatically in daemon mode after every capture. It can also
be opened from the one-shot modes using `--edit`:

```powershell
screensaver --once --edit
screensaver --select --edit
```

**Toolbar actions:**

| Button / Key | Action |
|---|---|
| `P` or Pen button | Freehand pen stroke |
| `R` or Rect button | Hollow rectangle |
| `A` or Arrow button | Arrow with arrowhead |
| `T` or Text button | Text annotation |
| Color button | Pick annotation colour |
| Undo / `Ctrl+Z` | Remove last annotation |
| Copy | Copy annotated image to clipboard |
| Save | Save annotated image via file dialog |
| `Ctrl+0` | Fit image to window |
| `Ctrl+1` | 1:1 actual pixels |
| `+` / `-` | Step zoom in / out |
| Mouse wheel | Zoom toward cursor |
| Middle-click drag | Pan |
| Right-click drag | Pan |

Status bar shows: `Tool │ Zoom % │ Image W×H │ Cursor X,Y`.

---

## Background daemon

```powershell
# Start with default hotkey (Ctrl+Shift+S)
screensaver

# Use a custom hotkey
screensaver --hotkey "ctrl+shift+p"
```

Press the hotkey to capture the full screen and open the editor.  
Right-click the tray icon for the context menu:
- **Take Screenshot** — full-screen capture → editor
- **Select Region** — overlay selection → editor
- **Recent Screenshots ▶** — last 5 captures; click to open
- **Open Last Screenshot** — opens the most recently saved file
- **Quit** — exit the daemon

---

## Capture history

Every screenshot saved to disk is recorded in `%APPDATA%\screensaver\history.json` (max 200 entries).

```powershell
# List the 20 most recent screenshots (default)
screensaver --history

# List the 50 most recent
screensaver --history --history-n 50

# Clear all history
screensaver --history-clear
```

Output format:
```
#     Captured             Size      Path
--------------------------------------------------------------------------------
1     2026-05-03 14:22:33  1.2MB     C:\Users\You\Pictures\Screenshots\screenshot_20260503_142233.png
2     2026-05-03 13:10:01  845.3KB   C:\Users\You\Pictures\Screenshots\screenshot_20260503_131001.png
```

---

## Autostart (Windows)

```powershell
# Register to start automatically on Windows login
screensaver --install

# Remove from autostart
screensaver --uninstall
```

This writes/removes an entry at `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`.  
The `install.ps1` script also prompts to register autostart during initial setup.

---

## Config file

Screensaver reads settings from a YAML config file on startup. CLI flags always override config file values.

**File locations:**

| Platform | Path |
|----------|------|
| Windows  | `%APPDATA%\screensaver\config.yaml` |
| Other    | `~/.screensaver/config.yaml` |

**Supported keys and defaults:**

```yaml
hotkey: ctrl+shift+s          # Global hotkey for daemon mode
save_dir: ""                  # Auto-save directory (empty = ~/Pictures/Screenshots)
format: png                   # Output format: "png" or "jpeg"
quality: 90                   # JPEG quality 1–100 (ignored for PNG)
```

### Flags that override config

| Flag | Config key | Description |
|------|-----------|-------------|
| `--hotkey` | `hotkey` | Global hotkey combination |
| `--save-dir` | `save_dir` | Auto-save directory |
| `--format` | `format` | Output format (`png` or `jpeg`) |
| `--quality` | `quality` | JPEG quality 1–100 |
| `--config` | — | Path to a custom config file |

### Config subcommand

```powershell
# Print current effective config (file + defaults merged) as YAML
screensaver config show

# Write default config to the config file (errors if file already exists)
screensaver config init

# Overwrite an existing config with defaults
screensaver config init --force

# Print the config file path
screensaver config path
```

---

## Version

```powershell
screensaver --version
```
