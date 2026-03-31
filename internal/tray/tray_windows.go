//go:build windows

package tray

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ---------------------------------------------------------------------------
// Win32 constants
// ---------------------------------------------------------------------------

const (
	nimAdd    = 0x00000000
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	wmApp     = 0x8000
	wmTrayMsg = wmApp + 1

	wmCommand   = 0x0111
	wmDestroy   = 0x0002
	wmRButtonUp = 0x0205

	tpmBottomAlign = 0x0020
	tpmLeftAlign   = 0x0000

	mfString    = 0x0000
	mfSeparator = 0x0800

	idCapture = 1001
	idSelect  = 1002
	idQuit    = 1003

	idiApplication = 32512
)

// ---------------------------------------------------------------------------
// Win32 structures
// ---------------------------------------------------------------------------

type notifyIconDataW struct {
	Size            uint32
	HWnd            uintptr
	UID             uint32
	Flags           uint32
	CallbackMessage uint32
	HIcon           uintptr
	Tip             [128]uint16
}

type trayMsg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

type trayPoint struct {
	X, Y int32
}

type trayWndClassExW struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

// ---------------------------------------------------------------------------
// Win32 syscall procs
// ---------------------------------------------------------------------------

var (
	trayShell32  = windows.NewLazySystemDLL("shell32.dll")
	trayUser32   = windows.NewLazySystemDLL("user32.dll")
	trayKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procShellNotifyIconW       = trayShell32.NewProc("Shell_NotifyIconW")
	trayProcRegisterClassExW   = trayUser32.NewProc("RegisterClassExW")
	trayProcCreateWindowExW    = trayUser32.NewProc("CreateWindowExW")
	trayProcDestroyWindow      = trayUser32.NewProc("DestroyWindow")
	trayProcDefWindowProcW     = trayUser32.NewProc("DefWindowProcW")
	trayProcGetMessageW        = trayUser32.NewProc("GetMessageW")
	trayProcTranslateMessage   = trayUser32.NewProc("TranslateMessage")
	trayProcDispatchMessageW   = trayUser32.NewProc("DispatchMessageW")
	trayProcPostQuitMessage    = trayUser32.NewProc("PostQuitMessage")
	trayProcLoadIconW          = trayUser32.NewProc("LoadIconW")
	trayProcCreatePopupMenu    = trayUser32.NewProc("CreatePopupMenu")
	trayProcAppendMenuW        = trayUser32.NewProc("AppendMenuW")
	trayProcTrackPopupMenu     = trayUser32.NewProc("TrackPopupMenu")
	trayProcDestroyMenu        = trayUser32.NewProc("DestroyMenu")
	trayProcSetForegroundWindow = trayUser32.NewProc("SetForegroundWindow")
	trayProcGetCursorPos       = trayUser32.NewProc("GetCursorPos")
	trayProcGetModuleHandleW   = trayKernel32.NewProc("GetModuleHandleW")
)

// ---------------------------------------------------------------------------
// Tray state (shared with the window proc via package-level variables
// because Win32 WndProc callbacks cannot carry user data directly).
// ---------------------------------------------------------------------------

var trayState struct {
	hwnd uintptr
	nid  notifyIconDataW
	cbs  Callbacks
}

// trayWndProc handles Win32 messages for the hidden tray window.
func trayWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmTrayMsg:
		switch lParam {
		case wmRButtonUp:
			showTrayMenu(hwnd)
		}
		return 0

	case wmCommand:
		id := int(wParam & 0xFFFF)
		switch id {
		case idCapture:
			if trayState.cbs.OnCapture != nil {
				go trayState.cbs.OnCapture()
			}
		case idSelect:
			if trayState.cbs.OnSelectRegion != nil {
				go trayState.cbs.OnSelectRegion()
			}
		case idQuit:
			removeTrayIcon()
			trayProcDestroyWindow.Call(hwnd)
		}
		return 0

	case wmDestroy:
		if trayState.cbs.OnQuit != nil {
			trayState.cbs.OnQuit()
		}
		trayProcPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := trayProcDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// showTrayMenu creates and displays a context menu at the cursor position.
