"""Tests for the capture module."""

from __future__ import annotations

import pytest
from unittest.mock import MagicMock, patch

from screensaver.capture import (
    capture_full_screen,
    capture_region,
    get_monitor_count,
    get_monitor_info,
    normalize_region,
)


# ---------------------------------------------------------------------------
# normalize_region
# ---------------------------------------------------------------------------

class TestNormalizeRegion:
    def test_top_left_to_bottom_right(self):
        result = normalize_region(10, 20, 100, 200)
        assert result == (10, 20, 90, 180)

    def test_bottom_right_to_top_left(self):
        """Dragging in the opposite direction should still give a valid region."""
        result = normalize_region(100, 200, 10, 20)
        assert result == (10, 20, 90, 180)

    def test_zero_width_returns_none(self):
        assert normalize_region(50, 20, 50, 100) is None

    def test_zero_height_returns_none(self):
        assert normalize_region(20, 50, 100, 50) is None

    def test_zero_area_returns_none(self):
        assert normalize_region(50, 50, 50, 50) is None

    def test_single_pixel_region(self):
        result = normalize_region(5, 5, 6, 6)
        assert result == (5, 5, 1, 1)


# ---------------------------------------------------------------------------
# capture_region – validation
# ---------------------------------------------------------------------------

class TestCaptureRegion:
    def test_raises_on_zero_width(self):
        with pytest.raises(ValueError, match="positive"):
            capture_region((0, 0, 0, 100))

    def test_raises_on_zero_height(self):
        with pytest.raises(ValueError, match="positive"):
            capture_region((0, 0, 100, 0))

    def test_raises_on_negative_width(self):
        with pytest.raises(ValueError, match="positive"):
            capture_region((0, 0, -10, 100))

    @patch("screensaver.capture.mss.mss")
    def test_returns_rgb_image(self, mock_mss_cls):
        """capture_region should return an RGB Pillow image."""
        from PIL import Image

        # Build a fake mss context.
        fake_grab = MagicMock()
        fake_grab.size = (10, 10)
        fake_grab.bgra = b"\x00" * (10 * 10 * 4)

        fake_sct = MagicMock()
        fake_sct.__enter__ = lambda s: s
        fake_sct.__exit__ = MagicMock(return_value=False)
        fake_sct.monitors = [
            {"left": 0, "top": 0, "width": 1920, "height": 1080},  # virtual
            {"left": 0, "top": 0, "width": 1920, "height": 1080},  # monitor 1
        ]
        fake_sct.grab.return_value = fake_grab
        mock_mss_cls.return_value = fake_sct

        img = capture_region((0, 0, 10, 10))
        assert img.mode == "RGB"
        assert img.size == (10, 10)


# ---------------------------------------------------------------------------
# capture_full_screen
# ---------------------------------------------------------------------------

class TestCaptureFullScreen:
    @patch("screensaver.capture.mss.mss")
    def test_returns_rgb_image(self, mock_mss_cls):
        from PIL import Image

        fake_grab = MagicMock()
        fake_grab.size = (1920, 1080)
        fake_grab.bgra = b"\x00" * (1920 * 1080 * 4)

        fake_sct = MagicMock()
        fake_sct.__enter__ = lambda s: s
        fake_sct.__exit__ = MagicMock(return_value=False)
        fake_sct.monitors = [
            {"left": 0, "top": 0, "width": 1920, "height": 1080},
            {"left": 0, "top": 0, "width": 1920, "height": 1080},
        ]
        fake_sct.grab.return_value = fake_grab
        mock_mss_cls.return_value = fake_sct

        img = capture_full_screen()
        assert img.mode == "RGB"
        assert img.size == (1920, 1080)


# ---------------------------------------------------------------------------
# get_monitor_info / get_monitor_count
# ---------------------------------------------------------------------------

class TestMonitorHelpers:
    @patch("screensaver.capture.mss.mss")
    def test_get_monitor_info(self, mock_mss_cls):
        fake_sct = MagicMock()
        fake_sct.__enter__ = lambda s: s
        fake_sct.__exit__ = MagicMock(return_value=False)
        fake_sct.monitors = [
            {"left": 0, "top": 0, "width": 3840, "height": 1080},
            {"left": 0, "top": 0, "width": 1920, "height": 1080},
            {"left": 1920, "top": 0, "width": 1920, "height": 1080},
        ]
        mock_mss_cls.return_value = fake_sct

        info = get_monitor_info(1)
        assert info["width"] == 1920
        assert info["height"] == 1080

    @patch("screensaver.capture.mss.mss")
    def test_get_monitor_count(self, mock_mss_cls):
        fake_sct = MagicMock()
        fake_sct.__enter__ = lambda s: s
        fake_sct.__exit__ = MagicMock(return_value=False)
        fake_sct.monitors = [
            {"left": 0, "top": 0, "width": 3840, "height": 1080},
            {"left": 0, "top": 0, "width": 1920, "height": 1080},
            {"left": 1920, "top": 0, "width": 1920, "height": 1080},
        ]
        mock_mss_cls.return_value = fake_sct

        count = get_monitor_count()
        assert count == 2
