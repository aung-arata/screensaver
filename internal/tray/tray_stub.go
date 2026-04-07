//go:build !windows

package tray

// runPlatform blocks indefinitely on non-Windows platforms so that the daemon
// goroutine stays alive.  The process can still be stopped with Ctrl+C /
// SIGTERM via Go's default signal handling.
func runPlatform(_ Config) error {
	select {} // block forever
}