func showTrayMenu(hwnd uintptr) {
	menu, _, _ := trayProcCreatePopupMenu.Call()
	if menu == 0 {
		return
	}

	captureStr, _ := windows.UTF16PtrFromString("Capture Screen")
	selectStr, _ := windows.UTF16PtrFromString("Select Region")
	quitStr, _ := windows.UTF16PtrFromString("Quit")

	trayProcAppendMenuW.Call(menu, mfString, idCapture, uintptr(unsafe.Pointer(captureStr)))
	trayProcAppendMenuW.Call(menu, mfString, idSelect, uintptr(unsafe.Pointer(selectStr)))
	trayProcAppendMenuW.Call(menu, mfSeparator, 0, 0)
	trayProcAppendMenuW.Call(menu, mfString, idQuit, uintptr(unsafe.Pointer(quitStr)))

	// Required so the menu dismisses when clicking outside it.
	trayProcSetForegroundWindow.Call(hwnd)

	var pt trayPoint
	trayProcGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	trayProcTrackPopupMenu.Call(menu, tpmBottomAlign|tpmLeftAlign,
		uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	trayProcDestroyMenu.Call(menu)
}

// addTrayIcon registers a system tray icon with a tooltip.
func addTrayIcon(hwnd uintptr, tooltip string) {
	icon, _, _ := trayProcLoadIconW.Call(0, idiApplication)

	nid := &trayState.nid
	nid.Size = uint32(unsafe.Sizeof(notifyIconDataW{}))
	nid.HWnd = hwnd
	nid.UID = 1
	nid.Flags = nifMessage | nifIcon | nifTip
	nid.CallbackMessage = wmTrayMsg
	nid.HIcon = icon

	tip, _ := windows.UTF16FromString(tooltip)
	copy(nid.Tip[:], tip)

	procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(nid)))
}

// removeTrayIcon removes the system tray icon.
func removeTrayIcon() {
	nid := &trayState.nid
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(nid)))
}

// ---------------------------------------------------------------------------
// run creates a hidden message-only window, adds a system tray icon,
// and enters a Win32 message loop.  It blocks until the user selects
// "Quit" from the tray menu or the window is destroyed.
// ---------------------------------------------------------------------------

func run(cfg Config, cbs Callbacks) error {
	// Pin to OS thread: Win32 window messages are thread-local.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	trayState.cbs = cbs

	hInst, _, _ := trayProcGetModuleHandleW.Call(0)

	// Register a window class for the hidden tray message window.
	className, _ := windows.UTF16PtrFromString("ScreensaverTray")
	wc := trayWndClassExW{
		Size:      uint32(unsafe.Sizeof(trayWndClassExW{})),
		WndProc:   windows.NewCallback(trayWndProc),
		Instance:  hInst,
		ClassName: className,
	}
	atom, _, regErr := trayProcRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 && regErr != windows.ERROR_CLASS_ALREADY_EXISTS {
		return fmt.Errorf("tray: RegisterClassExW failed: %v", regErr)
	}

	// Create a message-only window (HWND_MESSAGE parent).
	windowName, _ := windows.UTF16PtrFromString("Screensaver Tray")
	hwndMessage := ^uintptr(2) // HWND_MESSAGE = (HWND)-3
	hwnd, _, cwErr := trayProcCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		0, // no style — hidden window
		0, 0, 0, 0,
		hwndMessage, 0, hInst, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("tray: CreateWindowExW failed: %v", cwErr)
	}
	trayState.hwnd = hwnd

	addTrayIcon(hwnd, cfg.Tooltip)

	// Message loop.
	var m trayMsg
	for {
		ret, _, _ := trayProcGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		trayProcTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		trayProcDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	return nil
}
