package overlay

import (
	"image"
	"testing"
)

// ---------------------------------------------------------------------------
// normalizeRect
// ---------------------------------------------------------------------------

func TestNormalizeRect_TopLeftToBottomRight(t *testing.T) {
	r := normalizeRect(10, 20, 100, 200)
	want := image.Rect(10, 20, 100, 200)
	if r != want {
		t.Errorf("got %v, want %v", r, want)
	}
}

func TestNormalizeRect_BottomRightToTopLeft(t *testing.T) {
	r := normalizeRect(100, 200, 10, 20)
	want := image.Rect(10, 20, 100, 200)
	if r != want {
		t.Errorf("got %v, want %v", r, want)
	}
}

func TestNormalizeRect_TopRightToBottomLeft(t *testing.T) {
	r := normalizeRect(100, 20, 10, 200)
	want := image.Rect(10, 20, 100, 200)
	if r != want {
		t.Errorf("got %v, want %v", r, want)
	}
}

func TestNormalizeRect_BottomLeftToTopRight(t *testing.T) {
	r := normalizeRect(10, 200, 100, 20)
	want := image.Rect(10, 20, 100, 200)
	if r != want {
		t.Errorf("got %v, want %v", r, want)
	}
}

func TestNormalizeRect_ZeroWidth(t *testing.T) {
	r := normalizeRect(50, 20, 50, 100)
	if r != (image.Rectangle{}) {
		t.Errorf("expected image.Rectangle{} for zero-width, got %v", r)
	}
}

func TestNormalizeRect_ZeroHeight(t *testing.T) {
	r := normalizeRect(20, 50, 100, 50)
	if r != (image.Rectangle{}) {
		t.Errorf("expected image.Rectangle{} for zero-height, got %v", r)
	}
}

func TestNormalizeRect_ZeroArea(t *testing.T) {
	r := normalizeRect(50, 50, 50, 50)
	if r != (image.Rectangle{}) {
		t.Errorf("expected image.Rectangle{} for zero-area, got %v", r)
	}
}

func TestNormalizeRect_SinglePixel(t *testing.T) {
	r := normalizeRect(5, 5, 6, 6)
	want := image.Rect(5, 5, 6, 6)
	if r != want {
		t.Errorf("got %v, want %v", r, want)
	}
}

// ---------------------------------------------------------------------------
// SelectionState
// ---------------------------------------------------------------------------

func TestSelectionState_InitiallyInactive(t *testing.T) {
	var s SelectionState
	if s.IsActive() {
		t.Error("expected inactive initially")
	}
	if s.Bounds() != (image.Rectangle{}) {
		t.Errorf("expected image.Rectangle{} initially, got %v", s.Bounds())
	}
}

func TestSelectionState_BeginActivates(t *testing.T) {
	var s SelectionState
	s.Begin(10, 20)
	if !s.IsActive() {
		t.Error("expected active after Begin")
	}
	// Before any Update the selection has zero area.
	if s.Bounds() != (image.Rectangle{}) {
		t.Errorf("expected image.Rectangle{} before drag, got %v", s.Bounds())
	}
}

func TestSelectionState_UpdateChangeBounds(t *testing.T) {
	var s SelectionState
	s.Begin(10, 20)
	s.Update(100, 200)
	want := image.Rect(10, 20, 100, 200)
	if s.Bounds() != want {
		t.Errorf("got %v, want %v", s.Bounds(), want)
	}
}

func TestSelectionState_EndDeactivates(t *testing.T) {
	var s SelectionState
	s.Begin(10, 20)
	s.Update(100, 200)
	r := s.End()
	if s.IsActive() {
		t.Error("expected inactive after End")
	}
	want := image.Rect(10, 20, 100, 200)
	if r != want {
		t.Errorf("got %v, want %v", r, want)
	}
}

func TestSelectionState_ReverseDrag(t *testing.T) {
	var s SelectionState
	s.Begin(100, 200)
	s.Update(10, 20)
	want := image.Rect(10, 20, 100, 200)
	if s.Bounds() != want {
		t.Errorf("got %v, want %v", s.Bounds(), want)
	}
}

func TestSelectionState_Reset(t *testing.T) {
	var s SelectionState
	s.Begin(10, 20)
	s.Update(100, 200)
	s.Reset()
	if s.IsActive() {
		t.Error("expected inactive after Reset")
	}
	if s.Bounds() != (image.Rectangle{}) {
		t.Errorf("expected image.Rectangle{} after Reset, got %v", s.Bounds())
	}
}

func TestSelectionState_MultipleSelections(t *testing.T) {
	var s SelectionState

	// First selection.
	s.Begin(0, 0)
	s.Update(50, 50)
	r1 := s.End()
	if r1 != image.Rect(0, 0, 50, 50) {
		t.Errorf("first: got %v, want (0,0)-(50,50)", r1)
	}

	// Second selection.
	s.Begin(200, 200)
	s.Update(300, 400)
	r2 := s.End()
	if r2 != image.Rect(200, 200, 300, 400) {
		t.Errorf("second: got %v, want (200,200)-(300,400)", r2)
	}
}

func TestSelectionState_EndWithZeroArea(t *testing.T) {
	var s SelectionState
	s.Begin(50, 50)
	// No Update — same point.
	r := s.End()
	if r != (image.Rectangle{}) {
		t.Errorf("expected image.Rectangle{} for zero-area end, got %v", r)
	}
}
