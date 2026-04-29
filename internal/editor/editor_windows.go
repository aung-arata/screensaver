//go:build windows

package editor

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"
	"runtime"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/aung-arata/screensaver/internal/savedialog"
)

// ---------------------------------------------------------------------------
// Win32 constants
// ---------------------------------------------------------------------------

const (
	// Window styles
	edWsOverlappedWindow = 0x00CF0000
	edWsVisible          = 0x10000000
	edWsChild            = 0x40000000
	edWsExClientEdge     = 0x00000200

	edCsHRedraw = 0x0002
	edCsVRedraw = 0x0001

	// Messages
	edWmCreate      = 0x0001
	edWmDestroy     = 0x0002
	edWmSize        = 0x0005
	edWmPaint       = 0x000F
	edWmEraseBkgnd  = 0x0014
	edWmSetCursor   = 0x0020
	edWmKeyDown     = 0x0100
	edWmLButtonDown = 0x0201
	edWmLButtonUp   = 0x0202
	edWmMouseMove   = 0x0200
	edWmCommand     = 0x0111
	edWmGetText     = 0x000D
	edWmGetTextLen  = 0x000E

	// Virtual keys
	edVkEscape  = 0x1B
	edVkReturn  = 0x0D
	edVkControl = 0x11
	edVkZ       = 0x5A
	edVkC       = 0x43
	edVkS       = 0x53

	// GDI
	edBiRGB       = 0
	edDibRGBColors = 0
	edSrcCopy     = 0x00CC0020
	edPsSolid     = 0
	edNullBrush   = 5
	edTransparent = 1
	edSwShow      = 5

	// DrawText flags
	edDtCenter     = 0x00000001
	edDtVCenter    = 0x00000004
	edDtSingleLine = 0x00000020

	// GetSystemMetrics
	edSmCXScreen = 0
	edSmCYScreen = 1

	// Edit styles
	edEsAutoHScroll = 0x0080
	edWsBorder      = 0x00800000
	edWsSysMenu     = 0x00080000
	edWsCaption     = 0x00C00000

	// GWLP_WNDPROC
	edGwlpWndProc = ^uintptr(3) // -4 as uintptr

	// Dialog button IDs
	edIdOK     = 1
	edIdCancel = 2

	// Toolbar button IDs
	idBtnPen   = 101
	idBtnRect  = 102
	idBtnArrow = 103
	idBtnText  = 104
	idBtnColor = 105
	idBtnUndo  = 106
	idBtnCopy  = 107
	idBtnSave  = 108

	// Layout
	toolbarHeight int32 = 40
	btnW          int32 = 60
	btnH          int32 = 28
	btnMarginTop  int32 = 6
	minWinW       int32 = 640
	minWinH       int32 = 480

	// ChooseColor flags
	edCcRGBInit = 0x00000001

	// htClient for WM_SETCURSOR
	edHtClient = 1
)

// ---------------------------------------------------------------------------
// Win32 structs
// ---------------------------------------------------------------------------

type edWndClassExW struct {
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

type edWinMsg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      edPoint
}

type edPoint struct {
	X, Y int32
}

type edPaintStruct struct {
	HDC         uintptr
	Erase       int32
	RcPaint     edRect
	Restore     int32
	IncUpdate   int32
	RGBReserved [32]byte
}

type edRect struct {
	Left, Top, Right, Bottom int32
}

