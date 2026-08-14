//go:build windows

package winfocus

import (
	"fmt"
	"os"
	"sort"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	gwlExStyle     = ^uintptr(19) // -20 (GWL_EXSTYLE)
	wsExToolWindow = 0x00000080
	wsExAppWindow  = 0x00040000

	swRestore = 9
)

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procIsIconic                 = user32.NewProc("IsIconic")
	procShowWindow               = user32.NewProc("ShowWindow")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput        = user32.NewProc("AttachThreadInput")
)

func enumerateWindows() ([]Window, error) {
	var wins []Window
	cb := windows.NewCallback(func(hwnd, lParam uintptr) uintptr {
		if w, ok := usableWindow(hwnd); ok {
			wins = append(wins, w)
		}
		return 1 // continue enumeration
	})
	if err := windows.EnumWindows(cb, nil); err != nil {
		return nil, fmt.Errorf("EnumWindows: %w", err)
	}

	sort.SliceStable(wins, func(i, j int) bool {
		return wins[i].Title < wins[j].Title
	})
	return wins, nil
}

// usableWindow reports whether hwnd is a candidate for focus, and returns the
// corresponding Window if so. It skips invisible windows, tool windows (unless
// they declare an app window style), windows without a title, the desktop and
// shell windows, and windows owned by our own process.
func usableWindow(hwnd uintptr) (Window, bool) {
	if hwnd == 0 {
		return Window{}, false
	}
	if !windows.IsWindowVisible(windows.HWND(hwnd)) {
		return Window{}, false
	}

	// Skip our own process's windows (e.g. the selection overlay).
	var pid uint32
	if _, err := windows.GetWindowThreadProcessId(windows.HWND(hwnd), &pid); err != nil {
		return Window{}, false
	}
	if int(pid) == os.Getpid() {
		return Window{}, false
	}

	// Skip the desktop and shell windows.
	if desktop := windows.GetDesktopWindow(); hwnd == uintptr(desktop) {
		return Window{}, false
	}
	if shell := windows.GetShellWindow(); shell != 0 && hwnd == uintptr(shell) {
		return Window{}, false
	}

	title, ok := windowTitle(hwnd)
	if !ok || title == "" {
		return Window{}, false
	}

	exStyle, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlExStyle)
	if exStyle&wsExToolWindow != 0 && exStyle&wsExAppWindow == 0 {
		return Window{}, false
	}

	return Window{Handle: hwnd, Title: title}, true
}

func windowTitle(hwnd uintptr) (string, bool) {
	buf := make([]uint16, 512)
	n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return "", false
	}
	return windows.UTF16ToString(buf[:n]), true
}

// focus brings the given top-level window to the foreground.
//
// If the window is minimized it is restored first. It then attaches the
// calling thread to the current foreground window's thread so that
// SetForegroundWindow is permitted (see MSDN: a process may set the
// foreground window only if it is the foreground process or received the last
// input event), and detaches afterwards. It blocks briefly to let the
// foreground activation settle.
func focus(hwnd uintptr) error {
	if hwnd == 0 {
		return fmt.Errorf("winfocus: invalid window handle")
	}

	// Restore a minimized window before foregrounding it.
	if icon, _, _ := procIsIconic.Call(hwnd); icon != 0 {
		procShowWindow.Call(hwnd, swRestore)
	}

	currentThread := uintptr(windows.GetCurrentThreadId())
	if currentThread == 0 {
		return fmt.Errorf("GetCurrentThreadId failed")
	}

	var foregroundThread uintptr
	if foregroundWindow, _, _ := procGetForegroundWindow.Call(); foregroundWindow != 0 {
		foregroundThread, _, _ = procGetWindowThreadProcessId.Call(foregroundWindow, 0)
		if foregroundThread == 0 {
			return fmt.Errorf("GetWindowThreadProcessId: failed to resolve foreground window thread")
		}
	}

	attached := false
	if foregroundThread != 0 && foregroundThread != currentThread {
		if ret, _, _ := procAttachThreadInput.Call(currentThread, foregroundThread, 1); ret != 0 {
			attached = true
		}
	}

	ret, _, _ := procSetForegroundWindow.Call(hwnd)

	if attached {
		procAttachThreadInput.Call(currentThread, foregroundThread, 0)
	}

	if ret == 0 {
		return fmt.Errorf("SetForegroundWindow: failed to focus target window")
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}
