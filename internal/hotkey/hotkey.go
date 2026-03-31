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
	done     chan struct{}
	threadID uint32 // OS thread ID; used on Windows for Stop()
}

// NewListener creates a new hotkey listener for the given key combination.
//
// Supported combo format: "ctrl+shift+s", "ctrl+alt+p", etc.
func NewListener(combo string, callback func()) *Listener {
	return &Listener{
		Combo:    combo,
		Callback: callback,
		done:     make(chan struct{}),
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

// Stop terminates the hotkey listener and closes the Done channel.
func (l *Listener) Stop() {
	l.stop()
}

// Done returns a channel that is closed when the listener has stopped.
func (l *Listener) Done() <-chan struct{} {
	return l.done
}

// closeDone closes the done channel if it has not been closed already.
func (l *Listener) closeDone() {
	select {
	case <-l.done:
	default:
		close(l.done)
	}
}
