// Package tray provides system tray integration.
//
// On Windows this uses Win32 Shell_NotifyIcon to display a tray icon
// with a context menu.  On other platforms it serves as a stub.
//
// The tray icon offers quick access to screenshot capture, region
// selection, and quit actions.
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

// Callbacks holds the functions invoked by tray menu actions.
// Any callback may be nil, in which case the corresponding menu item
// is silently ignored when clicked.
type Callbacks struct {
	OnCapture      func() // "Capture Screen" menu item
	OnSelectRegion func() // "Select Region" menu item
	OnQuit         func() // "Quit" menu item
}

// Run starts the system tray icon and blocks until the user quits or
// an error occurs.  The implementation is platform-specific.
func Run(cfg Config, cbs Callbacks) error {
	return run(cfg, cbs)
}
