//go:build windows

// Package singleinstance provides a named Windows mutex to prevent multiple
// daemon instances from running simultaneously.
package singleinstance

import (
	"golang.org/x/sys/windows"
)

const mutexName = "Global\\ScreensaverDaemonMutex"

// mutexHandle holds the open mutex handle for the lifetime of the process.
// It must not be garbage-collected or closed while the process is running.
var mutexHandle windows.Handle

// Acquire tries to create a named global mutex.
// Returns true if this is the first instance, or false if another instance is
// already running.
// On success the mutex handle is kept alive as a package-level variable for
// the lifetime of the process; the caller does not need to manage it.
func Acquire() bool {
	name, _ := windows.UTF16PtrFromString(mutexName)
	h, err := windows.CreateMutex(nil, false, name)
	if err == windows.ERROR_ALREADY_EXISTS {
		return false
	}
	mutexHandle = h
	return true
}
