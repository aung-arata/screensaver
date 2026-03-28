"""Fullscreen selection overlay window.

The :class:`SelectionOverlay` darkens the entire screen, then lets the user
drag a rubber-band rectangle.  When the mouse button is released the selected
region (as a Pillow :class:`~PIL.Image.Image`) is returned via an
``on_capture`` callback and the window destroys itself.

Pressing ``Escape`` cancels the selection without capturing anything.
"""

from __future__ import annotations

import tkinter as tk
from typing import Callable, Optional

from PIL import Image, ImageTk

from .capture import capture_full_screen, normalize_region


# Semi-transparent overlay darkness (0–255).  128 ≈ 50% grey.
_OVERLAY_ALPHA = 128
# Colour of the selection rectangle border.
_SELECTION_COLOUR = "#00AAFF"


class SelectionOverlay:
    """Fullscreen overlay that lets the user drag a capture rectangle.

    Args:
        on_capture:  Callback called with the cropped :class:`~PIL.Image.Image`
                     when a region is successfully selected.  Called with
                     ``None`` if the user cancels.
        monitor:     1-based monitor index to capture from.
    """

    def __init__(
        self,
        on_capture: Callable[[Optional[Image.Image]], None],
        monitor: int = 1,
    ) -> None:
        self._on_capture = on_capture
        self._monitor = monitor

        # 1. Grab the full-screen image BEFORE the overlay appears.
        self._screenshot = capture_full_screen(monitor)

        # 2. Build the tkinter window.
        self._root = tk.Tk()
        self._root.withdraw()
        self._setup_window()
        self._setup_canvas()
        self._bind_events()
        self._root.deiconify()

    # ------------------------------------------------------------------
    # Setup helpers
    # ------------------------------------------------------------------

    def _setup_window(self) -> None:
        root = self._root
        root.overrideredirect(True)  # no window decorations
        root.attributes("-topmost", True)
        root.attributes("-fullscreen", True)
        root.config(cursor="crosshair")
        # On Linux with a compositor the window can be made translucent;
        # on Windows use "-alpha" attribute instead.
        try:
            root.attributes("-alpha", 0.85)
        except tk.TclError:
            pass

    def _setup_canvas(self) -> None:
        w = self._root.winfo_screenwidth()
        h = self._root.winfo_screenheight()

        self._canvas = tk.Canvas(
            self._root,
            width=w,
            height=h,
            highlightthickness=0,
            cursor="crosshair",
        )
        self._canvas.pack(fill="both", expand=True)

        # Darken the screenshot and use it as the canvas background.
        darkened = self._screenshot.copy().convert("RGBA")
        overlay_layer = Image.new("RGBA", darkened.size, (0, 0, 0, _OVERLAY_ALPHA))
        darkened = Image.alpha_composite(darkened, overlay_layer)
        self._bg_photo = ImageTk.PhotoImage(darkened)
        self._canvas.create_image(0, 0, anchor="nw", image=self._bg_photo)

        # Placeholder for the "bright" selection rectangle image.
        self._selection_image_id: Optional[int] = None
        self._rect_id: Optional[int] = None

        # Drag state.
        self._start_x: Optional[int] = None
        self._start_y: Optional[int] = None

    def _bind_events(self) -> None:
        self._canvas.bind("<ButtonPress-1>", self._on_press)
        self._canvas.bind("<B1-Motion>", self._on_drag)
        self._canvas.bind("<ButtonRelease-1>", self._on_release)
        self._root.bind("<Escape>", self._on_cancel)

    # ------------------------------------------------------------------
    # Event handlers
    # ------------------------------------------------------------------

    def _on_press(self, event: tk.Event) -> None:
        self._start_x = event.x
        self._start_y = event.y
        # Remove any previous rubber-band artefacts.
        if self._rect_id is not None:
            self._canvas.delete(self._rect_id)
            self._rect_id = None
        if self._selection_image_id is not None:
            self._canvas.delete(self._selection_image_id)
            self._selection_image_id = None

    def _on_drag(self, event: tk.Event) -> None:
        if self._start_x is None:
            return

        x1, y1 = self._start_x, self._start_y
        x2, y2 = event.x, event.y

        # Draw the unmasked (bright) screenshot region inside the selection.
        region = normalize_region(x1, y1, x2, y2)
        if region:
            left, top, width, height = region
            crop = self._screenshot.crop((left, top, left + width, top + height))
            self._selection_photo = ImageTk.PhotoImage(crop)
            if self._selection_image_id is not None:
                self._canvas.delete(self._selection_image_id)
            self._selection_image_id = self._canvas.create_image(
                left, top, anchor="nw", image=self._selection_photo
            )

        # Draw a coloured border around the selection.
        if self._rect_id is not None:
            self._canvas.delete(self._rect_id)
        self._rect_id = self._canvas.create_rectangle(
            x1, y1, x2, y2,
            outline=_SELECTION_COLOUR,
            width=2,
        )

    def _on_release(self, event: tk.Event) -> None:
        if self._start_x is None:
            self._root.destroy()
            self._on_capture(None)
            return

        region = normalize_region(
            self._start_x, self._start_y, event.x, event.y
        )
        self._root.destroy()

        if region is None:
            self._on_capture(None)
            return

        left, top, width, height = region
        cropped = self._screenshot.crop((left, top, left + width, top + height))
        self._on_capture(cropped)

    def _on_cancel(self, event: tk.Event) -> None:
        self._root.destroy()
        self._on_capture(None)

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def run(self) -> None:
        """Enter the tkinter event loop (blocking until capture or cancel)."""
        self._root.mainloop()
