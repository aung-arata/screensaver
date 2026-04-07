// Package tray provides system tray integration.
//
// On Windows this creates a Shell_NotifyIconW tray icon with a context menu
// offering Capture, Select Region, and Quit actions.  On other platforms it
// blocks until the process is terminated.
package tray

// Config holds system tray configuration.
type Config struct {
	Tooltip   string
	OnCapture func() // called when "Take Screenshot" is selected
	OnSelect  func() // called when "Select Region" is selected
	OnQuit    func() // called when "Quit" is selected
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Tooltip: "Screensaver – Screenshot Tool",
	}
}

// Run starts the system tray icon and blocks until the tray is dismissed or
// OnQuit is called.
func Run(cfg Config) error {
	return runPlatform(cfg)
}
