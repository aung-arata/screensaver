//go:build !windows

package overlay

import "fmt"

func showPlatform(monitor int) (*Result, error) {
	return nil, fmt.Errorf("overlay: selection overlay not yet implemented for this platform; use --once for headless capture")
}
