"""Screen capture utilities using mss and Pillow."""

from __future__ import annotations

from typing import Optional, Tuple

import mss
import mss.tools
from PIL import Image


# A region is expressed as (left, top, width, height) in screen pixels.
Region = Tuple[int, int, int, int]


def capture_full_screen(monitor: int = 1) -> Image.Image:
    """Capture the full contents of *monitor* (1-based index) and return a
    Pillow :class:`~PIL.Image.Image`.

    Args:
        monitor: 1-based monitor index.  ``1`` is the primary monitor.

    Returns:
        A ``PIL.Image`` in ``RGB`` mode.
    """
    with mss.mss() as sct:
        raw = sct.grab(sct.monitors[monitor])
    return Image.frombytes("RGB", raw.size, raw.bgra, "raw", "BGRX")


def capture_region(region: Region, monitor: int = 1) -> Image.Image:
    """Capture a rectangular *region* of *monitor* and return a Pillow image.

    Args:
        region: ``(left, top, width, height)`` in screen pixels, relative to
                the top-left corner of *monitor*.
        monitor: 1-based monitor index.

    Returns:
        A ``PIL.Image`` in ``RGB`` mode.

    Raises:
        ValueError: If *width* or *height* is not positive.
    """
    left, top, width, height = region
    if width <= 0 or height <= 0:
        raise ValueError(
            f"Region width and height must be positive, got ({width}, {height})"
        )

    with mss.mss() as sct:
        mon = sct.monitors[monitor]
        bbox = {
            "left": mon["left"] + left,
            "top": mon["top"] + top,
            "width": width,
            "height": height,
        }
        raw = sct.grab(bbox)
    return Image.frombytes("RGB", raw.size, raw.bgra, "raw", "BGRX")


def get_monitor_info(monitor: int = 1) -> dict:
    """Return geometry information for *monitor*.

    Returns a dict with keys ``left``, ``top``, ``width``, ``height``.
    """
    with mss.mss() as sct:
        return dict(sct.monitors[monitor])


def get_monitor_count() -> int:
    """Return the number of available monitors (excluding the virtual 'all' entry)."""
    with mss.mss() as sct:
        # monitors[0] is the combined virtual screen; real monitors start at index 1
        return max(0, len(sct.monitors) - 1)


def normalize_region(
    x1: int, y1: int, x2: int, y2: int
) -> Optional[Region]:
    """Convert two arbitrary corner points into a normalised ``(left, top, width, height)``
    region tuple, ensuring *width* and *height* are positive.

    Returns ``None`` if the two points are identical (zero-area selection).
    """
    left = min(x1, x2)
    top = min(y1, y2)
    width = abs(x2 - x1)
    height = abs(y2 - y1)
    if width == 0 or height == 0:
        return None
    return (left, top, width, height)