type edBitmapInfoHeader struct {
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

type edBitmapInfo struct {
	Header edBitmapInfoHeader
	Colors [1]uint32
}

// chooseColor mirrors the Win32 CHOOSECOLORW structure (72 bytes on amd64).
// Go's struct alignment rules insert padding after LStructSize, RgbResult
// and Flags to match the C ABI.
type chooseColor struct {
	LStructSize    uint32
	HWndOwner      uintptr
	HInstance      uintptr
	RgbResult      uint32
	LpCustColors   uintptr
	Flags          uint32
	LCustData      uintptr
	LpfnHook       uintptr
	LpTemplateName uintptr
}

// ---------------------------------------------------------------------------
// DLL procs
// ---------------------------------------------------------------------------

var (
	edUser32   = windows.NewLazySystemDLL("user32.dll")
	edGdi32    = windows.NewLazySystemDLL("gdi32.dll")
	edKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	edComDlg32 = windows.NewLazySystemDLL("comdlg32.dll")

	edProcRegisterClassExW  = edUser32.NewProc("RegisterClassExW")
	edProcCreateWindowExW   = edUser32.NewProc("CreateWindowExW")
	edProcDestroyWindow     = edUser32.NewProc("DestroyWindow")
	edProcShowWindow        = edUser32.NewProc("ShowWindow")
	edProcUpdateWindow      = edUser32.NewProc("UpdateWindow")
	edProcGetMessageW       = edUser32.NewProc("GetMessageW")
	edProcTranslateMessage  = edUser32.NewProc("TranslateMessage")
	edProcDispatchMessageW  = edUser32.NewProc("DispatchMessageW")
	edProcDefWindowProcW    = edUser32.NewProc("DefWindowProcW")
	edProcPostQuitMessage   = edUser32.NewProc("PostQuitMessage")
	edProcBeginPaint        = edUser32.NewProc("BeginPaint")
	edProcEndPaint          = edUser32.NewProc("EndPaint")
	edProcInvalidateRect    = edUser32.NewProc("InvalidateRect")
	edProcGetClientRect     = edUser32.NewProc("GetClientRect")
	edProcSetCapture        = edUser32.NewProc("SetCapture")
	edProcReleaseCapture    = edUser32.NewProc("ReleaseCapture")
	edProcLoadCursorW       = edUser32.NewProc("LoadCursorW")
	edProcSetCursor         = edUser32.NewProc("SetCursor")
	edProcSendMessageW      = edUser32.NewProc("SendMessageW")
	edProcSetWindowLongPtrW = edUser32.NewProc("SetWindowLongPtrW")
	edProcCallWindowProcW   = edUser32.NewProc("CallWindowProcW")
	edProcGetWindowTextLenW = edUser32.NewProc("GetWindowTextLengthW")
	edProcGetWindowTextW    = edUser32.NewProc("GetWindowTextW")
	edProcEnableWindow      = edUser32.NewProc("EnableWindow")
	edProcSetForegroundWin  = edUser32.NewProc("SetForegroundWindow")
	edProcGetSystemMetrics  = edUser32.NewProc("GetSystemMetrics")
	edProcMoveWindow        = edUser32.NewProc("MoveWindow")
	edProcSetWindowTextW    = edUser32.NewProc("SetWindowTextW")
	edProcPostMessageW      = edUser32.NewProc("PostMessageW")
	edProcFillRect          = edUser32.NewProc("FillRect")
	edProcDrawTextW         = edUser32.NewProc("DrawTextW")
	edProcGetKeyState       = edUser32.NewProc("GetKeyState")
	edProcScreenToClient    = edUser32.NewProc("ScreenToClient")
	edProcSetFocus          = edUser32.NewProc("SetFocus")
	edProcGetCursorPosEd    = edUser32.NewProc("GetCursorPos")
	edProcMessageBoxW       = edUser32.NewProc("MessageBoxW")
	edProcGetModuleHandleW  = edKernel32.NewProc("GetModuleHandleW")
	edProcSetBkModeGDI      = edGdi32.NewProc("SetBkMode")
	edProcSetTextColorGDI   = edGdi32.NewProc("SetTextColor")
	edProcStretchDIBits     = edGdi32.NewProc("StretchDIBits")
	edProcCreatePen         = edGdi32.NewProc("CreatePen")
	edProcCreateSolidBrush  = edGdi32.NewProc("CreateSolidBrush")
	edProcSelectObject      = edGdi32.NewProc("SelectObject")
	edProcGetStockObject    = edGdi32.NewProc("GetStockObject")
	edProcDeleteObject      = edGdi32.NewProc("DeleteObject")
	edProcRectangleGDI      = edGdi32.NewProc("Rectangle")
	edProcMoveToEx          = edGdi32.NewProc("MoveToEx")
	edProcLineTo            = edGdi32.NewProc("LineTo")
	edProcPolyline          = edGdi32.NewProc("Polyline")
	edProcChooseColorW          = edComDlg32.NewProc("ChooseColorW")
	edProcCreateCompatibleDC    = edGdi32.NewProc("CreateCompatibleDC")
	edProcCreateCompatibleBitmap = edGdi32.NewProc("CreateCompatibleBitmap")
	edProcBitBlt                = edGdi32.NewProc("BitBlt")
	edProcDeleteDC              = edGdi32.NewProc("DeleteDC")
)

// ---------------------------------------------------------------------------
// Editor window state
// (package-level to be accessible from window-proc callbacks)
// ---------------------------------------------------------------------------

var edState struct {
	ed          *Editor
	hwnd        uintptr
	activeTool  Tool
	colour      string
	penWidth    float64
	fontSize    float64
	crossCursor uintptr
	arrowCursor uintptr

	// Current drawing gesture
	dragging  bool
	dragSX    float64 // start X in image coords
	dragSY    float64 // start Y in image coords
	dragCX    float64 // current X in image coords
	dragCY    float64 // current Y in image coords
	penStroke *PenStroke

	// Canvas layout (recalculated on WM_SIZE / WM_PAINT)
	imgX, imgY int32   // top-left of displayed image in client coords
	imgW, imgH int32   // displayed image size (after scaling)
	scale      float64 // image→canvas scale factor
	natW, natH int32   // natural image size (pixels)

	// Render cache
	cachedBGRA []byte // bottom-up BGRA buffer for StretchDIBits
	frameDirty bool   // true → regenerate cachedBGRA before next paint

	// Double-buffer backbuffer (cached across frames, recreated only on resize)
	bbDC  uintptr // compatible memory DC
	bbBmp uintptr // compatible bitmap selected into bbDC
	bbW   int32   // width of current backbuffer (0 → not allocated)
	bbH   int32   // height of current backbuffer

	// Guard against multiple simultaneous editors
	running int32 // atomic: 0=idle, 1=running
}

// Text-input dialog state (package-level for callback access)
var textDlgState struct {
	open        bool
	hwnd        uintptr
	editHwnd    uintptr
	text        string
	accepted    bool
	editOrigProc uintptr
}

// Persistent custom colour array for ChooseColorW
var edCustColors [16]uint32

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

// runPlatform opens a native Win32 editor window and blocks until the user
// closes it.  Only one editor may be open at a time.
func (e *Editor) runPlatform() error {
	if !atomic.CompareAndSwapInt32(&edState.running, 0, 1) {
		return fmt.Errorf("editor is already open")
	}
	defer atomic.StoreInt32(&edState.running, 0)

	// All Win32 window creation and message dispatch must run on the same OS
	// thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialise state.
	bounds := e.Image.Bounds()
	edState.ed = e
	edState.activeTool = ToolPen
	edState.colour = e.Config.PenColour
	edState.penWidth = e.Config.PenWidth
	edState.fontSize = e.Config.FontSize
	edState.natW = int32(bounds.Dx())
	edState.natH = int32(bounds.Dy())
	edState.frameDirty = true
	edState.dragging = false
	edState.penStroke = nil

	// Load cursors (IDC_CROSS = 32515, IDC_ARROW = 32512).
	edState.crossCursor, _, _ = edProcLoadCursorW.Call(0, 32515)
	edState.arrowCursor, _, _ = edProcLoadCursorW.Call(0, 32512)

	hInst, _, _ := edProcGetModuleHandleW.Call(0)

	// Register window class.
	className, _ := windows.UTF16PtrFromString("ScreensaverEditor")
	wc := edWndClassExW{
		Size:      uint32(unsafe.Sizeof(edWndClassExW{})),
		Style:     edCsHRedraw | edCsVRedraw,
		WndProc:   windows.NewCallback(editorWndProc),
		Instance:  hInst,
		Cursor:    edState.arrowCursor,
		ClassName: className,
	}
	atom, _, regErr := edProcRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 && regErr != windows.ERROR_CLASS_ALREADY_EXISTS {
		return fmt.Errorf("editor: RegisterClassExW: %v", regErr)
	}

	// Compute initial window size: scale to fit 90% of screen.
	screenW, _, _ := edProcGetSystemMetrics.Call(edSmCXScreen)
	screenH, _, _ := edProcGetSystemMetrics.Call(edSmCYScreen)
	maxW := int32(float64(screenW) * 0.9)
	maxH := int32(float64(screenH) * 0.9)
	scaleX := float64(maxW) / float64(edState.natW)
	scaleY := float64(maxH-toolbarHeight) / float64(edState.natH)
	initialScale := math.Min(scaleX, scaleY)
	if initialScale > 1 {
		initialScale = 1
	}
	winW := int32(math.Max(float64(edState.natW)*initialScale, float64(minWinW)))
	winH := toolbarHeight + int32(math.Max(float64(edState.natH)*initialScale, float64(minWinH-toolbarHeight)))

	startX := (int32(screenW) - winW) / 2
	startY := (int32(screenH) - winH) / 2

	// Create the editor window.
	title, _ := windows.UTF16PtrFromString("Screensaver – Editor   [Ctrl+Z Undo | Ctrl+C Copy | Ctrl+S Save | Esc Close]")
	hwnd, _, cwErr := edProcCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		edWsOverlappedWindow|edWsVisible,
		uintptr(startX), uintptr(startY),
		uintptr(winW), uintptr(winH),
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("editor: CreateWindowExW: %v", cwErr)
	}
	edState.hwnd = hwnd

	edProcShowWindow.Call(hwnd, edSwShow)
	edProcUpdateWindow.Call(hwnd)

	// Message loop.
	var m edWinMsg
	for {
		ret, _, _ := edProcGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == -1 {
			return fmt.Errorf("editor: GetMessageW: %v", windows.GetLastError())
		}
		if ret == 0 {
			break
		}
		edProcTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		edProcDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

// ---------------------------------------------------------------------------
// Window procedure
// ---------------------------------------------------------------------------

func editorWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case edWmEraseBkgnd:
		return 1 // prevent flicker

	case edWmSetCursor:
		// Use crosshair in the canvas, arrow in the toolbar.
		if edLoWord(lParam) == edHtClient {
			// Convert cursor screen position to client coordinates so we can
			// check whether it is inside the toolbar strip.
			var pt edPoint
			edProcGetCursorPosEd.Call(uintptr(unsafe.Pointer(&pt)))
			edProcScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&pt)))
			if pt.Y < toolbarHeight {
				edProcSetCursor.Call(edState.arrowCursor)
			} else {
				edProcSetCursor.Call(edState.crossCursor)
			}
			return 1
		}

	case edWmPaint:
		var ps edPaintStruct
		hdc, _, _ := edProcBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		paintEditor(hwnd, hdc)
		edProcEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case edWmSize:
		// Recompute layout so the image scales to the new canvas size.
		var cr edRect
		edProcGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&cr)))
		computeLayout(cr.Right, cr.Bottom)
		// Invalidate the cached backbuffer so it is recreated at the new size.
		edDestroyBackbuffer()
		edProcInvalidateRect.Call(hwnd, 0, 0)
		return 0

	case edWmCommand:
		handleToolbarCommand(hwnd, edLoWord(wParam))
		return 0

	case edWmKeyDown:
		ctrl := edGetKeyState(edVkControl)
		switch {
		case ctrl && wParam == edVkZ:
			doUndo()
		case ctrl && wParam == edVkC:
			doCopy()
		case ctrl && wParam == edVkS:
			doSave()
		case wParam == edVkEscape:
			edProcDestroyWindow.Call(hwnd)
		}
		return 0

	case edWmLButtonDown:
		x, y := int32(edLoWord(lParam)), int32(edHiWord(lParam))
		if y < toolbarHeight {
			if id := toolbarHitTest(x, y); id != 0 {
				handleToolbarCommand(hwnd, id)
			}
			return 0
		}
		ix, iy := canvasToImage(x, y)
		ix, iy = clampToImage(ix, iy)
		edState.dragging = true
		edState.dragSX, edState.dragSY = ix, iy
		edState.dragCX, edState.dragCY = ix, iy
		if edState.activeTool == ToolPen {
			edState.penStroke = &PenStroke{
				Points: []Point{{ix, iy}},
				Colour: edState.colour,
				Width:  edState.penWidth,
			}
		}
		if edState.activeTool == ToolText {
			// Text tool: no drag, just a click → open input dialog.
			edState.dragging = false
			text := showTextInput(hwnd)
			if text != "" {
				edState.ed.AddAnnotation(&TextAnnotation{
					X: ix, Y: iy,
					Content: text,
					Colour:  edState.colour,
					Size:    edState.fontSize,
				})
				edState.frameDirty = true
				edProcInvalidateRect.Call(hwnd, 0, 0)
			}
			return 0
		}
		edProcSetCapture.Call(hwnd)
		return 0

	case edWmMouseMove:
		if !edState.dragging {
			return 0
		}
		x, y := int32(edLoWord(lParam)), int32(edHiWord(lParam))
		ix, iy := canvasToImage(x, y)
		ix, iy = clampToImage(ix, iy)
		edState.dragCX, edState.dragCY = ix, iy
		if edState.activeTool == ToolPen && edState.penStroke != nil {
			edState.penStroke.Points = append(edState.penStroke.Points, Point{ix, iy})
		}
		edProcInvalidateRect.Call(hwnd, 0, 0)
		return 0

	case edWmLButtonUp:
		if !edState.dragging {
			return 0
		}
		x, y := int32(edLoWord(lParam)), int32(edHiWord(lParam))
		ix, iy := canvasToImage(x, y)
		ix, iy = clampToImage(ix, iy)
		edState.dragCX, edState.dragCY = ix, iy
		edProcReleaseCapture.Call()

		commitDragAnnotation(ix, iy)
		edState.dragging = false
		edState.penStroke = nil
		edState.frameDirty = true
		edProcInvalidateRect.Call(hwnd, 0, 0)
		return 0

	case edWmDestroy:
		edDestroyBackbuffer()
		edProcPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := edProcDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// handleToolbarCommand processes toolbar button clicks.
