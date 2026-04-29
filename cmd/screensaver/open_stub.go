//go:build !windows

package main

// openLastScreenshot is a no-op on non-Windows platforms.
func openLastScreenshot(_ string) {}
