//go:build windows

// Package savedialog provides a native Windows Save File dialog.
package savedialog

import (
	"fmt"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Win32 OPENFILENAMEW flags.
const (
	ofnOverwritePrompt = 0x00000002
	ofnPathMustExist   = 0x00000800
	ofnNoChangeDir     = 0x00000008
	maxPath            = 260
)

// openFileName mirrors the Win32 OPENFILENAMEW structure (152 bytes on 64-bit
// Windows). Go's alignment rules produce the same field offsets as the C ABI.
type openFileName struct {
	LStructSize       uint32
	HwndOwner         uintptr
	HInstance         uintptr
	LpstrFilter       uintptr
	LpstrCustomFilter uintptr
	NMaxCustFilter    uint32
	NFilterIndex      uint32
	LpstrFile         uintptr
	NMaxFile          uint32
	LpstrFileTitle    uintptr
	NMaxFileTitle     uint32
	LpstrInitialDir   uintptr
	LpstrTitle        uintptr
	Flags             uint32
	NFileOffset       uint16
	NFileExtension    uint16
	LpstrDefExt       uintptr
	LCustData         uintptr
	LpfnHook          uintptr
	LpTemplateName    uintptr
	PvReserved        uintptr
	DwReserved        uint32
	FlagsEx           uint32
}

// Compile-time assertion: openFileName must be 152 bytes on 64-bit Windows.
var _ [152]byte = [unsafe.Sizeof(openFileName{})]byte{}

var (
	sdComdlg32              = windows.NewLazySystemDLL("comdlg32.dll")
	sdGetSaveFileName       = sdComdlg32.NewProc("GetSaveFileNameW")
	sdCommDlgExtendedError  = sdComdlg32.NewProc("CommDlgExtendedError")
)

// buildFilterUTF16 returns a double-null-terminated UTF-16 filter string
// from a list of (description, extension) pairs, as required by OPENFILENAMEW.
func buildFilterUTF16(pairs [][2]string) []uint16 {
	var result []uint16
	for _, pair := range pairs {
		desc, _ := windows.UTF16FromString(pair[0])
		ext, _ := windows.UTF16FromString(pair[1])
		result = append(result, desc...) // includes null terminator
		result = append(result, ext...)  // includes null terminator
	}
	result = append(result, 0) // extra terminating null
	return result
}

// ShowSaveDialog opens a native Windows Save File dialog.
// ownerHWND may be 0 for a top-level dialog.
// Returns the chosen path, or ("", nil) if the user cancelled.
func ShowSaveDialog(ownerHWND uintptr, defaultDir string) (string, error) {
	filter := buildFilterUTF16([][2]string{
		{"PNG Image", "*.png"},
		{"JPEG Image", "*.jpg;*.jpeg"},
		{"All Files", "*.*"},
	})

	title, err := windows.UTF16PtrFromString("Save Screenshot")
	if err != nil {
		return "", fmt.Errorf("savedialog: encoding title: %w", err)
	}
	defExt, err := windows.UTF16PtrFromString("png")
	if err != nil {
		return "", fmt.Errorf("savedialog: encoding default extension: %w", err)
	}

	// File path buffer (MAX_PATH characters, zero-initialized).
	fileBuf := make([]uint16, maxPath)

	ofn := openFileName{
		LStructSize: uint32(unsafe.Sizeof(openFileName{})),
		HwndOwner:   ownerHWND,
		LpstrFilter: uintptr(unsafe.Pointer(&filter[0])),
		LpstrFile:   uintptr(unsafe.Pointer(&fileBuf[0])),
		NMaxFile:    maxPath,
		LpstrTitle:  uintptr(unsafe.Pointer(title)),
		LpstrDefExt: uintptr(unsafe.Pointer(defExt)),
		Flags:       ofnOverwritePrompt | ofnPathMustExist | ofnNoChangeDir,
	}

	var initialDir *uint16
	if defaultDir != "" {
		initialDir, err = windows.UTF16PtrFromString(defaultDir)
		if err != nil {
			return "", fmt.Errorf("savedialog: encoding initial dir: %w", err)
		}
		ofn.LpstrInitialDir = uintptr(unsafe.Pointer(initialDir))
	}

	ret, _, _ := sdGetSaveFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	// Keep all Go-allocated memory alive until after the syscall returns so the
	// GC cannot collect objects that are only referenced via uintptr fields.
	runtime.KeepAlive(filter)
	runtime.KeepAlive(fileBuf)
	runtime.KeepAlive(title)
	runtime.KeepAlive(defExt)
	runtime.KeepAlive(initialDir)

	if ret == 0 {
		// Distinguish user cancellation from a real error via CommDlgExtendedError.
		code, _, _ := sdCommDlgExtendedError.Call()
		if code == 0 {
			// code == 0 means the user simply cancelled — not an error.
			return "", nil
		}
		return "", fmt.Errorf("savedialog: GetSaveFileNameW failed: CommDlgExtendedError=0x%04x", code)
	}

	return windows.UTF16ToString(fileBuf), nil
}