func handleToolbarCommand(hwnd uintptr, id int) {
	switch id {
	case idBtnPen:
		edState.activeTool = ToolPen
		updateTitle(hwnd)
	case idBtnRect:
		edState.activeTool = ToolRect
		updateTitle(hwnd)
	case idBtnArrow:
		edState.activeTool = ToolArrow
		updateTitle(hwnd)
	case idBtnText:
		edState.activeTool = ToolText
		updateTitle(hwnd)
	case idBtnColor:
		openColorPicker(hwnd)
	case idBtnUndo:
		doUndo()
	case idBtnCopy:
		doCopy()
	case idBtnSave:
		doSave()
	}
}

// commitDragAnnotation adds the completed drag annotation to the editor.
func commitDragAnnotation(endX, endY float64) {
	sx, sy := edState.dragSX, edState.dragSY
	switch edState.activeTool {
	case ToolPen:
		if edState.penStroke != nil && len(edState.penStroke.Points) >= 2 {
			edState.ed.AddAnnotation(edState.penStroke)
		}
	case ToolRect:
		if math.Abs(endX-sx) > 1 || math.Abs(endY-sy) > 1 {
			edState.ed.AddAnnotation(&RectAnnotation{
				X1: sx, Y1: sy, X2: endX, Y2: endY,
				Colour: edState.colour, Width: edState.penWidth,
			})
		}
	case ToolArrow:
		if math.Abs(endX-sx) > 1 || math.Abs(endY-sy) > 1 {
			edState.ed.AddAnnotation(&ArrowAnnotation{
				X1: sx, Y1: sy, X2: endX, Y2: endY,
				Colour: edState.colour, Width: edState.penWidth,
			})
		}
	}
}

