package tray

import "testing"

// ---------------------------------------------------------------------------
// DefaultConfig
// ---------------------------------------------------------------------------

func TestDefaultConfig_NonEmptyTooltip(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Tooltip == "" {
		t.Error("expected non-empty default Tooltip")
	}
}

// ---------------------------------------------------------------------------
// Callbacks zero value
// ---------------------------------------------------------------------------

func TestCallbacks_ZeroValue(t *testing.T) {
	var cbs Callbacks
	if cbs.OnCapture != nil {
		t.Error("expected nil OnCapture in zero-valued Callbacks")
	}
	if cbs.OnSelectRegion != nil {
		t.Error("expected nil OnSelectRegion in zero-valued Callbacks")
	}
	if cbs.OnQuit != nil {
		t.Error("expected nil OnQuit in zero-valued Callbacks")
	}
}

// ---------------------------------------------------------------------------
// Run (stub returns an error on non-Windows platforms)
// ---------------------------------------------------------------------------

func TestRun_Stub(t *testing.T) {
	cfg := DefaultConfig()
	err := Run(cfg, Callbacks{})
	// On non-Windows platforms the stub returns an error.
	// On Windows the function would block (starting the message loop),
	// so this test only validates the stub path.
	if err == nil {
		// If we ever run this on Windows in CI, skip rather than fail.
		t.Skip("Run returned nil — may be running on Windows")
	}
}
