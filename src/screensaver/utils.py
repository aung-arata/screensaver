"""File-save and clipboard utilities for captured screenshots."""

from __future__ import annotations

import io
import os
import platform
import subprocess
import tempfile
from datetime import datetime
from pathlib import Path
from typing import Optional

from PIL import Image


def default_save_directory() -> Path:
    """Return a platform-appropriate default directory for saved screenshots.

    On Windows this is ``%USERPROFILE%\\Pictures\\Screenshots``;
    on macOS / Linux it falls back to ``~/Pictures/Screenshots``.
    """
    pictures = Path.home() / "Pictures" / "Screenshots"
    pictures.mkdir(parents=True, exist_ok=True)
    return pictures


def generate_filename(directory: Optional[Path] = None, fmt: str = "png") -> Path:
    """Generate a timestamped filename inside *directory*.

    Args:
        directory: Target folder.  Defaults to :func:`default_save_directory`.
        fmt: File extension / format (``"png"``, ``"jpg"``).

    Returns:
        An absolute :class:`~pathlib.Path` that does not yet exist.
    """
    if directory is None:
        directory = default_save_directory()
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    return Path(directory) / f"screenshot_{timestamp}.{fmt}"


def save_image(image: Image.Image, path: Optional[Path] = None, fmt: str = "png") -> Path:
    """Save *image* to *path* (or an auto-generated path) and return the final path.

    Args:
        image: A Pillow :class:`~PIL.Image.Image`.
        path:  Destination file path.  Auto-generated when ``None``.
        fmt:   Image format (``"png"``, ``"jpeg"``).

    Returns:
        The absolute path where the image was saved.
    """
    if path is None:
        path = generate_filename(fmt=fmt)
    path = Path(path)
    image.save(str(path), fmt.upper())
    return path


def copy_to_clipboard(image: Image.Image) -> bool:
    """Copy *image* to the system clipboard.

    Supports Windows (via ``win32clipboard``), macOS (via ``pbcopy``), and
    Linux X11 (via ``xclip`` or ``xsel``).

    Returns:
        ``True`` on success, ``False`` if no clipboard backend is available.
    """
    system = platform.system()

    if system == "Windows":
        return _copy_clipboard_windows(image)
    elif system == "Darwin":
        return _copy_clipboard_macos(image)
    else:
        return _copy_clipboard_linux(image)


# ---------------------------------------------------------------------------
# Platform-specific helpers
# ---------------------------------------------------------------------------

def _copy_clipboard_windows(image: Image.Image) -> bool:
    try:
        import win32clipboard  # type: ignore[import]
        from io import BytesIO
        import win32con

        output = BytesIO()
        image.convert("RGB").save(output, "BMP")
        data = output.getvalue()[14:]  # strip BMP file header
        output.close()

        win32clipboard.OpenClipboard()
        win32clipboard.EmptyClipboard()
        win32clipboard.SetClipboardData(win32con.CF_DIB, data)
        win32clipboard.CloseClipboard()
        return True
    except Exception:
        return False


def _copy_clipboard_macos(image: Image.Image) -> bool:
    try:
        buf = io.BytesIO()
        image.save(buf, "PNG")
        proc = subprocess.Popen(
            ["osascript", "-e",
             'set the clipboard to (read (POSIX file "/dev/stdin") as «class PNGf»)'],
            stdin=subprocess.PIPE,
        )
        proc.communicate(buf.getvalue())
        return proc.returncode == 0
    except Exception:
        return _copy_clipboard_xclip(image)


def _copy_clipboard_linux(image: Image.Image) -> bool:
    return _copy_clipboard_xclip(image)


def _copy_clipboard_xclip(image: Image.Image) -> bool:
    """Try xclip, then xsel, then tkinter as clipboard backends."""
    buf = io.BytesIO()
    image.save(buf, "PNG")
    png_bytes = buf.getvalue()

    for cmd in (
        ["xclip", "-selection", "clipboard", "-t", "image/png"],
        ["xsel", "--clipboard", "--input"],
    ):
        try:
            proc = subprocess.Popen(cmd, stdin=subprocess.PIPE)
            proc.communicate(png_bytes)
            if proc.returncode == 0:
                return True
        except FileNotFoundError:
            continue

    # Last resort: tkinter clipboard (text only – save path instead)
    try:
        import tkinter as tk
        with tempfile.NamedTemporaryFile(suffix=".png", delete=False) as tmp:
            image.save(tmp.name, "PNG")
            root = tk.Tk()
            root.withdraw()
            root.clipboard_clear()
            root.clipboard_append(tmp.name)
            root.update()
            root.destroy()
        return True
    except Exception:
        return False
