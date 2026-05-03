//go:build !windows

// Package autostart manages Windows Registry autostart entries.
package autostart

// Install is a no-op stub on non-Windows platforms.
func Install() error { return nil }

// Uninstall is a no-op stub on non-Windows platforms.
func Uninstall() error { return nil }

// IsInstalled always returns false on non-Windows platforms.
func IsInstalled() (bool, error) { return false, nil }
