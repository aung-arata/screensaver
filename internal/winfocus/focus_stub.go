//go:build !windows

package winfocus

import "fmt"

func enumerateWindows() ([]Window, error) {
	return nil, fmt.Errorf("winfocus: window enumeration is only supported on Windows")
}

func focus(hwnd uintptr) error {
	return fmt.Errorf("winfocus: window focusing is only supported on Windows")
}