// ---------------------------------------------------------------------------
// Painting
// ---------------------------------------------------------------------------

// edDestroyBackbuffer frees the cached off-screen DC and bitmap.
func edDestroyBackbuffer() {
	if edState.bbDC != 0 {
		if edState.bbBmp != 0 {
			edProcDeleteObject.Call(edState.bbBmp)
			edState.bbBmp = 0
		}
		edProcDeleteDC.Call(edState.bbDC)
		edState.bbDC = 0
	}
	edState.bbW = 0
	edState.bbH = 0
}

// paintEditor renders the full editor: toolbar + image canvas.
func paintEditor(hwnd, hdc uintptr) {
	var cr edRect
	edProcGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&cr)))
	clientW, clientH := cr.Right, cr.Bottom

	// Skip painting for zero-size / minimised windows.
	if clientW <= 0 || clientH <= 0 {
		return
	}

	// Ensure layout is computed.
	if edState.imgW == 0 {
		computeLayout(clientW, clientH)
	}

	// Regenerate the BGRA cache when dirty.
	if edState.frameDirty {
		updateCachedFrame()
		edState.frameDirty = false
	}

	// Lazily create (or recreate after resize) the persistent backbuffer.
	if edState.bbDC == 0 || edState.bbW != clientW || edState.bbH != clientH {
		edDestroyBackbuffer()
		memDC, _, _ := edProcCreateCompatibleDC.Call(hdc)
		memBmp, _, _ := edProcCreateCompatibleBitmap.Call(hdc, uintptr(clientW), uintptr(clientH))
		if memDC == 0 || memBmp == 0 {
			// GDI allocation failed (e.g., out of resources); fall back to
			// painting directly on the window DC.
			if memDC != 0 {
				edProcDeleteDC.Call(memDC)
			}
			if memBmp != 0 {
				edProcDeleteObject.Call(memBmp)
			}
			edPaintOnDC(hdc, clientW, clientH)
			return
		}
		edProcSelectObject.Call(memDC, memBmp)
		edState.bbDC = memDC
		edState.bbBmp = memBmp
		edState.bbW = clientW
		edState.bbH = clientH
	}

	// Paint all content into the persistent off-screen backbuffer.
	edPaintOnDC(edState.bbDC, clientW, clientH)

	// Blit the completed frame to the real window DC in one atomic operation.
	edProcBitBlt.Call(hdc, 0, 0, uintptr(clientW), uintptr(clientH), edState.bbDC, 0, 0, edSrcCopy)
}

