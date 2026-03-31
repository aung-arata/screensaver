package hotkey

import "testing"

// ---------------------------------------------------------------------------
// NewListener
// ---------------------------------------------------------------------------

func TestNewListener_SetsCombo(t *testing.T) {
	l := NewListener("ctrl+shift+s", func() {})
	if l.Combo != "ctrl+shift+s" {
		t.Errorf("got %q, want %q", l.Combo, "ctrl+shift+s")
	}
}

func TestNewListener_CallbackNotNil(t *testing.T) {
	called := false
	l := NewListener("ctrl+shift+s", func() { called = true })
	l.Callback()
	if !called {
		t.Error("expected callback to be called")
	}
}

func TestNewListener_DoneChannel(t *testing.T) {
	l := NewListener("ctrl+shift+s", func() {})
	if l.Done() == nil {
		t.Error("expected non-nil Done channel")
	}
}

// ---------------------------------------------------------------------------
// Stop / Done
// ---------------------------------------------------------------------------

func TestListener_StopClosesDone(t *testing.T) {
	l := NewListener("ctrl+shift+s", func() {})
	l.Stop()
	select {
	case <-l.Done():
		// OK — channel closed
	default:
		t.Error("expected Done channel to be closed after Stop")
	}
}

func TestListener_StopIdempotent(t *testing.T) {
	l := NewListener("ctrl+shift+s", func() {})
	l.Stop()
	l.Stop() // must not panic
}
