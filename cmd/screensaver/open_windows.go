//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	openShell32           = windows.NewLazySystemDLL("shell32.dll")
	openUser32            = windows.NewLazySystemDLL("user32.dll")
	openProcShellExecuteW = openShell32.NewProc("ShellExecuteW")
	openProcMessageBoxW   = openUser32.NewProc("MessageBoxW")
)

// openLastScreenshot opens path in the default application via ShellExecuteW.
// If path is empty, a message box is shown instead.
func openLastScreenshot(path string) {
	if path == "" {
		msg, _ := windows.UTF16PtrFromString("No screenshot taken yet.")
		title, _ := windows.UTF16PtrFromString("Open Last Screenshot")
		openProcMessageBoxW.Call(0, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), 0)
		return
	}
	verb, _ := windows.UTF16PtrFromString("open")
	pathPtr, _ := windows.UTF16PtrFromString(path)
	openProcShellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(pathPtr)), 0, 0, 1)
}