// edPaintOnDC performs all drawing into the given DC (either the cached
// backbuffer or, as a fallback, the real window DC).
func edPaintOnDC(dc uintptr, clientW, clientH int32) {
	// ---------- canvas ----------

	// Fill canvas background (black).
	canvasTop := toolbarHeight
	brushBlack, _, _ := edProcCreateSolidBrush.Call(0x00000000)
	canvasR := edRect{0, canvasTop, clientW, clientH}
	edProcFillRect.Call(dc, uintptr(unsafe.Pointer(&canvasR)), brushBlack)
	edProcDeleteObject.Call(brushBlack)

	// Blit the (possibly scaled) cached image.
	if len(edState.cachedBGRA) > 0 {
		bi := edBitmapInfo{
			Header: edBitmapInfoHeader{
				Size:        uint32(unsafe.Sizeof(edBitmapInfoHeader{})),
				Width:       edState.natW,
				Height:      edState.natH, // positive = bottom-up
				Planes:      1,
				BitCount:    32,
				Compression: edBiRGB,
			},
		}
		edProcStretchDIBits.Call(
			dc,
			uintptr(edState.imgX), uintptr(edState.imgY),
			uintptr(edState.imgW), uintptr(edState.imgH),
			0, 0,
			uintptr(edState.natW), uintptr(edState.natH),
			uintptr(unsafe.Pointer(&edState.cachedBGRA[0])),
			uintptr(unsafe.Pointer(&bi)),
			edDibRGBColors, edSrcCopy,
		)
	}

	// Draw live drag preview using GDI (no gg re-render needed).
	if edState.dragging {
		drawDragPreview(dc)
	}

	// ---------- toolbar ----------
	drawToolbar(dc, clientW)
}

// drawDragPreview draws the in-progress annotation using GDI.
func drawDragPreview(hdc uintptr) {
	colorref := hexToColorref(edState.colour)
	penWidth := int32(math.Max(1, edState.penWidth))
	pen, _, _ := edProcCreatePen.Call(edPsSolid, uintptr(penWidth), uintptr(colorref))
	nullBrushH, _, _ := edProcGetStockObject.Call(edNullBrush)
	oldPen, _, _ := edProcSelectObject.Call(hdc, pen)
	oldBrush, _, _ := edProcSelectObject.Call(hdc, nullBrushH)
	defer func() {
		edProcSelectObject.Call(hdc, oldPen)
		edProcSelectObject.Call(hdc, oldBrush)
		edProcDeleteObject.Call(pen)
	}()

	sx, sy := imageToCanvas(edState.dragSX, edState.dragSY)
	cx, cy := imageToCanvas(edState.dragCX, edState.dragCY)

	switch edState.activeTool {
	case ToolRect:
		edProcRectangleGDI.Call(hdc, uintptr(sx), uintptr(sy), uintptr(cx), uintptr(cy))
	case ToolArrow:
		edProcMoveToEx.Call(hdc, uintptr(sx), uintptr(sy), 0)
		edProcLineTo.Call(hdc, uintptr(cx), uintptr(cy))
	case ToolPen:
		if edState.penStroke == nil || len(edState.penStroke.Points) < 2 {
			return
		}
		pts := make([]edPoint, len(edState.penStroke.Points))
		for i, p := range edState.penStroke.Points {
			px, py := imageToCanvas(p.X, p.Y)
			pts[i] = edPoint{px, py}
		}
		edProcPolyline.Call(hdc, uintptr(unsafe.Pointer(&pts[0])), uintptr(len(pts)))
	}
}

// ---------------------------------------------------------------------------
// Toolbar rendering (painted manually for active-state highlighting)
// ---------------------------------------------------------------------------

type toolbarBtn struct {
	id   int
	text string
	tool Tool // non-zero only for the four drawing tools
}

var toolbarBtns = []toolbarBtn{
	{idBtnPen, "Pen", ToolPen},
	{idBtnRect, "Rect", ToolRect},
	{idBtnArrow, "Arrow", ToolArrow},
	{idBtnText, "Text", ToolText},
	{idBtnColor, "Color", ""},
	{idBtnUndo, "Undo", ""},
	{idBtnCopy, "Copy", ""},
	{idBtnSave, "Save", ""},
}

// toolbarBtnRect returns the bounding rectangle of button at position i.
func toolbarBtnRect(i int) edRect {
	// Group separator gap after Text (index 3) and Color (index 4).
	x := int32(8)
	for j := 0; j < i; j++ {
		x += btnW + 4
		if j == 3 || j == 4 {
			x += 8 // extra gap before Color, before Undo
		}
	}
	return edRect{
		Left:   x,
		Top:    btnMarginTop,
		Right:  x + btnW,
		Bottom: btnMarginTop + btnH,
	}
}

