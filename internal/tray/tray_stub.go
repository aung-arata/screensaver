//go:build !windows

package tray

import "fmt"

func run(cfg Config, cbs Callbacks) error {
	return fmt.Errorf("tray: system tray not yet implemented for this platform")
}
