//go:build !windows

package overlay

import "fmt"

// showPlatform is the non-Windows stub for the selection overlay.
//
// The interactive overlay currently requires Windows (Win32 APIs).
// Linux and macOS support will be added via a GUI toolkit (e.g. Fyne)
// in a future release.
func showPlatform(monitor int) (*Result, error) {
	return nil, fmt.Errorf("overlay: interactive selection overlay is not yet implemented for this platform; use --once for headless capture")
}
