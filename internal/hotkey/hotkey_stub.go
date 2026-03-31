//go:build !windows

package hotkey

import "fmt"

func (l *Listener) start() error {
	return fmt.Errorf("hotkey: global hotkey listener not yet implemented for this platform; use --once mode")
}

func (l *Listener) stop() {
	l.closeDone()
}
