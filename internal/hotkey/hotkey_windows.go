//go:build windows

package hotkey

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                 = windows.NewLazySystemDLL("user32.dll")
	hotkeyKernel32         = windows.NewLazySystemDLL("kernel32.dll")
	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procGetMessage         = user32.NewProc("GetMessageW")
	procPostThreadMessageW = user32.NewProc("PostThreadMessageW")
	procGetCurrentThreadId = hotkeyKernel32.NewProc("GetCurrentThreadId")
)

const (
	modAlt   = 0x0001
	modCtrl  = 0x0002
	modShift = 0x0004

	wmHotkey = 0x0312
	wmQuit   = 0x0012
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
	// Parse the key combination using the cross-platform parser.
	combo, err := ParseCombo(l.Combo)
	if err != nil {
		return err
	}

	// Convert Combo to Win32 modifier flags and virtual-key code.
	var mod uint32
	if combo.Ctrl {
		mod |= modCtrl
	}
	if combo.Alt {
		mod |= modAlt
	}
	if combo.Shift {
		mod |= modShift
	}
	// ASCII uppercase letter as virtual-key code.
	vk := uint32(combo.Key[0])

	// Pin to OS thread: Win32 message loops are thread-local.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Store thread ID so Stop() can post WM_QUIT from another goroutine.
	tid, _, _ := procGetCurrentThreadId.Call()
	atomic.StoreUint32(&l.threadID, uint32(tid))

	ret, _, _ := procRegisterHotKey.Call(0, 1, uintptr(mod), uintptr(vk))
	if ret == 0 {
		return fmt.Errorf("hotkey: RegisterHotKey failed for %q", l.Combo)
	}
	defer procUnregisterHotKey.Call(0, 1)

	var m msg
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		// GetMessageW returns:
		//   >0: message retrieved
		//    0: WM_QUIT (normal termination)
		//   -1: error
		if int32(ret) <= 0 {
			break
		}
		if m.Message == wmHotkey {
			go l.Callback()
		}
	}

	l.closeDone()
	return nil
}

func (l *Listener) stop() {
	tid := atomic.LoadUint32(&l.threadID)
	if tid != 0 {
		// Post WM_QUIT to break the message loop running on the listener thread.
		procPostThreadMessageW.Call(uintptr(tid), wmQuit, 0, 0)
	}
	l.closeDone()
}
