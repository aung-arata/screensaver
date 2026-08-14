// Package overlay provides a fullscreen selection overlay.
//
// The overlay darkens the screen and lets the user drag a rubber-band
// rectangle to select a capture region. On Windows this is implemented
// as a transparent layered window via Win32 APIs; on other platforms
// a GUI toolkit like Fyne would be used.
//
// This package defines the public interface and types. Platform-specific
// implementations are in separate files with build tags.
package overlay

import "image"

// Result holds the outcome of a selection overlay interaction.
type Result struct {
	// Region is the selected rectangle, or image.ZR if cancelled.
	Region image.Rectangle
	// Cancelled is true if the user pressed Escape.
	Cancelled bool
}

// Show displays the fullscreen selection overlay on the given monitor
// (0-based index) and blocks until the user completes or cancels the
// selection.
//
// Returns a Result describing the selected region.
func Show(monitor int) (*Result, error) {
	return showPlatform(monitor)
}

// PickWindow dims the screen and waits for the user to click a window,
// bringing that window to the foreground. The user can cancel with Escape.
//
// The returned Result has Cancelled set when the user pressed Escape; its
// Region is always image.ZR. On non-Windows platforms an error is returned.
func PickWindow(monitor int) (*Result, error) {
	return pickPlatform(monitor)
}