// drawToolbar paints the toolbar strip at the top of the window.
func drawToolbar(hdc uintptr, clientW int32) {
	// Toolbar background.
	bgBrush, _, _ := edProcCreateSolidBrush.Call(0x00F0F0F0)
	tbRect := edRect{0, 0, clientW, toolbarHeight}
	edProcFillRect.Call(hdc, uintptr(unsafe.Pointer(&tbRect)), bgBrush)
	edProcDeleteObject.Call(bgBrush)

	edProcSetBkModeGDI.Call(hdc, edTransparent)

	for i, btn := range toolbarBtns {
		r := toolbarBtnRect(i)
		active := btn.tool != "" && btn.tool == edState.activeTool

		// Button fill: blue for active tool, white for others; use current
		// colour swatch for the Color button.
		var fillColor uint32
		switch {
		case btn.id == idBtnColor:
			fillColor = hexToColorref(edState.colour)
		case active:
			fillColor = 0x00D46E00 // dark orange-blue accent (BGR)
		default:
			fillColor = 0x00FFFFFF
		}
		btnBrush, _, _ := edProcCreateSolidBrush.Call(uintptr(fillColor))
		edProcFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), btnBrush)
		edProcDeleteObject.Call(btnBrush)

		// Border.
		borderPen, _, _ := edProcCreatePen.Call(edPsSolid, 1, 0x00AAAAAA)
		oldPen, _, _ := edProcSelectObject.Call(hdc, borderPen)
		nullBrushH, _, _ := edProcGetStockObject.Call(edNullBrush)
		oldBrush, _, _ := edProcSelectObject.Call(hdc, nullBrushH)
		edProcRectangleGDI.Call(hdc, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right), uintptr(r.Bottom))
		edProcSelectObject.Call(hdc, oldPen)
		edProcSelectObject.Call(hdc, oldBrush)
		edProcDeleteObject.Call(borderPen)

		// Text label.
		textColor := uint32(0x00000000)
		if active {
			textColor = 0x00FFFFFF
		}
		if btn.id == idBtnColor {
			// Choose black or white text based on perceived brightness.
			cr := fillColor
			rb := cr & 0xFF
			gb := (cr >> 8) & 0xFF
			bb := (cr >> 16) & 0xFF
			brightness := (rb*299 + gb*587 + bb*114) / 1000
			if brightness < 128 {
				textColor = 0x00FFFFFF
			}
		}
		edProcSetTextColorGDI.Call(hdc, uintptr(textColor))
		text, _ := windows.UTF16PtrFromString(btn.text)
		edProcDrawTextW.Call(
			hdc,
			uintptr(unsafe.Pointer(text)), ^uintptr(0),
			uintptr(unsafe.Pointer(&r)),
			edDtCenter|edDtVCenter|edDtSingleLine,
		)
	}
}

