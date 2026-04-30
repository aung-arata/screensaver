# Usage

## One-shot mode (capture full screen)

```bash
# Copy screenshot to clipboard
screensaver --once

# Save screenshot to a file
screensaver --once --output screenshot.png

# Open annotation editor after capture
screensaver --once --edit
```

---

## Interactive region selection (Windows only)

```bash
# Select a region and copy to clipboard
screensaver --select

# Select a region and save to a file
screensaver --select --output region.png

# Select a region and open the annotation editor
screensaver --select --edit
```

The screen dims and a crosshair cursor appears. Click and drag to select the
region you want to capture. Release the mouse to capture the selection, or
press **Escape** to cancel.

> **Note:** `--select` requires Windows (Win32 APIs). On non-Windows platforms
> the command returns a "not yet implemented" error (see `overlay_stub.go`).

---

## Annotation editor

Pass `--edit` together with `--once` or `--select` to open the post-capture
annotation editor:

```bash
screensaver --once --edit        # full-screen capture + editor
screensaver --select --edit      # region selection + editor (Windows only)
```

> **Note:** `--edit` must be combined with `--once` or `--select`.  Using
> `--edit` alone has no effect — the tool falls back to daemon mode and the
> flag is ignored.

The editor renders annotations (pen strokes, rectangles, arrows, text) onto
the captured image using `fogleman/gg` and saves the result to a timestamped
PNG file.  An interactive GUI toolbar (pen / rectangle / arrow / text tools,
Copy and Save buttons) is in progress.

---

## Background daemon

```bash
screensaver
# Press Ctrl+Shift+S to take a screenshot
# Press Ctrl+C to quit
```

---

## Custom hotkey

```bash
screensaver --hotkey "ctrl+shift+p"
```

---

## Version

```bash
screensaver --version
```

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

```bash
# Print current effective config (file + defaults merged) as YAML
screensaver config show

# Write default config to the config file (errors if file already exists)
screensaver config init

# Overwrite an existing config with defaults
screensaver config init --force

# Print the config file path
screensaver config path
```
