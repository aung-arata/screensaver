//go:build windows

package tray

import (
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/aung-arata/screensaver/internal/appicon"
	"github.com/aung-arata/screensaver/internal/history"
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
	mfPopup     = 0x00000010

	// Menu item IDs
	idTrayCapture  = 1001
	idTraySelect   = 1002
	idTrayOpenLast = 1003
	idTrayQuit     = 1004

	// Recent screenshot submenu IDs (5 slots)
	idTrayRecent0 = 2001
	idTrayRecent1 = 2002
	idTrayRecent2 = 2003
	idTrayRecent3 = 2004
	idTrayRecent4 = 2005

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

// Compile-time assertion: notifyIconData must be 976 bytes on 64-bit Windows.
var _ [976]byte = [unsafe.Sizeof(notifyIconData{})]byte{}

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
	trayProcLoadIconW              = trayUser32.NewProc("LoadIconW")
	trayProcCreateIconFromResourceEx = trayUser32.NewProc("CreateIconFromResourceEx")
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

// trayRecentPaths stores the file paths of the up to 5 recent screenshots
// shown in the "Recent Screenshots" submenu. Populated in showTrayMenu and
// consumed in trayWndProc.
var trayRecentPaths [5]string

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

	// Load the application icon from embedded ICO data.
	// CreateIconFromResourceEx interprets the raw ICO bytes and returns an HICON.
	hIcon, _, _ := trayProcCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&appicon.Data[0])),
		uintptr(len(appicon.Data)),
		1,    // fIcon = TRUE (icon, not cursor)
		0x00030000, // dwVer = 0x00030000 (Windows 3.x+ icon format)
		0, 0, // cxDesired, cyDesired = 0 means use actual size
		0,    // flags = 0
	)
	if hIcon == 0 {
		// Fall back to the default application icon if custom icon fails to load.
		hIcon, _, _ = trayProcLoadIconW.Call(0, idiApplication)
	}

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
		case idTrayOpenLast:
			if trayState.cfg.OnOpenLast != nil {
				go trayState.cfg.OnOpenLast()
			}
		case idTrayQuit:
			// DestroyWindow triggers WM_DESTROY → PostQuitMessage → message
			// loop exits → icon is deleted once after the loop in runPlatform.
			trayProcDestroyWindow.Call(hwnd)
			if trayState.cfg.OnQuit != nil {
				trayState.cfg.OnQuit()
			}
		case idTrayRecent0, idTrayRecent1, idTrayRecent2, idTrayRecent3, idTrayRecent4:
			idx := id - idTrayRecent0
			if idx >= 0 && idx < 5 && trayRecentPaths[idx] != "" && trayState.cfg.OnRecent != nil {
				go trayState.cfg.OnRecent(trayRecentPaths[idx])
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
	openLastText, _ := windows.UTF16PtrFromString("Open Last Screenshot")
	recentText, _ := windows.UTF16PtrFromString("Recent Screenshots \u25ba")
	quitText, _ := windows.UTF16PtrFromString("Quit")

	trayProcAppendMenuW.Call(hMenu, mfString, idTrayCapture, uintptr(unsafe.Pointer(captureText)))
	trayProcAppendMenuW.Call(hMenu, mfString, idTraySelect, uintptr(unsafe.Pointer(selectText)))
	trayProcAppendMenuW.Call(hMenu, mfString, idTrayOpenLast, uintptr(unsafe.Pointer(openLastText)))

	// Build the "Recent Screenshots" submenu from the last 5 history entries.
	hSubMenu, _, _ := trayProcCreatePopupMenu.Call()
	if hSubMenu != 0 {
		// Clear the paths array so stale entries are not re-used.
		trayRecentPaths = [5]string{}

		entries, err := history.Recent(5)
		if err != nil || len(entries) == 0 {
			noItemText, _ := windows.UTF16PtrFromString("No screenshots yet")
			trayProcAppendMenuW.Call(hSubMenu, mfString|mfGrayed, 0, uintptr(unsafe.Pointer(noItemText)))
		} else {
			for i, e := range entries {
				trayRecentPaths[i] = e.Path
				label, _ := windows.UTF16PtrFromString(filepath.Base(e.Path))
				trayProcAppendMenuW.Call(hSubMenu, mfString, uintptr(idTrayRecent0+i), uintptr(unsafe.Pointer(label)))
			}
		}
		// Attach submenu to main menu. When MF_POPUP is set, uIDNewItem is the HMENU.
		trayProcAppendMenuW.Call(hMenu, mfPopup, hSubMenu, uintptr(unsafe.Pointer(recentText)))
	}

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
