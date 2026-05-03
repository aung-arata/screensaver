//go:build windows

package main

import (
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	siAlreadyUser32    = windows.NewLazySystemDLL("user32.dll")
	siAlreadyMsgBoxW   = siAlreadyUser32.NewProc("MessageBoxW")
)

// showAlreadyRunningMsg displays a Windows MessageBox informing the user that
// the daemon is already running.
func showAlreadyRunningMsg() {
	msg, _ := windows.UTF16PtrFromString("Screensaver is already running. Check the system tray.")
	title, _ := windows.UTF16PtrFromString("Screensaver")
	siAlreadyMsgBoxW.Call(0, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), 0)
	runtime.KeepAlive(msg)
	runtime.KeepAlive(title)
}
