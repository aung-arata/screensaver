"""Post-capture editor window.

Displays the captured screenshot in a resizable window with a toolbar that
offers basic annotation tools (pen, rectangle, arrow, text) plus *Copy* and
*Save* actions – mirroring the core Lightshot workflow.
"""

from __future__ import annotations

import tkinter as tk
from tkinter import colorchooser, filedialog, messagebox
from typing import Optional

from PIL import Image, ImageDraw, ImageTk

from .utils import copy_to_clipboard, save_image


# Default pen width and colour.
_DEFAULT_PEN_WIDTH = 2
_DEFAULT_PEN_COLOUR = "#FF0000"


class EditorWindow:
    """A simple post-capture image editor.

    Args:
        image:   The cropped screenshot to edit.
        on_done: Optional callback invoked when the window is closed.
    """

    def __init__(self, image: Image.Image, on_done: Optional[Callable] = None) -> None:
        self._original = image.copy()
        self._image = image.copy()
        self._on_done = on_done

        self._pen_colour = _DEFAULT_PEN_COLOUR
        self._pen_width = _DEFAULT_PEN_WIDTH
        self._tool = "pen"

        # Temporary drawing state.
        self._last_x: Optional[int] = None
        self._last_y: Optional[int] = None
        self._drag_start: Optional[tuple[int, int]] = None
        self._draw_preview_id: Optional[int] = None

        self._root = tk.Tk()
        self._root.title("Screensaver – Editor")
        self._root.resizable(True, True)
        self._root.protocol("WM_DELETE_WINDOW", self._on_close)

        self._build_toolbar()
        self._build_canvas()

    # ------------------------------------------------------------------
    # UI construction
    # ------------------------------------------------------------------

    def _build_toolbar(self) -> None:
        bar = tk.Frame(self._root, bd=1, relief=tk.RAISED)
        bar.pack(side=tk.TOP, fill=tk.X)

        buttons = [
            ("✏ Pen", "pen"),
            ("▭ Rect", "rect"),
            ("➜ Arrow", "arrow"),
            ("T Text", "text"),
        ]
        for label, tool in buttons:
            tk.Button(
                bar, text=label, relief=tk.FLAT, padx=6,
                command=lambda t=tool: self._set_tool(t),
            ).pack(side=tk.LEFT)

        tk.Separator(bar, orient=tk.VERTICAL).pack(side=tk.LEFT, fill=tk.Y, padx=4)

        tk.Button(bar, text="🎨 Colour", relief=tk.FLAT, padx=6,
                  command=self._pick_colour).pack(side=tk.LEFT)

        tk.Label(bar, text="Width:").pack(side=tk.LEFT)
        self._width_var = tk.IntVar(value=_DEFAULT_PEN_WIDTH)
        tk.Spinbox(bar, from_=1, to=20, width=3, textvariable=self._width_var,
                   command=self._update_width).pack(side=tk.LEFT)

        tk.Separator(bar, orient=tk.VERTICAL).pack(side=tk.LEFT, fill=tk.Y, padx=4)

        tk.Button(bar, text="↩ Undo", relief=tk.FLAT, padx=6,
                  command=self._undo).pack(side=tk.LEFT)

        tk.Separator(bar, orient=tk.VERTICAL).pack(side=tk.LEFT, fill=tk.Y, padx=4)

        tk.Button(bar, text="📋 Copy", relief=tk.FLAT, padx=6,
                  command=self._copy).pack(side=tk.LEFT)
        tk.Button(bar, text="💾 Save", relief=tk.FLAT, padx=6,
                  command=self._save).pack(side=tk.LEFT)

        # Status label on the right.
        self._status_var = tk.StringVar(value="Tool: pen")
        tk.Label(bar, textvariable=self._status_var, anchor="e").pack(
            side=tk.RIGHT, padx=8
        )

    def _build_canvas(self) -> None:
        frame = tk.Frame(self._root)
        frame.pack(fill=tk.BOTH, expand=True)

        self._canvas = tk.Canvas(
            frame,
            width=self._image.width,
            height=self._image.height,
            cursor="crosshair",
        )
        self._canvas.pack(fill=tk.BOTH, expand=True)

        self._photo = ImageTk.PhotoImage(self._image)
        self._canvas_image_id = self._canvas.create_image(
            0, 0, anchor="nw", image=self._photo
        )

        self._canvas.bind("<ButtonPress-1>", self._on_press)
        self._canvas.bind("<B1-Motion>", self._on_drag)
        self._canvas.bind("<ButtonRelease-1>", self._on_release)

        # History stack for undo (stores PIL images).
        self._history: list[Image.Image] = []

    # ------------------------------------------------------------------
    # Tool selection helpers
    # ------------------------------------------------------------------

    def _set_tool(self, tool: str) -> None:
        self._tool = tool
        self._status_var.set(f"Tool: {tool}")
        cursor = "xterm" if tool == "text" else "crosshair"
        self._canvas.config(cursor=cursor)

    def _pick_colour(self) -> None:
        colour = colorchooser.askcolor(color=self._pen_colour, title="Pick colour")
        if colour and colour[1]:
            self._pen_colour = colour[1]

    def _update_width(self) -> None:
        try:
            self._pen_width = int(self._width_var.get())
        except (ValueError, tk.TclError):
            pass

    # ------------------------------------------------------------------
    # Drawing event handlers
    # ------------------------------------------------------------------

    def _on_press(self, event: tk.Event) -> None:
        self._history.append(self._image.copy())
        self._last_x = event.x
        self._last_y = event.y
        self._drag_start = (event.x, event.y)

        if self._tool == "text":
            self._prompt_text(event.x, event.y)

    def _on_drag(self, event: tk.Event) -> None:
        if self._tool == "pen":
            if self._last_x is not None:
                draw = ImageDraw.Draw(self._image)
                draw.line(
                    [self._last_x, self._last_y, event.x, event.y],
                    fill=self._pen_colour,
                    width=self._pen_width,
                )
                self._last_x, self._last_y = event.x, event.y
                self._refresh_canvas()

        elif self._tool in ("rect", "arrow"):
            # Show a live preview on the canvas (don't commit to the image yet).
            if self._draw_preview_id is not None:
                self._canvas.delete(self._draw_preview_id)
            x0, y0 = self._drag_start
            if self._tool == "rect":
                self._draw_preview_id = self._canvas.create_rectangle(
                    x0, y0, event.x, event.y,
                    outline=self._pen_colour,
                    width=self._pen_width,
                )
            else:
                self._draw_preview_id = self._canvas.create_line(
                    x0, y0, event.x, event.y,
                    fill=self._pen_colour,
                    width=self._pen_width,
                    arrow=tk.LAST,
                )

    def _on_release(self, event: tk.Event) -> None:
        if self._draw_preview_id is not None:
            self._canvas.delete(self._draw_preview_id)
            self._draw_preview_id = None

        if self._tool in ("rect", "arrow") and self._drag_start:
            x0, y0 = self._drag_start
            draw = ImageDraw.Draw(self._image)
            if self._tool == "rect":
                draw.rectangle(
                    [x0, y0, event.x, event.y],
                    outline=self._pen_colour,
                    width=self._pen_width,
                )
            else:
                # Arrow: draw a line with a small arrowhead triangle.
                _draw_arrow(draw, x0, y0, event.x, event.y,
                            self._pen_colour, self._pen_width)
            self._refresh_canvas()

        self._last_x = None
        self._last_y = None
        self._drag_start = None

    # ------------------------------------------------------------------
    # Text tool
    # ------------------------------------------------------------------

    def _prompt_text(self, x: int, y: int) -> None:
        dialog = tk.Toplevel(self._root)
        dialog.title("Add text")
        dialog.resizable(False, False)
        dialog.transient(self._root)
        dialog.grab_set()

        tk.Label(dialog, text="Text:").pack(padx=8, pady=(8, 0))
        entry = tk.Entry(dialog, width=30)
        entry.pack(padx=8, pady=4)
        entry.focus_set()

        def _apply() -> None:
            text = entry.get()
            if text:
                draw = ImageDraw.Draw(self._image)
                draw.text((x, y), text, fill=self._pen_colour)
                self._refresh_canvas()
            dialog.destroy()

        tk.Button(dialog, text="OK", command=_apply).pack(pady=(0, 8))
        dialog.bind("<Return>", lambda _e: _apply())

    # ------------------------------------------------------------------
    # Image / canvas helpers
    # ------------------------------------------------------------------

    def _refresh_canvas(self) -> None:
        self._photo = ImageTk.PhotoImage(self._image)
        self._canvas.itemconfig(self._canvas_image_id, image=self._photo)

    def _undo(self) -> None:
        if self._history:
            self._image = self._history.pop()
            self._refresh_canvas()

    # ------------------------------------------------------------------
    # Toolbar actions
    # ------------------------------------------------------------------

    def _copy(self) -> None:
        ok = copy_to_clipboard(self._image)
        msg = "Image copied to clipboard!" if ok else (
            "Clipboard copy is not available on this platform.\n"
            "Install xclip or xsel to enable clipboard support."
        )
        messagebox.showinfo("Copy", msg, parent=self._root)

    def _save(self) -> None:
        path = filedialog.asksaveasfilename(
            defaultextension=".png",
            filetypes=[("PNG image", "*.png"), ("JPEG image", "*.jpg"), ("All files", "*.*")],
            title="Save screenshot",
        )
        if path:
            saved = save_image(self._image, path)
            messagebox.showinfo("Saved", f"Screenshot saved to:\n{saved}", parent=self._root)

    def _on_close(self) -> None:
        self._root.destroy()
        if self._on_done:
            self._on_done()

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def run(self) -> None:
        """Enter the tkinter event loop (blocking)."""
        self._root.mainloop()


# ---------------------------------------------------------------------------
# Drawing helpers
# ---------------------------------------------------------------------------

def _draw_arrow(
    draw: ImageDraw.ImageDraw,
    x0: int, y0: int, x1: int, y1: int,
    colour: str, width: int,
) -> None:
    """Draw a line with a filled arrowhead at ``(x1, y1)``."""
    import math

    draw.line([x0, y0, x1, y1], fill=colour, width=width)

    arrow_length = max(10, width * 4)
    angle = math.atan2(y1 - y0, x1 - x0)
    spread = math.pi / 6  # 30 degrees

    left_x = x1 - arrow_length * math.cos(angle - spread)
    left_y = y1 - arrow_length * math.sin(angle - spread)
    right_x = x1 - arrow_length * math.cos(angle + spread)
    right_y = y1 - arrow_length * math.sin(angle + spread)

    draw.polygon(
        [(x1, y1), (int(left_x), int(left_y)), (int(right_x), int(right_y))],
        fill=colour,
    )
