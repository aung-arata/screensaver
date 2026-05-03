//go:build !windows

// Package singleinstance provides a named Windows mutex to prevent multiple
// daemon instances from running simultaneously.
package singleinstance

// Acquire is a no-op stub on non-Windows platforms.
// It always reports that this is the first (and only) instance.
func Acquire() bool {
	return true
}
