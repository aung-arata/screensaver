// Package capture provides screen capture utilities.
//
// It wraps the kbinani/screenshot library to capture full-screen or
// rectangular regions across multiple monitors.
package capture

import (
	"fmt"
	"image"

	"github.com/kbinani/screenshot"
)

// Region represents a rectangular area on screen as (X, Y, Width, Height).
type Region struct {
	X, Y, Width, Height int
}

// FullScreen captures the full contents of the given monitor (0-based index)
// and returns an *image.RGBA.
//
// Pass -1 to capture the combined virtual screen (all monitors).
func FullScreen(monitor int) (*image.RGBA, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, fmt.Errorf("no active displays found")
	}

	if monitor == -1 {
		// Capture all monitors combined.
		bounds := image.Rect(0, 0, 0, 0)
		for i := 0; i < n; i++ {
			b := screenshot.GetDisplayBounds(i)
			bounds = bounds.Union(b)
		}
		return screenshot.Capture(bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy())
	}

	if monitor < 0 || monitor >= n {
		return nil, fmt.Errorf("monitor index %d out of range (0..%d)", monitor, n-1)
	}

	bounds := screenshot.GetDisplayBounds(monitor)
	return screenshot.Capture(bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy())
}

// CaptureRegion captures a rectangular region from the given monitor.
//
// The region coordinates are relative to the top-left corner of the monitor.
func CaptureRegion(region Region, monitor int) (*image.RGBA, error) {
	if region.Width <= 0 || region.Height <= 0 {
		return nil, fmt.Errorf("region width and height must be positive, got (%d, %d)", region.Width, region.Height)
	}

	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, fmt.Errorf("no active displays found")
	}
	if monitor < 0 || monitor >= n {
		return nil, fmt.Errorf("monitor index %d out of range (0..%d)", monitor, n-1)
	}

	bounds := screenshot.GetDisplayBounds(monitor)
	return screenshot.Capture(
		bounds.Min.X+region.X,
		bounds.Min.Y+region.Y,
		region.Width,
		region.Height,
	)
}

// MonitorInfo returns geometry information for a monitor (0-based index).
func MonitorInfo(monitor int) (image.Rectangle, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return image.Rectangle{}, fmt.Errorf("no active displays found")
	}
	if monitor < 0 || monitor >= n {
		return image.Rectangle{}, fmt.Errorf("monitor index %d out of range (0..%d)", monitor, n-1)
	}
	return screenshot.GetDisplayBounds(monitor), nil
}

// MonitorCount returns the number of active monitors.
func MonitorCount() int {
	return screenshot.NumActiveDisplays()
}

// NormalizeRegion converts two arbitrary corner points into a normalised
// Region with positive Width and Height. Returns nil if the selection
// has zero area.
func NormalizeRegion(x1, y1, x2, y2 int) *Region {
	left := min(x1, x2)
	top := min(y1, y2)
	width := abs(x2 - x1)
	height := abs(y2 - y1)

	if width == 0 || height == 0 {
		return nil
	}

	return &Region{X: left, Y: top, Width: width, Height: height}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
