//go:build windows

package overlay

import (
	"fmt"
	"image"
	"runtime"
	"unsafe"

	"github.com/aung-arata/screensaver/internal/capture"
	"golang.org/x/sys/windows"
)

// ---------------------------------------------------------------------------
// Win32 constants
// ---------------------------------------------------------------------------

const (
	csHRedraw = 0x0002
	csVRedraw = 0x0001

	cwUseDefault = ^0x7fffffff

	wsPopup   = 0x80000000
	wsVisible = 0x10000000

	wsExTopmost    = 0x00000008
	wsExToolWindow = 0x00000080

	wmDestroy     = 0x0002
	wmPaint       = 0x000F
	wmKeyDown     = 0x0100
	wmLButtonDown = 0x0201
	wmLButtonUp   = 0x0202
	wmMouseMove   = 0x0200
	wmSetCursor   = 0x0020
	wmEraseBkgnd  = 0x0014

	vkEscape = 0x1B

	swShow = 5

	dibRGBColors = 0
	srcCopy      = 0x00CC0020

	biRGB = 0

	psSolid = 0

	nullBrush = 5

	idcCross = 32515
	htClient = 1
)

// ---------------------------------------------------------------------------
// Win32 structures
// ---------------------------------------------------------------------------

type wndClassExW struct {
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

type winMsg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type point struct {
	X, Y int32
}

type paintStruct struct {
	HDC         uintptr
	Erase       int32
	RcPaint     rect
	Restore     int32
	IncUpdate   int32
	RGBReserved [32]byte
}

type rect struct {
	Left, Top, Right, Bottom int32
}

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}

// ---------------------------------------------------------------------------
// Win32 syscall procs
// ---------------------------------------------------------------------------

var (
	overlayUser32   = windows.NewLazySystemDLL("user32.dll")
	overlayGdi32    = windows.NewLazySystemDLL("gdi32.dll")
	overlayKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassExW = overlayUser32.NewProc("RegisterClassExW")
	procCreateWindowExW  = overlayUser32.NewProc("CreateWindowExW")
	procDestroyWindow    = overlayUser32.NewProc("DestroyWindow")
	procShowWindow       = overlayUser32.NewProc("ShowWindow")
	procUpdateWindow     = overlayUser32.NewProc("UpdateWindow")
	procGetMessageW      = overlayUser32.NewProc("GetMessageW")
	procTranslateMessage = overlayUser32.NewProc("TranslateMessage")
	procDispatchMessageW = overlayUser32.NewProc("DispatchMessageW")
	procDefWindowProcW   = overlayUser32.NewProc("DefWindowProcW")
	procPostQuitMessage  = overlayUser32.NewProc("PostQuitMessage")
	procBeginPaint       = overlayUser32.NewProc("BeginPaint")
	procEndPaint         = overlayUser32.NewProc("EndPaint")
	procInvalidateRect   = overlayUser32.NewProc("InvalidateRect")
	procLoadCursorW      = overlayUser32.NewProc("LoadCursorW")
	procSetCursor        = overlayUser32.NewProc("SetCursor")
	procSetCapture       = overlayUser32.NewProc("SetCapture")
	procReleaseCapture   = overlayUser32.NewProc("ReleaseCapture")
	procGetModuleHandleW = overlayKernel32.NewProc("GetModuleHandleW")

	procCreateCompatibleDC = overlayGdi32.NewProc("CreateCompatibleDC")
	procDeleteDC           = overlayGdi32.NewProc("DeleteDC")
	procDeleteObject       = overlayGdi32.NewProc("DeleteObject")
	procSelectObject       = overlayGdi32.NewProc("SelectObject")
	procCreatePen          = overlayGdi32.NewProc("CreatePen")
	procGetStockObject     = overlayGdi32.NewProc("GetStockObject")
	procRectangle          = overlayGdi32.NewProc("Rectangle")
	procStretchDIBits      = overlayGdi32.NewProc("StretchDIBits")
	procCreateDIBSection   = overlayGdi32.NewProc("CreateDIBSection")
	procBitBlt             = overlayGdi32.NewProc("BitBlt")
)

