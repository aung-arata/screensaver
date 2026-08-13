//go:build windows

package scrollcapture

import (
	"fmt"
	"image"
	"time"
	"unsafe"

	"github.com/aung-arata/screensaver/internal/capture"
	"golang.org/x/sys/windows"
)

const (
	inputMouse       = 0
	mouseeventfWheel = 0x0800
)

type mouseInput struct {
	Dx        int32
	Dy        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type input struct {
	Type uint32
	MI   mouseInput
}

var (
	user32                       = windows.NewLazySystemDLL("user32.dll")
	procSendInput                = user32.NewProc("SendInput")
	procSetCursorPos             = user32.NewProc("SetCursorPos")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procWindowFromPoint          = user32.NewProc("WindowFromPoint")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput        = user32.NewProc("AttachThreadInput")
	procGetAncestor              = user32.NewProc("GetAncestor")
)

const gaRoot = 2

type point struct {
	X int32
	Y int32
}

func getCursorPos() (point, error) {
	var p point
	ret, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	if ret == 0 {
		return point{}, fmt.Errorf("GetCursorPos failed: %v", err)
	}
	return p, nil
}

func setCursorPos(p point) error {
	ret, _, err := procSetCursorPos.Call(uintptr(p.X), uintptr(p.Y))
	if ret == 0 {
		return fmt.Errorf("SetCursorPos failed: %v", err)
	}
	return nil
}

func focusWindowAt(x, y int) error {
	p := point{X: int32(x), Y: int32(y)}
	hwnd, _, _ := procWindowFromPoint.Call(uintptr(unsafe.Pointer(&p)))
	if hwnd == 0 {
		return fmt.Errorf("WindowFromPoint: no window at (%d, %d)", x, y)
	}
	if root, _, _ := procGetAncestor.Call(hwnd, gaRoot); root != 0 {
		hwnd = root
	}

	targetThread, _, _ := procGetWindowThreadProcessId.Call(hwnd, 0)
	if targetThread == 0 {
		return fmt.Errorf("GetWindowThreadProcessId: failed to resolve target window thread")
	}

	var fgThread uintptr
	if fgHwnd, _, _ := procGetForegroundWindow.Call(); fgHwnd != 0 {
		fgThread, _, _ = procGetWindowThreadProcessId.Call(fgHwnd, 0)
	}

	attached := false
	if fgThread != 0 && fgThread != targetThread {
		if ret, _, _ := procAttachThreadInput.Call(targetThread, fgThread, 1); ret != 0 {
			attached = true
		}
	}

	ret, _, _ := procSetForegroundWindow.Call(hwnd)

	if attached {
		procAttachThreadInput.Call(targetThread, fgThread, 0)
	}

	if ret == 0 {
		return fmt.Errorf("SetForegroundWindow: failed to focus target window")
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

func capturePlatform(region image.Rectangle, cfg Config) (image.Image, error) {
	if orig, err := getCursorPos(); err == nil {
		defer func() { _ = setCursorPos(orig) }()
	}
	// If getCursorPos fails, capture continues without cursor restoration.

	centerX := region.Min.X + region.Dx()/2
	centerY := region.Min.Y + region.Dy()/2
	if err := focusWindowAt(centerX, centerY); err != nil {
		return nil, err
	}

	prev, err := captureRegion(region)
	if err != nil {
		return nil, err
	}

	parts := []*image.RGBA{cloneRGBA(prev)}
	last := prev
	for i := 1; i < cfg.MaxFrames; i++ {
		if err := sendScrollAtRegionCenter(region, cfg.WheelStep); err != nil {
			return nil, err
		}
		time.Sleep(time.Duration(cfg.DelayMs) * time.Millisecond)

		curr, err := captureRegion(region)
		if err != nil {
			return nil, err
		}
		if framesAreNearIdentical(last, curr) {
			break
		}

		overlap, ok := findBestVerticalOverlap(last, curr)
		if !ok {
			break
		}
		newHeight := curr.Bounds().Dy() - overlap
		if newHeight < minNewContentPx {
			break
		}

		parts = append(parts, cropRows(curr, overlap))
		last = curr
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("long-page capture produced no frames")
	}
	return stitchFrames(parts), nil
}

func captureRegion(region image.Rectangle) (*image.RGBA, error) {
	img, err := capture.CaptureRegion(capture.Region{
		X:      region.Min.X,
		Y:      region.Min.Y,
		Width:  region.Dx(),
		Height: region.Dy(),
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("capture region: %w", err)
	}
	return img, nil
}

func sendScrollAtRegionCenter(region image.Rectangle, step int) error {
	if step == 0 {
		step = defaultWheelStep
	}
	if step < 0 {
		step = -step
	}

	centerX := region.Min.X + region.Dx()/2
	centerY := region.Min.Y + region.Dy()/2
	if err := setCursorPos(point{X: int32(centerX), Y: int32(centerY)}); err != nil {
		return err
	}

	in := input{
		Type: inputMouse,
		MI: mouseInput{
			MouseData: uint32(int32(-step)),
			Flags:     mouseeventfWheel,
		},
	}
	sent, _, sendErr := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&in)),
		unsafe.Sizeof(in),
	)
	if sent != 1 {
		return fmt.Errorf("SendInput failed: %v", sendErr)
	}
	return nil
}
