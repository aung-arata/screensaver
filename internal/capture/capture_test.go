package capture

import (
	"testing"
)

// ---------------------------------------------------------------------------
// NormalizeRegion
// ---------------------------------------------------------------------------

func TestNormalizeRegion_TopLeftToBottomRight(t *testing.T) {
	r := NormalizeRegion(10, 20, 100, 200)
	if r == nil {
		t.Fatal("expected non-nil region")
	}
	if r.X != 10 || r.Y != 20 || r.Width != 90 || r.Height != 180 {
		t.Errorf("got (%d, %d, %d, %d), want (10, 20, 90, 180)", r.X, r.Y, r.Width, r.Height)
	}
}

func TestNormalizeRegion_BottomRightToTopLeft(t *testing.T) {
	r := NormalizeRegion(100, 200, 10, 20)
	if r == nil {
		t.Fatal("expected non-nil region")
	}
	if r.X != 10 || r.Y != 20 || r.Width != 90 || r.Height != 180 {
		t.Errorf("got (%d, %d, %d, %d), want (10, 20, 90, 180)", r.X, r.Y, r.Width, r.Height)
	}
}

func TestNormalizeRegion_ZeroWidth(t *testing.T) {
	r := NormalizeRegion(50, 20, 50, 100)
	if r != nil {
		t.Errorf("expected nil for zero-width selection, got %+v", r)
	}
}

func TestNormalizeRegion_ZeroHeight(t *testing.T) {
	r := NormalizeRegion(20, 50, 100, 50)
	if r != nil {
		t.Errorf("expected nil for zero-height selection, got %+v", r)
	}
}

func TestNormalizeRegion_ZeroArea(t *testing.T) {
	r := NormalizeRegion(50, 50, 50, 50)
	if r != nil {
		t.Errorf("expected nil for zero-area selection, got %+v", r)
	}
}

func TestNormalizeRegion_SinglePixel(t *testing.T) {
	r := NormalizeRegion(5, 5, 6, 6)
	if r == nil {
		t.Fatal("expected non-nil region")
	}
	if r.X != 5 || r.Y != 5 || r.Width != 1 || r.Height != 1 {
		t.Errorf("got (%d, %d, %d, %d), want (5, 5, 1, 1)", r.X, r.Y, r.Width, r.Height)
	}
}

// ---------------------------------------------------------------------------
// CaptureRegion — validation
// ---------------------------------------------------------------------------

func TestCaptureRegion_ZeroWidth(t *testing.T) {
	_, err := CaptureRegion(Region{X: 0, Y: 0, Width: 0, Height: 100}, 0)
	if err == nil {
		t.Error("expected error for zero-width region")
	}
}

func TestCaptureRegion_ZeroHeight(t *testing.T) {
	_, err := CaptureRegion(Region{X: 0, Y: 0, Width: 100, Height: 0}, 0)
	if err == nil {
		t.Error("expected error for zero-height region")
	}
}

func TestCaptureRegion_NegativeWidth(t *testing.T) {
	_, err := CaptureRegion(Region{X: 0, Y: 0, Width: -10, Height: 100}, 0)
	if err == nil {
		t.Error("expected error for negative-width region")
	}
}

// ---------------------------------------------------------------------------
// MonitorCount
// ---------------------------------------------------------------------------

func TestMonitorCount(t *testing.T) {
	// In a headless CI environment this may return 0, which is fine.
	count := MonitorCount()
	if count < 0 {
		t.Errorf("expected non-negative monitor count, got %d", count)
	}
}
