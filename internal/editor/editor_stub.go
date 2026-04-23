//go:build !windows

package editor

import (
	"fmt"

	"github.com/aung-arata/screensaver/internal/utils"
)

// runPlatform is the non-Windows fallback: it renders all annotations and
// saves the result to a timestamped PNG file.
func (e *Editor) runPlatform() error {
	path, err := utils.GenerateFilename("", "png")
	if err != nil {
		return fmt.Errorf("generating filename: %w", err)
	}
	if err := e.Save(path); err != nil {
		return fmt.Errorf("saving annotated screenshot: %w", err)
	}
	fmt.Printf("[editor] Annotated screenshot saved to %s\n", path)
	return nil
}
