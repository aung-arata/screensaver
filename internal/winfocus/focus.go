// Package winfocus provides helpers to locate and focus top-level windows on
// Windows. It is used to bring a chosen window to the foreground before
// sending synthetic scroll input, so wheel events reach the intended
// application.
package winfocus

// FocusAt resolves the top-level window under the given screen point and
// brings it to the foreground.
//
// On non-Windows platforms this always returns an error.
func FocusAt(x, y int) error {
	return focusAt(x, y)
}
