//go:build !windows

package scrollcapture

import (
	"fmt"
	"image"
)

func capturePlatform(region image.Rectangle, cfg Config) (image.Image, error) {
	return nil, fmt.Errorf("long-page capture is only supported on Windows right now")
}
