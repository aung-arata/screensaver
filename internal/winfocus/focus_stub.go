//go:build !windows

package winfocus

import "fmt"

func focusAt(x, y int) error {
	return fmt.Errorf("winfocus: window focusing is only supported on Windows")
}
