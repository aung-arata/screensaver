//go:build windows

package main

import (
	"runtime"
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
		runtime.KeepAlive(msg)
		runtime.KeepAlive(title)
		return
	}
	verb, _ := windows.UTF16PtrFromString("open")
	pathPtr, _ := windows.UTF16PtrFromString(path)
	ret, _, _ := openProcShellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(pathPtr)), 0, 0, 1)
	// Keep the Go allocations alive until after the syscall returns.
	runtime.KeepAlive(verb)
	runtime.KeepAlive(pathPtr)
	// ShellExecuteW returns an HINSTANCE; values <= 32 indicate failure.
	if ret <= 32 {
		errMsg, _ := windows.UTF16PtrFromString("Failed to open screenshot:\n" + path)
		title, _ := windows.UTF16PtrFromString("Open Last Screenshot")
		openProcMessageBoxW.Call(0, uintptr(unsafe.Pointer(errMsg)), uintptr(unsafe.Pointer(title)), 0)
		runtime.KeepAlive(errMsg)
		runtime.KeepAlive(title)
	}
}
