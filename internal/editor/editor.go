// Package editor provides a post-capture annotation editor.
//
// The editor holds a captured screenshot and a stack of annotations
// (pen strokes, rectangles, arrows, text).  Annotations are rendered
// onto the image by Render, which returns a new image.Image that can
// be passed to clipboard.CopyImage or utils.SaveImage.
//
// Undo pops the most-recently added annotation from the stack.
//
// For the GUI layer this package is designed to be driven by either:
//   - Fyne (fyne.io/fyne/v2) for cross-platform support
//   - Walk (github.com/lxn/walk) for a native Windows look
//   - Raw Win32 for maximum control
//
// The rendering engine uses fogleman/gg backed by golang/freetype.
package editor

import (
	"fmt"
	"image"
	"math"
	"os"

	"github.com/fogleman/gg"

	"github.com/aung-arata/screensaver/internal/clipboard"
	"github.com/aung-arata/screensaver/internal/utils"
)

// Tool represents an annotation tool.
type Tool string

const (
	ToolPen   Tool = "pen"
	ToolRect  Tool = "rect"
	ToolArrow Tool = "arrow"
	ToolText  Tool = "text"
)

// Point is a 2-D coordinate used by annotation types.
type Point struct {
	X, Y float64
}

// Annotation is the interface implemented by all annotation types.
// Draw renders the annotation onto ctx.
type Annotation interface {
	Draw(ctx *gg.Context)
}

// PenStroke is a freehand polyline drawn with the pen tool.
type PenStroke struct {
	Points []Point
	Colour string  // hex colour, e.g. "#FF0000"
	Width  float64 // line width in pixels
}

// Draw implements Annotation.
func (p *PenStroke) Draw(ctx *gg.Context) {
	if len(p.Points) < 2 {
		return
	}
	ctx.SetHexColor(p.Colour)
	ctx.SetLineWidth(p.Width)
	ctx.MoveTo(p.Points[0].X, p.Points[0].Y)
	for _, pt := range p.Points[1:] {
		ctx.LineTo(pt.X, pt.Y)
	}
	ctx.Stroke()
}

// RectAnnotation draws a hollow rectangle.
type RectAnnotation struct {
	X1, Y1, X2, Y2 float64
	Colour          string
	Width           float64
}

// Draw implements Annotation.
func (r *RectAnnotation) Draw(ctx *gg.Context) {
	ctx.SetHexColor(r.Colour)
	ctx.SetLineWidth(r.Width)
	x := math.Min(r.X1, r.X2)
	y := math.Min(r.Y1, r.Y2)
	w := math.Abs(r.X2 - r.X1)
	h := math.Abs(r.Y2 - r.Y1)
	ctx.DrawRectangle(x, y, w, h)
	ctx.Stroke()
}

// ArrowAnnotation draws a line with an arrowhead at the endpoint.
type ArrowAnnotation struct {
	X1, Y1, X2, Y2 float64
	Colour          string
	Width           float64
}

// Draw implements Annotation.
func (a *ArrowAnnotation) Draw(ctx *gg.Context) {
	ctx.SetHexColor(a.Colour)
	ctx.SetLineWidth(a.Width)

	// Shaft.
	ctx.DrawLine(a.X1, a.Y1, a.X2, a.Y2)
	ctx.Stroke()

	// Arrowhead: two lines fanning out from (X2, Y2).
	angle := math.Atan2(a.Y2-a.Y1, a.X2-a.X1)
	headLen := math.Max(10, a.Width*4)
	const spread = math.Pi / 6 // 30°

	for _, side := range []float64{spread, -spread} {
		hx := a.X2 - headLen*math.Cos(angle-side)
		hy := a.Y2 - headLen*math.Sin(angle-side)
		ctx.DrawLine(a.X2, a.Y2, hx, hy)
		ctx.Stroke()
	}
}

// TextAnnotation renders a string at the given position.
type TextAnnotation struct {
	X, Y    float64
	Content string
	Colour  string
	Size    float64 // font size in points
}

