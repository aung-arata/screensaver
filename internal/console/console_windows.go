//go:build windows

// Package console provides helpers for detaching from the console window.
package console

import "golang.org/x/sys/windows"

var (
	kernel32        = windows.NewLazySystemDLL("kernel32.dll")
	procFreeConsole = kernel32.NewProc("FreeConsole")
)

// Detach frees the calling process from its console window.
// Call this when entering background daemon mode to prevent a black console
// window from appearing on the taskbar.
func Detach() {
	procFreeConsole.Call() //nolint:errcheck
}