// ---------------------------------------------------------------------------
// Overlay state (shared with the window proc via package-level variables
// because Win32 WndProc callbacks cannot carry user data directly).
// ---------------------------------------------------------------------------

var (
	ovState struct {
		screenshot    *image.RGBA // original (un-dimmed) capture
		dimmed        *image.RGBA // dimmed background
		sel           SelectionState
		result        *Result
		hwnd          uintptr
		monitorBounds image.Rectangle
		crossCursor   uintptr
	}
)

// loWord extracts the low-order 16-bit signed value from an lParam, typically representing the X coordinate of a mouse message.
// It returns that value as an int.
func loWord(l uintptr) int { return int(int16(l & 0xFFFF)) }
// hiWord extracts the high-order 16 bits of l and returns them as a signed int (sign-extended from 16 bits).
func hiWord(l uintptr) int { return int(int16((l >> 16) & 0xFFFF)) }

// ---------------------------------------------------------------------------
// wndProc – handles paint, mouse, and keyboard events.
// wndProc handles Win32 messages for the overlay window, processing cursor updates,
// painting, mouse input to drive the selection lifecycle, keyboard cancellation, and
// window teardown.
// For messages it does not handle, it delegates to DefWindowProcW and returns that result.

func wndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case wmSetCursor:
		if loWord(lParam) == htClient {
			procSetCursor.Call(ovState.crossCursor)
			return 1
		}

	case wmPaint:
		var ps paintStruct
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		paintOverlay(hdc)
		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case wmEraseBkgnd:
		return 1 // prevent flickering

	case wmLButtonDown:
		x, y := loWord(lParam), hiWord(lParam)
		ovState.sel.Begin(x, y)
		procSetCapture.Call(hwnd)
		return 0

	case wmMouseMove:
		if ovState.sel.IsActive() {
			x, y := loWord(lParam), hiWord(lParam)
			ovState.sel.Update(x, y)
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		return 0

	case wmLButtonUp:
		if ovState.sel.IsActive() {
			x, y := loWord(lParam), hiWord(lParam)
			ovState.sel.Update(x, y)
			region := ovState.sel.End()
			procReleaseCapture.Call()
			ovState.result = &Result{Region: region, Cancelled: region == image.ZR}
			procDestroyWindow.Call(hwnd)
		}
		return 0

	case wmKeyDown:
		if wParam == vkEscape {
			ovState.sel.Reset()
			ovState.result = &Result{Cancelled: true}
			procDestroyWindow.Call(hwnd)
		}
		return 0

	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// paintOverlay renders the dimmed screenshot, then composites the bright
// hdc is a Win32 device context handle where the overlay will be drawn.
func paintOverlay(hdc uintptr) {
	if ovState.dimmed == nil {
		return
	}

	// Decide which image to render: compose selection into dimmed if active.
	frame := image.NewRGBA(ovState.dimmed.Bounds())
	copy(frame.Pix, ovState.dimmed.Pix)

	sel := ovState.sel.Bounds()
	if sel != image.ZR {
		ComposeSelection(frame, ovState.screenshot, sel)
	}

	bounds := frame.Bounds()
	w, h := int32(bounds.Dx()), int32(bounds.Dy())

	// Convert RGBA (top-down) to BGRA (bottom-up) for GDI StretchDIBits.
	pixels := make([]byte, len(frame.Pix))
	stride := int(frame.Stride)
	for y := 0; y < int(h); y++ {
		srcRow := frame.Pix[y*stride : y*stride+stride]
		dstRow := pixels[(int(h)-1-y)*stride : (int(h)-1-y)*stride+stride]
		for x := 0; x < stride; x += 4 {
			dstRow[x+0] = srcRow[x+2] // B
			dstRow[x+1] = srcRow[x+1] // G
			dstRow[x+2] = srcRow[x+0] // R
			dstRow[x+3] = srcRow[x+3] // A
		}
	}

	bi := bitmapInfo{
		Header: bitmapInfoHeader{
			Size:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
			Width:       w,
			Height:      h, // positive = bottom-up
			Planes:      1,
			BitCount:    32,
			Compression: biRGB,
		},
	}

	procStretchDIBits.Call(
		hdc,
		0, 0, uintptr(w), uintptr(h),
		0, 0, uintptr(w), uintptr(h),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bi)),
		dibRGBColors, srcCopy,
	)

	// Draw a 2px white border around the selection rectangle.
	if sel != image.ZR {
		pen, _, _ := procCreatePen.Call(psSolid, 2, 0x00FFFFFF) // white
		oldBrush, _, _ := procSelectObject.Call(hdc, uintptr(nullBrushHandle()))
		oldPen, _, _ := procSelectObject.Call(hdc, pen)
		procRectangle.Call(hdc,
			uintptr(sel.Min.X), uintptr(sel.Min.Y),
			uintptr(sel.Max.X), uintptr(sel.Max.Y),
		)
		procSelectObject.Call(hdc, oldPen)
		procSelectObject.Call(hdc, oldBrush)
		procDeleteObject.Call(pen)
	}
}

