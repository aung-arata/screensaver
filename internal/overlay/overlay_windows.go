//go:build windows

package overlay

import (
	"fmt"
	"image"
)

// showPlatform implements the overlay for Windows using Win32 layered windows.
//
// A full implementation would:
//   1. Create a transparent, topmost, full-screen layered window.
//   2. Capture the screen and draw a darkened version as the background.
//   3. Track mouse drag events to draw a bright selection rectangle.
//   4. On mouse-up, return the selected region.
//   5. On Escape, cancel and return Cancelled=true.
//
// This scaffold returns an error directing users to --once mode until
// the Win32 implementation is complete.
func showPlatform(monitor int) (*Result, error) {
	_ = image.Rect(0, 0, 0, 0) // suppress unused import
	return nil, fmt.Errorf("overlay: Win32 overlay implementation in progress; use --once for headless capture")
}
