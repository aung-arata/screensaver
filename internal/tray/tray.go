// Package tray provides system tray integration using getlantern/systray.
//
// The tray icon offers quick access to screenshot capture, preferences,
// and quit actions.
package tray

// Config holds system tray configuration.
type Config struct {
	Tooltip string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Tooltip: "Screensaver – Screenshot Tool",
	}
}

// Run starts the system tray icon. onCapture is called when the user
// clicks the "Capture" menu item.
//
// This is a placeholder; the full implementation would use
// github.com/getlantern/systray.
func Run(cfg Config, onCapture func()) error {
	// TODO: implement system tray with getlantern/systray.
	return nil
}
