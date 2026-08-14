//go:build windows

package winfocus

import (
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const gaRoot = 2

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procWindowFromPoint     = user32.NewProc("WindowFromPoint")
	procGetAncestor         = user32.NewProc("GetAncestor")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procAttachThreadInput   = user32.NewProc("AttachThreadInput")
)

type point struct {
	X int32
	Y int32
}

func focusAt(x, y int) error {
	hwnd, err := windowAt(x, y)
	if err != nil {
		return err
	}
	return focusWindow(hwnd)
}

// windowAt resolves the top-level window under the given screen point,
// resolving any child window to its top-level ancestor.
func windowAt(x, y int) (uintptr, error) {
	p := point{X: int32(x), Y: int32(y)}
	hwnd, _, _ := procWindowFromPoint.Call(uintptr(unsafe.Pointer(&p)))
	if hwnd == 0 {
		return 0, fmt.Errorf("WindowFromPoint: no window at (%d, %d)", x, y)
	}
	if root, _, _ := procGetAncestor.Call(hwnd, gaRoot); root != 0 {
		hwnd = root
	}
	return hwnd, nil
}

// focusWindow brings the given top-level window to the foreground.
//
// It attaches the calling thread to the foreground thread so that the
// SetForegroundWindow call is permitted (see MSDN: a process may set the
// foreground window only if it is the foreground process or received the last
// input event), then detaches again. It blocks briefly afterwards to let the
// foreground activation settle.
func focusWindow(hwnd uintptr) error {
	currentThread := uintptr(windows.GetCurrentThreadId())
	if currentThread == 0 {
		return fmt.Errorf("GetCurrentThreadId failed")
	}

	// Resolve the foreground thread, if there is a foreground window.
	var fgThread uintptr
	if fgHwnd := windows.GetForegroundWindow(); fgHwnd != 0 {
		tid, err := windows.GetWindowThreadProcessId(fgHwnd, nil)
		if err != nil {
			return fmt.Errorf("GetWindowThreadProcessId: failed to resolve foreground window thread: %w", err)
		}
		fgThread = uintptr(tid)
	}

	attached := false
	if fgThread != 0 && fgThread != currentThread {
		if ret, _, _ := procAttachThreadInput.Call(currentThread, fgThread, 1); ret != 0 {
			attached = true
		}
	}

	ret, _, _ := procSetForegroundWindow.Call(hwnd)

	if attached {
		procAttachThreadInput.Call(currentThread, fgThread, 0)
	}

	if ret == 0 {
		return fmt.Errorf("SetForegroundWindow: failed to focus target window")
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}
