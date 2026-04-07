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
	// Window styles
	trayWsOverlapped = 0x00000000

	// Window messages
	trayWmDestroy  = 0x0002
	trayWmCommand  = 0x0111
	trayWmClose    = 0x0010

	// Custom tray callback message (WM_APP + 1)
	trayWmTrayMsg = 0x8000 + 1

	// Mouse messages used inside tray callback
	trayWmRButtonUp = 0x0205
	trayWmLButtonUp = 0x0202

	// Shell_NotifyIcon dwMessage values
	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	// uFlags
	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	// Stock icon
	idiApplication = 32512

	// TrackPopupMenu flags
	tpmRightAlign     = 0x0008
	tpmBottomAlign    = 0x0020
	tpmRightButton    = 0x0002
	tpmReturncmd      = 0x0100
	tpmNoNotify       = 0x0080

	// AppendMenuW flags
	mfString    = 0x00000000
	mfSeparator = 0x00000800
	mfGrayed    = 0x00000001

	// Menu item IDs
	idTrayCapture = 1001
	idTraySelect  = 1002
	idTrayQuit    = 1003

	// ShowWindow
	traySwHide = 0

	// GetSystemMetrics
	traySmCXScreen = 0
	traySmCYScreen = 1
)

// ---------------------------------------------------------------------------
// Win32 structs
// ---------------------------------------------------------------------------

// notifyIconData mirrors NOTIFYICONDATAW (976 bytes on 64-bit Windows).
// Go's alignment rules produce the same layout as the C struct.
type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr     // 8-byte aligned; Go inserts 4 bytes padding after CbSize
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr     // Go inserts 4 bytes padding to align to 8
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte    // GUID (4-byte align, no padding needed after DwInfoFlags)
	HBalloonIcon     uintptr
}

type trayWinMsg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      trayPoint
}

type trayPoint struct {
	X, Y int32
}

// ---------------------------------------------------------------------------
// DLL procs
// ---------------------------------------------------------------------------

var (
	trayUser32   = windows.NewLazySystemDLL("user32.dll")
	trayKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	trayShell32  = windows.NewLazySystemDLL("shell32.dll")

	trayProcRegisterClassExW = trayUser32.NewProc("RegisterClassExW")
	trayProcCreateWindowExW  = trayUser32.NewProc("CreateWindowExW")
	trayProcDestroyWindow    = trayUser32.NewProc("DestroyWindow")
	trayProcShowWindow       = trayUser32.NewProc("ShowWindow")
	trayProcGetMessageW      = trayUser32.NewProc("GetMessageW")
	trayProcTranslateMessage = trayUser32.NewProc("TranslateMessage")
	trayProcDispatchMessageW = trayUser32.NewProc("DispatchMessageW")
	trayProcDefWindowProcW   = trayUser32.NewProc("DefWindowProcW")
	trayProcPostQuitMessage  = trayUser32.NewProc("PostQuitMessage")
	trayProcLoadIconW        = trayUser32.NewProc("LoadIconW")
	trayProcCreatePopupMenu  = trayUser32.NewProc("CreatePopupMenu")
	trayProcAppendMenuW      = trayUser32.NewProc("AppendMenuW")
	trayProcTrackPopupMenu   = trayUser32.NewProc("TrackPopupMenu")
	trayProcDestroyMenu      = trayUser32.NewProc("DestroyMenu")
	trayProcGetCursorPos     = trayUser32.NewProc("GetCursorPos")
	trayProcSetForegroundWin = trayUser32.NewProc("SetForegroundWindow")
	trayProcGetModuleHandle  = trayKernel32.NewProc("GetModuleHandleW")
	trayProcShellNotifyIcon  = trayShell32.NewProc("Shell_NotifyIconW")
)

// ---------------------------------------------------------------------------
// Tray state
// ---------------------------------------------------------------------------

var trayState struct {
	hwnd uintptr
	cfg  Config
	nid  notifyIconData
}

// Tray window class
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
// Platform implementation
// ---------------------------------------------------------------------------

