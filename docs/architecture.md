# Architecture

## Directory Structure

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
