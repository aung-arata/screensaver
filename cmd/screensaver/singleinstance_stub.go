//go:build !windows

package main

// showAlreadyRunningMsg is a no-op on non-Windows platforms.
func showAlreadyRunningMsg() {}
