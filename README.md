# Screensaver 🖼️

A **lightweight, portable screenshot tool** for Windows (and Linux/macOS) inspired by [Lightshot](https://app.prntscr.com/).

---

## Features

| Feature | Status |
|---|---|
| Full-screen selection overlay | ✅ |
| Rubber-band region selection | ✅ |
| Post-capture editor | ✅ |
| Pen / freehand drawing | ✅ |
| Rectangle annotation | ✅ |
| Arrow annotation | ✅ |
| Text annotation | ✅ |
| Undo | ✅ |
| Copy to clipboard | ✅ |
| Save to PNG / JPEG | ✅ |
| Global hotkey daemon | ✅ |
| Multi-monitor support | ✅ |

---

## Requirements

- Python 3.9+
- `tkinter` (ships with most Python distributions; on Ubuntu: `sudo apt install python3-tk`)

Install Python dependencies:

```bash
pip install -r requirements.txt
```

---

## Installation

```bash
pip install .
```

Or run directly from source without installing:

```bash
python -m screensaver
```

---

## Usage

### Background daemon (recommended)

Launch the tool in the background.  Press **Ctrl+Shift+S** at any time to open
the selection overlay:

```bash
screensaver
# or
python -m screensaver
```

### One-shot mode

Capture a single screenshot and exit:

```bash
screensaver --once
python -m screensaver --once
```

### Custom hotkey

```bash
screensaver --hotkey "<ctrl>+<shift>+p"
```

---

## How it works

1. Press the hotkey (or run `--once`).
2. The screen dims and a crosshair cursor appears.
3. Click and drag to select the region you want to capture.
4. Release the mouse — the **Editor** window opens with the captured region.
5. Use the toolbar to annotate the image (pen, rectangle, arrow, text).
6. Click **Copy** to copy the image to your clipboard, or **Save** to write it
   to disk.
7. Press **Escape** at any time during selection to cancel.

---

## Project structure

```
src/
└── screensaver/
    ├── __init__.py       # Package metadata
    ├── __main__.py       # python -m screensaver entry point
    ├── main.py           # CLI argument parsing + hotkey daemon
    ├── capture.py        # Screen capture (mss + Pillow)
    ├── overlay.py        # Fullscreen selection overlay (tkinter)
    ├── editor.py         # Post-capture annotation editor (tkinter)
    └── utils.py          # Save & clipboard helpers
tests/
├── test_capture.py
└── test_utils.py
```

---

## Running tests

```bash
pip install pytest
pytest
```

---

## License

MIT
