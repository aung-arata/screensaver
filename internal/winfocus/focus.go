// Package winfocus provides helpers to enumerate and focus top-level windows
// on Windows. It is used to let the user pick a target window and then bring
// it to the foreground before sending synthetic scroll input, so wheel events
// reach the intended application.
package winfocus

// Window describes a top-level window that can be focused.
type Window struct {
	// Handle is the top-level window handle (HWND).
	Handle uintptr
	// Title is the window's caption text.
	Title string
}

// EnumerateWindows returns the visible top-level windows that are suitable
// candidates for focus, filtering out invisible, tool, empty-titled, and
// self-owned windows.
//
// On non-Windows platforms this always returns an error.
func EnumerateWindows() ([]Window, error) {
	return enumerateWindows()
}

// Focus brings the given top-level window to the foreground.
//
// On non-Windows platforms this always returns an error.
func Focus(hwnd uintptr) error {
	return focus(hwnd)
}