// systemFontPaths lists candidate TrueType font paths to try, in priority
// order.  Only the first path that exists on the current system is used.
var systemFontPaths = []string{
	// Linux – DejaVu and Liberation families are widely available.
	"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
	"/usr/share/fonts/TTF/DejaVuSans.ttf",
	// macOS – San Francisco and Helvetica are bundled with the OS.
	"/System/Library/Fonts/SFNSDisplay.ttf",
	"/System/Library/Fonts/Helvetica.ttc",
	"/Library/Fonts/Arial.ttf",
	// Windows
	`C:\Windows\Fonts\arial.ttf`,
}

// findSystemFont returns the first existing path in systemFontPaths, or "".
func findSystemFont() string {
	for _, p := range systemFontPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// Draw implements Annotation.
// It attempts to load a system font at the requested size so that Size is
// honoured.  If no suitable font file is found on the current platform the
// call falls back to fogleman/gg's built-in bitmap font.
func (t *TextAnnotation) Draw(ctx *gg.Context) {
	ctx.SetHexColor(t.Colour)
	if t.Size > 0 {
		if path := findSystemFont(); path != "" {
			// Ignore the error: if loading fails we continue with the
			// bitmap fallback that is already active on the context.
			_ = ctx.LoadFontFace(path, t.Size)
		}
		// If no system font is found, fogleman/gg falls back to the
		// built-in basicfont.Face7x13 (fixed size).
	}
	ctx.DrawStringAnchored(t.Content, t.X, t.Y, 0, 1)
}

// Config holds editor configuration.
//
// Note: PenWidth is float64 (not int) to match fogleman/gg's SetLineWidth API.
type Config struct {
	PenColour string  // hex colour, e.g. "#FF0000"
	PenWidth  float64 // line width in pixels
	FontSize  float64 // font size for text annotations
}

// DefaultConfig returns sensible defaults for the editor.
func DefaultConfig() Config {
	return Config{
		PenColour: "#FF0000",
		PenWidth:  2,
		FontSize:  16,
	}
}

// Editor holds a captured image and a stack of annotations that can be
// rendered, undone, copied to the clipboard, or saved to disk.
type Editor struct {
	Image  image.Image
	Config Config
	// OnSave is an optional callback invoked with the saved file path after a
	// successful Save call from the annotation editor GUI.  May be nil.
	OnSave      func(path string)
	annotations []Annotation
}

// New creates a new Editor for the given image with default configuration.
func New(img image.Image) *Editor {
	return &Editor{
		Image:  img,
		Config: DefaultConfig(),
	}
}

// AddAnnotation appends a to the annotation stack.
func (e *Editor) AddAnnotation(a Annotation) {
	e.annotations = append(e.annotations, a)
}

// Undo removes the most recently added annotation.
// It is a no-op when the stack is empty.
func (e *Editor) Undo() {
	if len(e.annotations) == 0 {
		return
	}
	e.annotations = e.annotations[:len(e.annotations)-1]
}

// AnnotationCount returns the number of annotations in the stack.
func (e *Editor) AnnotationCount() int {
	return len(e.annotations)
}

// Render composites all annotations onto a copy of the base image and
// returns the result.  The base image is never mutated.
func (e *Editor) Render() (image.Image, error) {
	ctx := gg.NewContextForImage(e.Image)
	for _, a := range e.annotations {
		a.Draw(ctx)
	}
	return ctx.Image(), nil
}

// CopyToClipboard renders the annotated image and copies it to the
// system clipboard.
func (e *Editor) CopyToClipboard() error {
	img, err := e.Render()
	if err != nil {
		return fmt.Errorf("rendering annotations: %w", err)
	}
	return clipboard.CopyImage(img)
}

// Save renders the annotated image and writes it to path.
// The format is inferred from the file extension (.png / .jpg / .jpeg).
func (e *Editor) Save(path string) error {
	img, err := e.Render()
	if err != nil {
		return fmt.Errorf("rendering annotations: %w", err)
	}
	return utils.SaveImage(img, path)
}

// Run opens the annotation editor window.
//
// On Windows this displays a native Win32 window with a toolbar
// (Pen / Rect / Arrow / Text tools, colour picker, Undo, Copy, Save).
// On other platforms it falls back to saving the annotated image to a
// timestamped file.
func (e *Editor) Run() error {
	return e.runPlatform()
}