// toolbarHitTest returns the button ID under the given client-coord point, or
// 0 if none.
func toolbarHitTest(x, y int32) int {
	if y < 0 || y >= toolbarHeight {
		return 0
	}
	for i, btn := range toolbarBtns {
		r := toolbarBtnRect(i)
		if x >= r.Left && x < r.Right && y >= r.Top && y < r.Bottom {
			return btn.id
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// Layout and coordinate mapping
// ---------------------------------------------------------------------------

// computeLayout recalculates how the image fits inside the canvas.
func computeLayout(clientW, clientH int32) {
	canvasW := clientW
	canvasH := clientH - toolbarHeight
	if canvasW <= 0 || canvasH <= 0 || edState.natW == 0 || edState.natH == 0 {
		return
	}
	scaleX := float64(canvasW) / float64(edState.natW)
	scaleY := float64(canvasH) / float64(edState.natH)
	s := math.Min(scaleX, scaleY)
	if s > 1 {
		s = 1
	}
	edState.scale = s
	edState.imgW = int32(float64(edState.natW) * s)
	edState.imgH = int32(float64(edState.natH) * s)
	edState.imgX = (canvasW - edState.imgW) / 2
	edState.imgY = toolbarHeight + (canvasH-edState.imgH)/2
}

// canvasToImage maps client-area coordinates to image coordinates.
func canvasToImage(cx, cy int32) (float64, float64) {
	if edState.scale == 0 {
		return 0, 0
	}
	ix := float64(cx-edState.imgX) / edState.scale
	iy := float64(cy-edState.imgY) / edState.scale
	return ix, iy
}

// imageToCanvas maps image coordinates to client-area coordinates.
func imageToCanvas(ix, iy float64) (int32, int32) {
	cx := int32(ix*edState.scale) + edState.imgX
	cy := int32(iy*edState.scale) + edState.imgY
	return cx, cy
}

// clampToImage clamps coordinates to the image bounds.
func clampToImage(ix, iy float64) (float64, float64) {
	ix = math.Max(0, math.Min(float64(edState.natW-1), ix))
	iy = math.Max(0, math.Min(float64(edState.natH-1), iy))
	return ix, iy
}

// ---------------------------------------------------------------------------
// Render cache
// ---------------------------------------------------------------------------

// updateCachedFrame re-renders all committed annotations via gg and converts
// the result to a bottom-up BGRA buffer suitable for StretchDIBits.
func updateCachedFrame() {
	rendered, err := edState.ed.Render()
	if err != nil {
		return
	}
	rgba, ok := rendered.(*image.RGBA)
	if !ok {
		// Convert to *image.RGBA using image/draw.
		bounds := rendered.Bounds()
		tmp := image.NewRGBA(bounds)
		draw.Draw(tmp, bounds, rendered, bounds.Min, draw.Src)
		rgba = tmp
	}
	w := int32(rgba.Bounds().Dx())
	h := int32(rgba.Bounds().Dy())
	stride := rgba.Stride
	needed := int(w) * int(h) * 4
	if len(edState.cachedBGRA) != needed {
		edState.cachedBGRA = make([]byte, needed)
	}
	// Convert RGBA top-down → BGRA bottom-up.
	for y := 0; y < int(h); y++ {
		srcRow := rgba.Pix[y*stride : y*stride+int(w)*4]
		dstRow := edState.cachedBGRA[(int(h)-1-y)*int(w)*4 : (int(h)-y)*int(w)*4]
		for x := 0; x < int(w)*4; x += 4 {
			dstRow[x+0] = srcRow[x+2] // B
			dstRow[x+1] = srcRow[x+1] // G
			dstRow[x+2] = srcRow[x+0] // R
			dstRow[x+3] = srcRow[x+3] // A
		}
	}
}

// ---------------------------------------------------------------------------
// Actions
// ---------------------------------------------------------------------------

func doUndo() {
	edState.ed.Undo()
	edState.frameDirty = true
	edProcInvalidateRect.Call(edState.hwnd, 0, 0)
}

func doCopy() {
	if err := edState.ed.CopyToClipboard(); err != nil {
		// Non-fatal: just ignore clipboard errors in daemon mode.
		_ = err
	}
}

func doSave() {
	path, err := savedialog.ShowSaveDialog(edState.hwnd, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[editor] ShowSaveDialog: %v\n", err)
		return
	}
	if path == "" {
		return // user cancelled
	}
	if err := edState.ed.Save(path); err != nil {
		fmt.Fprintf(os.Stderr, "[editor] Save: %v\n", err)
		edShowMessage(edState.hwnd, fmt.Sprintf("Save failed: %v", err))
		return
	}
	if edState.ed.OnSave != nil {
		edState.ed.OnSave(path)
	}
	edShowMessage(edState.hwnd, fmt.Sprintf("Saved to %s", path))
}

// updateTitle sets the window title to include the active tool name.
func updateTitle(hwnd uintptr) {
	title := fmt.Sprintf("Screensaver – Editor [%s]  Ctrl+Z Undo | Ctrl+C Copy | Ctrl+S Save | Esc Close", edState.activeTool)
	titlePtr, _ := windows.UTF16PtrFromString(title)
	edProcSetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(titlePtr)))
	edProcInvalidateRect.Call(hwnd, 0, 0)
}

// ---------------------------------------------------------------------------
// Colour picker
// ---------------------------------------------------------------------------

func openColorPicker(owner uintptr) {
	cc := chooseColor{
		LStructSize:  uint32(unsafe.Sizeof(chooseColor{})),
		HWndOwner:    owner,
		RgbResult:    hexToColorref(edState.colour),
		LpCustColors: uintptr(unsafe.Pointer(&edCustColors[0])),
		Flags:        edCcRGBInit,
	}
	ok, _, _ := edProcChooseColorW.Call(uintptr(unsafe.Pointer(&cc)))
	if ok != 0 {
		edState.colour = colorrefToHex(cc.RgbResult)
		edProcInvalidateRect.Call(owner, 0, 0) // redraw toolbar colour swatch
	}
}

// ---------------------------------------------------------------------------
// Text-input dialog
// ---------------------------------------------------------------------------

// showTextInput shows a small modal dialog that lets the user type a string.
// It returns the entered text, or "" if cancelled.
func showTextInput(owner uintptr) string {
	textDlgState.open = false
	textDlgState.text = ""
	textDlgState.accepted = false
	textDlgState.editHwnd = 0
	textDlgState.editOrigProc = 0

	hInst, _, _ := edProcGetModuleHandleW.Call(0)

	// Register the dialog window class (once; ignore already-registered error).
	dlgClass, _ := windows.UTF16PtrFromString("ScreensaverTextInput")
	dlgWC := edWndClassExW{
		Size:      uint32(unsafe.Sizeof(edWndClassExW{})),
		WndProc:   windows.NewCallback(textDlgWndProc),
		Instance:  hInst,
		ClassName: dlgClass,
	}
	dlgAtom, _, dlgErr := edProcRegisterClassExW.Call(uintptr(unsafe.Pointer(&dlgWC)))
	if dlgAtom == 0 && dlgErr != windows.ERROR_CLASS_ALREADY_EXISTS {
		return ""
	}

	// Create a small dialog window.
	dlgTitle, _ := windows.UTF16PtrFromString("Enter text")
	dlgW, dlgH := int32(340), int32(110)
	// Position dialog centered on screen.
	screenW, _, _ := edProcGetSystemMetrics.Call(edSmCXScreen)
	screenH, _, _ := edProcGetSystemMetrics.Call(edSmCYScreen)
	dlgX := (int32(screenW) - dlgW) / 2
	dlgY := (int32(screenH) - dlgH) / 2

	dlgHwnd, _, _ := edProcCreateWindowExW.Call(
		edWsExClientEdge,
		uintptr(unsafe.Pointer(dlgClass)),
		uintptr(unsafe.Pointer(dlgTitle)),
		edWsCaption|edWsSysMenu|edWsVisible,
		uintptr(dlgX), uintptr(dlgY),
		uintptr(dlgW), uintptr(dlgH),
		owner, 0, hInst, 0,
	)
	if dlgHwnd == 0 {
		return ""
	}
	textDlgState.hwnd = dlgHwnd
	textDlgState.open = true

	// Disable owner while dialog is open (modal behaviour).
	edProcEnableWindow.Call(owner, 0)

	// Create child EDIT control.
	editClass, _ := windows.UTF16PtrFromString("EDIT")
	editHwnd, _, _ := edProcCreateWindowExW.Call(
		edWsExClientEdge,
		uintptr(unsafe.Pointer(editClass)),
		0,
		edWsChild|edWsVisible|edEsAutoHScroll,
		uintptr(10), uintptr(10), uintptr(dlgW-20-2*2), uintptr(24),
		dlgHwnd, 0, hInst, 0,
	)
	textDlgState.editHwnd = editHwnd

	// Subclass the EDIT to intercept Enter/Escape.
	editCB := windows.NewCallback(textInputEditProc)
	origProc, _, _ := edProcSetWindowLongPtrW.Call(editHwnd, edGwlpWndProc, editCB)
	textDlgState.editOrigProc = origProc

	// Create OK button (default).
	btnClass, _ := windows.UTF16PtrFromString("BUTTON")
	btnOKText, _ := windows.UTF16PtrFromString("OK")
	edProcCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(btnOKText)),
		edWsChild|edWsVisible|0x00000001, // BS_DEFPUSHBUTTON
		uintptr(dlgW/2-90), uintptr(50), 80, 26,
		dlgHwnd, edIdOK, hInst, 0,
	)
	// Create Cancel button.
	btnCancelText, _ := windows.UTF16PtrFromString("Cancel")
	edProcCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(btnClass)),
		uintptr(unsafe.Pointer(btnCancelText)),
		edWsChild|edWsVisible,
		uintptr(dlgW/2+10), uintptr(50), 80, 26,
		dlgHwnd, edIdCancel, hInst, 0,
	)

	// Focus the edit control so the user can type immediately.
	edProcSetFocus.Call(editHwnd)

	edProcShowWindow.Call(dlgHwnd, edSwShow)
	edProcUpdateWindow.Call(dlgHwnd)

	// Nested message loop until dialog closes.
	var m edWinMsg
	for textDlgState.open {
		ret, _, _ := edProcGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		edProcTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		edProcDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	// Re-enable owner.
	edProcEnableWindow.Call(owner, 1)
	edProcSetForegroundWin.Call(owner)

	if textDlgState.accepted {
		return textDlgState.text
	}
	return ""
}

