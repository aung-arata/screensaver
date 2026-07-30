package overlay

import "image"

// SelectionState tracks the state of a rubber-band region selection.
//
// Usage:
//  1. Call Begin when the user presses the mouse button.
//  2. Call Update as the user drags.
//  3. Call End when the user releases the mouse button.
//  4. Use Bounds at any time to get the current rectangle.
type SelectionState struct {
	active   bool
	startX   int
	startY   int
	currentX int
	currentY int
}

// Begin starts a new selection at the given screen coordinates.
func (s *SelectionState) Begin(x, y int) {
	s.active = true
	s.startX = x
	s.startY = y
	s.currentX = x
	s.currentY = y
}

// Update moves the current endpoint of the selection to (x, y).
func (s *SelectionState) Update(x, y int) {
	s.currentX = x
	s.currentY = y
}

// End finalizes the selection and returns the selected rectangle.
// Returns image.Rectangle{} if the selection has zero area.
func (s *SelectionState) End() image.Rectangle {
	s.active = false
	return s.Bounds()
}

// Bounds returns the current selection rectangle normalized so that
// Min is the top-left corner and Max is the bottom-right corner.
// Returns image.Rectangle{} if the selection has zero area.
func (s *SelectionState) Bounds() image.Rectangle {
	return normalizeRect(s.startX, s.startY, s.currentX, s.currentY)
}

// IsActive reports whether a selection drag is in progress.
func (s *SelectionState) IsActive() bool {
	return s.active
}

// Reset clears the selection state.
func (s *SelectionState) Reset() {
	s.active = false
	s.startX = 0
	s.startY = 0
	s.currentX = 0
	s.currentY = 0
}

// normalizeRect converts two arbitrary corner points into a canonical
// image.Rectangle with Min <= Max. It returns image.Rectangle{} if the
// resulting rectangle has zero width or height.
func normalizeRect(x1, y1, x2, y2 int) image.Rectangle {
	r := image.Rect(x1, y1, x2, y2).Canon()
	if r.Dx() == 0 || r.Dy() == 0 {
		return image.Rectangle{}
	}
	return r
}