// runPlatform registers a system tray icon and runs the message loop until
// the user selects Quit or the process is killed.
func runPlatform(cfg Config) error {
	// All Win32 calls must be on the same OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	trayState.cfg = cfg

	hInst, _, _ := trayProcGetModuleHandle.Call(0)

	// Register a hidden window class to receive tray messages.
	className, _ := windows.UTF16PtrFromString("ScreensaverTray")
	wc := trayWndClassExW{
		Size:      uint32(unsafe.Sizeof(trayWndClassExW{})),
		WndProc:   windows.NewCallback(trayWndProc),
		Instance:  hInst,
		ClassName: className,
	}
	atom, _, regErr := trayProcRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 && regErr != windows.ERROR_CLASS_ALREADY_EXISTS {
		return fmt.Errorf("tray: RegisterClassExW: %v", regErr)
	}

	// Create a hidden message-only window.
	hwnd, _, cwErr := trayProcCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		0,
		0,
		0, 0, 0, 0,
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("tray: CreateWindowExW: %v", cwErr)
	}
	trayState.hwnd = hwnd
	trayProcShowWindow.Call(hwnd, traySwHide)

	// Load the application icon.
	hIcon, _, _ := trayProcLoadIconW.Call(0, idiApplication)

	// Build NOTIFYICONDATA and add the tray icon.
	nid := &trayState.nid
	nid.CbSize = uint32(unsafe.Sizeof(notifyIconData{}))
	nid.HWnd = hwnd
	nid.UID = 1
	nid.UFlags = nifMessage | nifIcon | nifTip
	nid.UCallbackMessage = trayWmTrayMsg
	nid.HIcon = hIcon
	tipRunes, _ := windows.UTF16FromString(cfg.Tooltip)
	copy(nid.SzTip[:], tipRunes)

	ok, _, addErr := trayProcShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(nid)))
	if ok == 0 {
		trayProcDestroyWindow.Call(hwnd)
		return fmt.Errorf("tray: Shell_NotifyIconW NIM_ADD: %v", addErr)
	}

	// Message loop.
	var m trayWinMsg
	for {
		ret, _, _ := trayProcGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == -1 {
			break
		}
		if ret == 0 {
			break
		}
		trayProcTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		trayProcDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	// Remove tray icon on exit.
	trayProcShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(nid)))
	return nil
}

// ---------------------------------------------------------------------------
// Window procedure
// ---------------------------------------------------------------------------

func trayWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case trayWmTrayMsg:
		// lParam holds the mouse message type for the tray icon.
		switch lParam {
		case trayWmRButtonUp, trayWmLButtonUp:
			showTrayMenu(hwnd)
		}
		return 0

	case trayWmCommand:
		id := int(wParam & 0xFFFF)
		switch id {
		case idTrayCapture:
			if trayState.cfg.OnCapture != nil {
				go trayState.cfg.OnCapture()
			}
		case idTraySelect:
			if trayState.cfg.OnSelect != nil {
				go trayState.cfg.OnSelect()
			}
		case idTrayQuit:
			// Remove icon first, then trigger quit.
			trayProcShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&trayState.nid)))
			trayProcDestroyWindow.Call(hwnd)
			if trayState.cfg.OnQuit != nil {
				trayState.cfg.OnQuit()
			}
		}
		return 0

	case trayWmDestroy:
		trayProcPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := trayProcDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// showTrayMenu creates and displays the right-click context menu.
func showTrayMenu(hwnd uintptr) {
	hMenu, _, _ := trayProcCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer trayProcDestroyMenu.Call(hMenu)

	captureText, _ := windows.UTF16PtrFromString("Take Screenshot")
	selectText, _ := windows.UTF16PtrFromString("Select Region")
	quitText, _ := windows.UTF16PtrFromString("Quit")

	trayProcAppendMenuW.Call(hMenu, mfString, idTrayCapture, uintptr(unsafe.Pointer(captureText)))
	trayProcAppendMenuW.Call(hMenu, mfString, idTraySelect, uintptr(unsafe.Pointer(selectText)))
	trayProcAppendMenuW.Call(hMenu, mfSeparator, 0, 0)
	trayProcAppendMenuW.Call(hMenu, mfString, idTrayQuit, uintptr(unsafe.Pointer(quitText)))

	// Get cursor position for menu placement.
	var pt trayPoint
	trayProcGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	// Required before TrackPopupMenu so the menu closes when clicking elsewhere.
	trayProcSetForegroundWin.Call(hwnd)

	// Display the menu; TPM_RETURNCMD returns the chosen item directly.
	cmd, _, _ := trayProcTrackPopupMenu.Call(
		hMenu,
		tpmRightButton|tpmReturncmd|tpmNoNotify|tpmRightAlign|tpmBottomAlign,
		uintptr(pt.X), uintptr(pt.Y),
		0, hwnd, 0,
	)
	if cmd != 0 {
		// Re-use the WM_COMMAND handler.
		trayWndProc(hwnd, trayWmCommand, cmd, 0)
	}
}
