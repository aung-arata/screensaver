//go:build !windows

package tray

import (
	"os"
	"os/signal"
	"syscall"
)

// runPlatform waits for SIGINT or SIGTERM on non-Windows platforms.
// When a signal arrives, cfg.OnQuit (if set) is called before returning.
func runPlatform(cfg Config) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	<-ch

	if cfg.OnQuit != nil {
		cfg.OnQuit()
	}
	return nil
}