// textDlgWndProc handles messages for the text-input dialog window.
func textDlgWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case edWmCommand:
		id := edLoWord(wParam)
		switch id {
		case edIdOK:
			textDlgState.text = getWindowText(textDlgState.editHwnd)
			textDlgState.accepted = true
			textDlgState.open = false
			edProcDestroyWindow.Call(hwnd)
			return 0
		case edIdCancel:
			textDlgState.accepted = false
			textDlgState.open = false
			edProcDestroyWindow.Call(hwnd)
			return 0
		}
	case edWmDestroy:
		textDlgState.open = false
		return 0
	}
	ret, _, _ := edProcDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// textInputEditProc is the subclassed window proc for the EDIT control in the
// text-input dialog.  It forwards Enter → OK and Escape → Cancel.
func textInputEditProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if msg == edWmKeyDown {
		switch wParam {
		case edVkReturn:
			edProcPostMessageW.Call(textDlgState.hwnd, edWmCommand, edIdOK, 0)
			return 0
		case edVkEscape:
			edProcPostMessageW.Call(textDlgState.hwnd, edWmCommand, edIdCancel, 0)
			return 0
		}
	}
	r, _, _ := edProcCallWindowProcW.Call(
		textDlgState.editOrigProc, hwnd, msg, wParam, lParam,
	)
	return r
}

// getWindowText reads the text of a Win32 window (typically an EDIT control).
func getWindowText(hwnd uintptr) string {
	n, _, _ := edProcGetWindowTextLenW.Call(hwnd)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	edProcGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return windows.UTF16ToString(buf)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// edShowMessage shows a simple Win32 MessageBox with the given text.
func edShowMessage(owner uintptr, text string) {
	msg, _ := windows.UTF16PtrFromString(text)
	title, _ := windows.UTF16PtrFromString("Screensaver – Editor")
	edProcMessageBoxW.Call(owner, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), 0)
}

// hexToColorref converts a "#RRGGBB" string to a Win32 COLORREF (0x00BBGGRR).
func hexToColorref(hex string) uint32 {
	if len(hex) != 7 || hex[0] != '#' {
		return 0
	}
	nibble := func(c byte) uint32 {
		switch {
		case c >= '0' && c <= '9':
			return uint32(c - '0')
		case c >= 'A' && c <= 'F':
			return uint32(c-'A') + 10
		case c >= 'a' && c <= 'f':
			return uint32(c-'a') + 10
		}
		return 0
	}
	r := nibble(hex[1])<<4 | nibble(hex[2])
	g := nibble(hex[3])<<4 | nibble(hex[4])
	b := nibble(hex[5])<<4 | nibble(hex[6])
	return b<<16 | g<<8 | r
}

// colorrefToHex converts a Win32 COLORREF (0x00BBGGRR) to a "#RRGGBB" string.
func colorrefToHex(c uint32) string {
	r := c & 0xFF
	g := (c >> 8) & 0xFF
	b := (c >> 16) & 0xFF
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// edLoWord extracts the low-order 16 bits (signed).
func edLoWord(l uintptr) int { return int(int16(l & 0xFFFF)) }

// edHiWord extracts the high-order 16 bits (signed).
func edHiWord(l uintptr) int { return int(int16((l >> 16) & 0xFFFF)) }

// edGetKeyState returns true when the given virtual key is pressed.
func edGetKeyState(vk uintptr) bool {
	state, _, _ := edProcGetKeyState.Call(vk)
	return (state & 0x8000) != 0
}
