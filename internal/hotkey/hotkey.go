// Package hotkey provides global hotkey registration.
//
// On Windows this uses RegisterHotKey via golang.org/x/sys/windows.
// On other platforms it serves as a stub that documents the intended
// approach; a full cross-platform implementation would use
// platform-specific APIs or CGo bindings.
package hotkey


// Listener represents a global hotkey listener.
type Listener struct {
	Combo    string
	Callback func()
}

// NewListener creates a new hotkey listener for the given key combination.
//
// Supported combo format: "ctrl+shift+s", "ctrl+alt+p", etc.
func NewListener(combo string, callback func()) *Listener {
	return &Listener{
		Combo:    combo,
		Callback: callback,
	}
}

// Start begins listening for the global hotkey. This blocks until
// Stop is called or the process exits.
//
// The implementation is platform-specific. On Windows it uses
// RegisterHotKey via Win32 APIs. On other platforms it returns an
// error indicating that hotkey registration is not yet available.
func (l *Listener) Start() error {
	return l.start()
}

// Stop terminates the hotkey listener.
func (l *Listener) Stop() {
	// Platform-specific cleanup would go here.
}
