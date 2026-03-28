"""Application entry point.

Usage
-----
Run directly::

    python -m screensaver

Or after installation::

    screensaver

The tool registers a global hotkey (``Ctrl+Shift+S`` by default) and waits in
the background.  Press the hotkey to launch the selection overlay.

For a one-shot capture without the background daemon you can also pass
``--once``::

    python -m screensaver --once
"""

from __future__ import annotations

import argparse
import sys
import threading
from typing import Optional

from PIL import Image

from .overlay import SelectionOverlay
from .editor import EditorWindow

try:
    from pynput import keyboard as _kb
    _PYNPUT_AVAILABLE = True
except ImportError:
    _PYNPUT_AVAILABLE = False


# Default global hotkey combination.
_HOTKEY_STR = "<ctrl>+<shift>+s"


def _launch_capture() -> None:
    """Run the selection overlay and, on success, open the editor."""

    def _handle_capture(image: Optional[Image.Image]) -> None:
        if image is not None:
            editor = EditorWindow(image)
            editor.run()

    overlay = SelectionOverlay(on_capture=_handle_capture)
    overlay.run()


def _start_hotkey_listener(hotkey_str: str = _HOTKEY_STR) -> None:
    """Register a global hotkey that calls :func:`_launch_capture`.

    The listener runs in a background daemon thread so that the main thread
    can block on ``input()`` (or just sleep) as a keep-alive.
    """
    if not _PYNPUT_AVAILABLE:
        print(
            "[screensaver] pynput is not installed – global hotkey is unavailable.\n"
            "Install it with:  pip install pynput\n"
            "Running in --once mode instead.",
            file=sys.stderr,
        )
        _launch_capture()
        return

    def _on_activate():
        # Run the capture in a separate thread to avoid blocking the listener.
        t = threading.Thread(target=_launch_capture, daemon=True)
        t.start()

    hotkey = _kb.GlobalHotKeys({hotkey_str: _on_activate})
    hotkey.daemon = True
    hotkey.start()

    print(
        f"[screensaver] Running in the background.\n"
        f"  Press {hotkey_str.upper()} to take a screenshot.\n"
        f"  Press Ctrl+C to quit."
    )
    try:
        hotkey.join()
    except KeyboardInterrupt:
        print("\n[screensaver] Bye!")


def main(argv: Optional[list[str]] = None) -> None:
    """CLI entry point."""
    parser = argparse.ArgumentParser(
        prog="screensaver",
        description="Lightweight screenshot tool – like Lightshot for your desktop.",
    )
    parser.add_argument(
        "--once",
        action="store_true",
        help="Capture one screenshot and exit (no background daemon).",
    )
    parser.add_argument(
        "--hotkey",
        default=_HOTKEY_STR,
        metavar="COMBO",
        help=(
            f"Global hotkey combination (default: {_HOTKEY_STR!r}).  "
            "Uses pynput syntax, e.g. '<ctrl>+<shift>+s'."
        ),
    )
    args = parser.parse_args(argv)

    if args.once:
        _launch_capture()
    else:
        _start_hotkey_listener(args.hotkey)


if __name__ == "__main__":
    main()
