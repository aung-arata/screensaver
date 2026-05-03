//go:build !windows

// Package console provides helpers for detaching from the console window.
package console

// Detach is a no-op on non-Windows platforms.
func Detach() {}
