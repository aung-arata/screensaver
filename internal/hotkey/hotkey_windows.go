//go:build windows

package hotkey

import (
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32              = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey  = user32.NewProc("RegisterHotKey")
	procGetMessage      = user32.NewProc("GetMessageW")
)

const (
	modAlt   = 0x0001
	modCtrl  = 0x0002
	modShift = 0x0004

	wmHotkey = 0x0312
)

// msg mirrors the Win32 MSG structure.
type msg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

func (l *Listener) start() error {
	mod, vk, err := parseCombo(l.Combo)
	if err != nil {
		return err
	}

	ret, _, _ := procRegisterHotKey.Call(0, 1, uintptr(mod), uintptr(vk))
	if ret == 0 {
		return fmt.Errorf("hotkey: RegisterHotKey failed for %q", l.Combo)
	}

	var m msg
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 {
			break
		}
		if m.Message == wmHotkey {
			go l.Callback()
		}
	}
	return nil
}

// parseCombo converts a combo string like "ctrl+shift+s" into Win32
// modifier flags and a virtual-key code.
func parseCombo(combo string) (uint32, uint32, error) {
	parts := strings.Split(strings.ToLower(combo), "+")
	if len(parts) == 0 {
		return 0, 0, fmt.Errorf("hotkey: empty combo")
	}

	var mod uint32
	var key string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch p {
		case "ctrl", "control":
			mod |= modCtrl
		case "alt":
			mod |= modAlt
		case "shift":
			mod |= modShift
		default:
			key = p
		}
	}

	if key == "" || len(key) != 1 {
		return 0, 0, fmt.Errorf("hotkey: unsupported key %q in combo %q", key, combo)
	}

	// ASCII uppercase letter as virtual-key code.
	vk := uint32(strings.ToUpper(key)[0])
	return mod, vk, nil
}