// nullBrushHandle returns a handle to the stock NULL_BRUSH GDI object.
func nullBrushHandle() uintptr {
	h, _, _ := procGetStockObject.Call(nullBrush)
	return h
}

// ---------------------------------------------------------------------------
// showPlatform — Win32 overlay entry point
// showPlatform displays a fullscreen overlay on the specified monitor that lets the user select a rectangular region.
// 
// showPlatform pins the goroutine to the OS thread, captures the given monitor's screen, presents a topmost
// fullscreen overlay window that accepts mouse and keyboard input to create or cancel a selection, and runs a
// Win32 message loop until the overlay is dismissed. It returns the final selection Result (with Cancelled set
// when the user aborts) or an error if setup (capture, monitor info, or window creation) fails.
//
// monitor is the index of the monitor to capture and display the overlay for.

func showPlatform(monitor int) (*Result, error) {
	// Pin to the OS thread since Win32 window messages are thread-local.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Capture the screen.
	img, err := capture.FullScreen(monitor)
	if err != nil {
		return nil, fmt.Errorf("overlay: screen capture failed: %w", err)
	}

	// Get monitor bounds.
	bounds, err := capture.MonitorInfo(monitor)
	if err != nil {
		return nil, fmt.Errorf("overlay: cannot get monitor info: %w", err)
	}

	// Prepare overlay state.
	ovState.screenshot = img
	ovState.dimmed = DimImage(img, 128) // 50% dimmed
	ovState.sel.Reset()
	ovState.result = nil
	ovState.monitorBounds = bounds

	// Load crosshair cursor.
	ovState.crossCursor, _, _ = procLoadCursorW.Call(0, idcCross)

	// Get module handle.
	hInst, _, _ := procGetModuleHandleW.Call(0)

	// Register window class.
	className, _ := windows.UTF16PtrFromString("ScreensaverOverlay")
	wc := wndClassExW{
		Size:      uint32(unsafe.Sizeof(wndClassExW{})),
		Style:     csHRedraw | csVRedraw,
		WndProc:   windows.NewCallback(wndProc),
		Instance:  hInst,
		Cursor:    ovState.crossCursor,
		ClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// Create fullscreen, topmost popup window.
	windowName, _ := windows.UTF16PtrFromString("Screensaver Overlay")
	hwnd, _, _ := procCreateWindowExW.Call(
		wsExTopmost|wsExToolWindow,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		wsPopup|wsVisible,
		uintptr(bounds.Min.X), uintptr(bounds.Min.Y),
		uintptr(bounds.Dx()), uintptr(bounds.Dy()),
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		return nil, fmt.Errorf("overlay: CreateWindowExW failed")
	}
	ovState.hwnd = hwnd

	procShowWindow.Call(hwnd, swShow)
	procUpdateWindow.Call(hwnd)

	// Message loop.
	var m winMsg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 || int32(ret) == -1 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	if ovState.result == nil {
		return &Result{Cancelled: true}, nil
	}
	return ovState.result, nil
}
