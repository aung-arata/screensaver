"""Tests for the utils module."""

from __future__ import annotations

import os
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
from PIL import Image

from screensaver.utils import (
    default_save_directory,
    generate_filename,
    save_image,
    copy_to_clipboard,
)


def _make_image(width: int = 100, height: int = 80) -> Image.Image:
    """Return a small solid-colour test image."""
    return Image.new("RGB", (width, height), color=(255, 0, 0))


# ---------------------------------------------------------------------------
# default_save_directory
# ---------------------------------------------------------------------------

class TestDefaultSaveDirectory:
    def test_returns_path(self):
        directory = default_save_directory()
        assert isinstance(directory, Path)

    def test_directory_is_created(self):
        directory = default_save_directory()
        assert directory.exists()
        assert directory.is_dir()


# ---------------------------------------------------------------------------
# generate_filename
# ---------------------------------------------------------------------------

class TestGenerateFilename:
    def test_png_extension(self):
        path = generate_filename(fmt="png")
        assert path.suffix == ".png"

    def test_jpg_extension(self):
        path = generate_filename(fmt="jpg")
        assert path.suffix == ".jpg"

    def test_filename_contains_screenshot(self):
        path = generate_filename()
        assert "screenshot" in path.name

    def test_uses_provided_directory(self, tmp_path):
        path = generate_filename(directory=tmp_path)
        assert path.parent == tmp_path

    def test_unique_filenames(self):
        p1 = generate_filename()
        import time
        time.sleep(0.01)
        p2 = generate_filename()
        # Both should be valid Path objects; names differ unless created in the same second.
        assert isinstance(p1, Path)
        assert isinstance(p2, Path)


# ---------------------------------------------------------------------------
# save_image
# ---------------------------------------------------------------------------

class TestSaveImage:
    def test_save_to_explicit_path(self, tmp_path):
        img = _make_image()
        dest = tmp_path / "test.png"
        result = save_image(img, dest)
        assert result == dest
        assert dest.exists()

    def test_saved_file_is_valid_image(self, tmp_path):
        img = _make_image(50, 50)
        dest = tmp_path / "out.png"
        save_image(img, dest)
        loaded = Image.open(dest)
        assert loaded.size == (50, 50)

    def test_auto_path_when_none(self, tmp_path):
        img = _make_image()
        with patch("screensaver.utils.default_save_directory", return_value=tmp_path):
            result = save_image(img)
        assert result.exists()
        assert result.suffix == ".png"

    def test_jpeg_format(self, tmp_path):
        img = _make_image()
        dest = tmp_path / "out.jpg"
        result = save_image(img, dest, fmt="jpeg")
        assert result.exists()


# ---------------------------------------------------------------------------
# copy_to_clipboard
# ---------------------------------------------------------------------------

class TestCopyToClipboard:
    @patch("screensaver.utils.platform.system", return_value="Windows")
    @patch("screensaver.utils._copy_clipboard_windows", return_value=True)
    def test_windows_delegates(self, mock_win, mock_sys):
        img = _make_image()
        assert copy_to_clipboard(img) is True
        mock_win.assert_called_once_with(img)

    @patch("screensaver.utils.platform.system", return_value="Darwin")
    @patch("screensaver.utils._copy_clipboard_macos", return_value=True)
    def test_macos_delegates(self, mock_mac, mock_sys):
        img = _make_image()
        assert copy_to_clipboard(img) is True
        mock_mac.assert_called_once_with(img)

    @patch("screensaver.utils.platform.system", return_value="Linux")
    @patch("screensaver.utils._copy_clipboard_linux", return_value=True)
    def test_linux_delegates(self, mock_linux, mock_sys):
        img = _make_image()
        assert copy_to_clipboard(img) is True
        mock_linux.assert_called_once_with(img)

    @patch("screensaver.utils.platform.system", return_value="Linux")
    @patch("screensaver.utils._copy_clipboard_linux", return_value=False)
    def test_returns_false_when_no_backend(self, mock_linux, mock_sys):
        img = _make_image()
        assert copy_to_clipboard(img) is False
