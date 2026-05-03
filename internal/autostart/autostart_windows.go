//go:build windows

// Package autostart manages Windows Registry autostart entries for the
// current user (HKCU\Software\Microsoft\Windows\CurrentVersion\Run).
package autostart

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	runKey  = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName = "Screensaver"
)

// Install registers the current executable for Windows autostart (current user).
// The exe is launched with no flags so it starts in daemon mode on login.
func Install() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving executable path: %w", err)
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("opening registry key: %w", err)
	}
	defer k.Close()
	return k.SetStringValue(appName, exePath)
}

// Uninstall removes the autostart registry entry.
// Returns nil if the entry does not exist.
func Uninstall() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("opening registry key: %w", err)
	}
	defer k.Close()
	err = k.DeleteValue(appName)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

// IsInstalled returns true if the autostart entry exists.
func IsInstalled() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("opening registry key: %w", err)
	}
	defer k.Close()
	_, _, err = k.GetStringValue(appName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
