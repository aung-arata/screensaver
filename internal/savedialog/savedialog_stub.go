//go:build !windows

// Package savedialog provides a native Windows Save File dialog.
package savedialog

import "errors"

// ShowSaveDialog is not supported on non-Windows platforms.
// It always returns ("", error).
func ShowSaveDialog(_ uintptr, _ string) (string, error) {
	return "", errors.New("save dialog not supported on this platform")
}
