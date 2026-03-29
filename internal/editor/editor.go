// Package editor provides a post-capture annotation editor.
//
// The editor displays the captured screenshot in a window with a toolbar
// offering annotation tools (pen, rectangle, arrow, text) plus Copy and
// Save actions — mirroring the core Lightshot workflow.
//
// For the GUI this package is designed to use either:
//   - Fyne (fyne.io/fyne/v2) for cross-platform support
//   - Walk (github.com/lxn/walk) for native Windows look
//   - Raw Win32 for maximum control
//
// The current scaffold defines the public types and interface.
package editor

import "image"

// Tool represents an annotation tool.
type Tool string

const (
	ToolPen   Tool = "pen"
	ToolRect  Tool = "rect"
	ToolArrow Tool = "arrow"
	ToolText  Tool = "text"
)

// Config holds editor configuration.
type Config struct {
	PenColour string // hex colour, e.g. "#FF0000"
	PenWidth  int
}

// DefaultConfig returns sensible defaults for the editor.
func DefaultConfig() Config {
	return Config{
		PenColour: "#FF0000",
		PenWidth:  2,
	}
}

// Editor represents a post-capture annotation editor window.
type Editor struct {
	Image  image.Image
	Config Config
}

// New creates a new Editor for the given image.
func New(img image.Image) *Editor {
	return &Editor{
		Image:  img,
		Config: DefaultConfig(),
	}
}

// Run opens the editor window. This is a placeholder that will be
// implemented with a GUI toolkit.
func (e *Editor) Run() error {
	// TODO: implement GUI editor with annotation tools.
	return nil
}
