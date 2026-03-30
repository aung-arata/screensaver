package editor

import (
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/fogleman/gg"
)

// makeTestImage returns a plain red image of size w×h.
func makeTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	return img
}

// ---------------------------------------------------------------------------
// DefaultConfig
// ---------------------------------------------------------------------------

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.PenColour == "" {
		t.Error("expected non-empty default PenColour")
	}
	if cfg.PenWidth <= 0 {
		t.Errorf("expected positive default PenWidth, got %v", cfg.PenWidth)
	}
	if cfg.FontSize <= 0 {
		t.Errorf("expected positive default FontSize, got %v", cfg.FontSize)
	}
}

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

func TestNew_SetsImage(t *testing.T) {
	img := makeTestImage(100, 80)
	e := New(img)
	if e.Image != img {
		t.Error("expected editor to hold the provided image")
	}
}

func TestNew_SetsDefaultConfig(t *testing.T) {
	e := New(makeTestImage(10, 10))
	if e.Config.PenColour == "" {
		t.Error("expected default PenColour to be set")
	}
}

func TestNew_EmptyAnnotations(t *testing.T) {
	e := New(makeTestImage(10, 10))
	if e.AnnotationCount() != 0 {
		t.Errorf("expected 0 annotations after New, got %d", e.AnnotationCount())
	}
}

// ---------------------------------------------------------------------------
// AddAnnotation / AnnotationCount
// ---------------------------------------------------------------------------

func TestAddAnnotation_IncrementsCount(t *testing.T) {
	e := New(makeTestImage(100, 100))
	e.AddAnnotation(&PenStroke{
		Points: []Point{{0, 0}, {10, 10}},
		Colour: "#0000FF",
		Width:  2,
	})
	if e.AnnotationCount() != 1 {
		t.Errorf("expected 1 annotation, got %d", e.AnnotationCount())
	}
}

func TestAddAnnotation_MultipleTypes(t *testing.T) {
	e := New(makeTestImage(100, 100))
	e.AddAnnotation(&PenStroke{Points: []Point{{0, 0}, {10, 10}}, Colour: "#FF0000", Width: 2})
	e.AddAnnotation(&RectAnnotation{X1: 5, Y1: 5, X2: 50, Y2: 50, Colour: "#00FF00", Width: 2})
	e.AddAnnotation(&ArrowAnnotation{X1: 0, Y1: 0, X2: 80, Y2: 80, Colour: "#0000FF", Width: 2})
	e.AddAnnotation(&TextAnnotation{X: 10, Y: 10, Content: "hello", Colour: "#FFFFFF", Size: 14})

	if e.AnnotationCount() != 4 {
		t.Errorf("expected 4 annotations, got %d", e.AnnotationCount())
	}
}

// ---------------------------------------------------------------------------
// Undo
// ---------------------------------------------------------------------------

func TestUndo_RemovesLastAnnotation(t *testing.T) {
	e := New(makeTestImage(100, 100))
	e.AddAnnotation(&PenStroke{Points: []Point{{0, 0}, {10, 10}}, Colour: "#FF0000", Width: 2})
	e.AddAnnotation(&RectAnnotation{X1: 5, Y1: 5, X2: 50, Y2: 50, Colour: "#00FF00", Width: 2})
	e.Undo()
	if e.AnnotationCount() != 1 {
		t.Errorf("expected 1 annotation after Undo, got %d", e.AnnotationCount())
	}
}

func TestUndo_EmptyStackIsNoop(t *testing.T) {
	e := New(makeTestImage(100, 100))
	e.Undo() // must not panic
	if e.AnnotationCount() != 0 {
		t.Errorf("expected 0 annotations, got %d", e.AnnotationCount())
	}
}

func TestUndo_AllAnnotations(t *testing.T) {
	e := New(makeTestImage(100, 100))
	for i := 0; i < 5; i++ {
		e.AddAnnotation(&PenStroke{
			Points: []Point{{0, 0}, {float64(i + 1), float64(i + 1)}},
			Colour: "#FF0000", Width: 1,
		})
	}
	for i := 0; i < 5; i++ {
		e.Undo()
	}
	if e.AnnotationCount() != 0 {
		t.Errorf("expected 0 annotations after undoing all, got %d", e.AnnotationCount())
	}
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func TestRender_ReturnsSameSizeImage(t *testing.T) {
	img := makeTestImage(200, 150)
	e := New(img)
	out, err := e.Render()
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	b := out.Bounds()
	if b.Dx() != 200 || b.Dy() != 150 {
		t.Errorf("expected 200×150, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestRender_NoAnnotationsPreservesPixels(t *testing.T) {
	img := makeTestImage(50, 50)
	e := New(img)
	out, err := e.Render()
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	// Sample the top-left pixel — should still be red.
	r, g, b, _ := out.At(0, 0).RGBA()
	if r>>8 < 200 || g>>8 > 10 || b>>8 > 10 {
		t.Errorf("expected red pixel at (0,0), got (%d,%d,%d)", r>>8, g>>8, b>>8)
	}
}

func TestRender_DoesNotMutateSource(t *testing.T) {
	img := makeTestImage(50, 50)
	e := New(img)
	e.AddAnnotation(&RectAnnotation{X1: 0, Y1: 0, X2: 20, Y2: 20, Colour: "#0000FF", Width: 3})
	_, err := e.Render()
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	// Original image corner must still be red.
	r, g, b, _ := img.At(0, 0).RGBA()
	if r>>8 < 200 || g>>8 > 10 || b>>8 > 10 {
		t.Errorf("Render mutated source image: r=%d g=%d b=%d", r>>8, g>>8, b>>8)
	}
}

func TestRender_WithPenStroke(t *testing.T) {
	img := makeTestImage(100, 100)
	e := New(img)
	e.AddAnnotation(&PenStroke{
		Points: []Point{{10, 50}, {90, 50}},
		Colour: "#0000FF",
		Width:  4,
	})
	out, err := e.Render()
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	// A pixel near the middle of the stroke should have blue.
	_, g, b, _ := out.At(50, 50).RGBA()
	if b>>8 < 100 || g>>8 > 100 {
		t.Errorf("expected blue stroke at (50,50), got g=%d b=%d", g>>8, b>>8)
	}
}

func TestRender_WithRect(t *testing.T) {
	img := makeTestImage(100, 100)
	e := New(img)
	e.AddAnnotation(&RectAnnotation{
		X1: 20, Y1: 20,
		X2: 80, Y2: 80,
		Colour: "#00FF00",
		Width:  3,
	})
	out, err := e.Render()
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	// A pixel on the top edge of the rectangle should be green-ish.
	_, g, b, _ := out.At(50, 20).RGBA()
	if g>>8 < 100 || b>>8 > 100 {
		t.Errorf("expected green rect edge at (50,20), got g=%d b=%d", g>>8, b>>8)
	}
}

func TestRender_WithArrow(t *testing.T) {
	img := makeTestImage(200, 200)
	e := New(img)
	e.AddAnnotation(&ArrowAnnotation{
		X1: 20, Y1: 20, X2: 150, Y2: 150,
		Colour: "#0000FF",
		Width:  2,
	})
	out, err := e.Render()
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if out == nil {
		t.Error("expected non-nil image")
	}
}

func TestRender_WithText(t *testing.T) {
	img := makeTestImage(200, 100)
	e := New(img)
	e.AddAnnotation(&TextAnnotation{
		X: 10, Y: 50,
		Content: "Test",
		Colour:  "#FFFFFF",
		Size:    14,
	})
	out, err := e.Render()
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if out == nil {
		t.Error("expected non-nil image")
	}
}

// ---------------------------------------------------------------------------
// PenStroke.Draw — direct unit tests
// ---------------------------------------------------------------------------

func TestPenStroke_Draw_SinglePoint_NoOp(t *testing.T) {
	ctx := gg.NewContext(100, 100)
	p := &PenStroke{
		Points: []Point{{50, 50}},
		Colour: "#FF0000",
		Width:  2,
	}
	// Should not panic for a single-point stroke.
	p.Draw(ctx)
}

func TestPenStroke_Draw_TwoPoints(t *testing.T) {
	ctx := gg.NewContext(100, 100)
	p := &PenStroke{
		Points: []Point{{0, 50}, {100, 50}},
		Colour: "#0000FF",
		Width:  4,
	}
	p.Draw(ctx)
	// The pixel in the middle should now be blue.
	_, _, b, _ := ctx.Image().At(50, 50).RGBA()
	if b>>8 < 100 {
		t.Errorf("expected blue stroke, got b=%d", b>>8)
	}
}

// ---------------------------------------------------------------------------
// RectAnnotation.Draw
// ---------------------------------------------------------------------------

func TestRectAnnotation_Draw_NormalOrientation(t *testing.T) {
	ctx := gg.NewContext(100, 100)
	ctx.SetRGB(0, 0, 0)
	ctx.Clear()
	r := &RectAnnotation{X1: 10, Y1: 10, X2: 90, Y2: 90, Colour: "#FF0000", Width: 2}
	r.Draw(ctx)
	// Top edge.
	rVal, _, _, _ := ctx.Image().At(50, 10).RGBA()
	if rVal>>8 < 100 {
		t.Errorf("expected red top edge, got r=%d", rVal>>8)
	}
}

func TestRectAnnotation_Draw_ReversedCorners(t *testing.T) {
	ctx := gg.NewContext(100, 100)
	ctx.SetRGB(0, 0, 0)
	ctx.Clear()
	// X2 < X1 and Y2 < Y1 — normalisation must handle this.
	r := &RectAnnotation{X1: 90, Y1: 90, X2: 10, Y2: 10, Colour: "#FF0000", Width: 2}
	r.Draw(ctx)
	rVal, _, _, _ := ctx.Image().At(50, 10).RGBA()
	if rVal>>8 < 100 {
		t.Errorf("expected red top edge with reversed corners, got r=%d", rVal>>8)
	}
}

// ---------------------------------------------------------------------------
// ArrowAnnotation.Draw
// ---------------------------------------------------------------------------

func TestArrowAnnotation_Draw_DoesNotPanic(t *testing.T) {
	ctx := gg.NewContext(200, 200)
	a := &ArrowAnnotation{X1: 10, Y1: 10, X2: 150, Y2: 150, Colour: "#FF0000", Width: 2}
	a.Draw(ctx)
}

func TestArrowAnnotation_Draw_Shaft(t *testing.T) {
	ctx := gg.NewContext(200, 200)
	ctx.SetRGB(0, 0, 0)
	ctx.Clear()
	// Horizontal arrow from (10,100) to (190,100).
	a := &ArrowAnnotation{X1: 10, Y1: 100, X2: 190, Y2: 100, Colour: "#0000FF", Width: 3}
	a.Draw(ctx)
	// Mid-shaft should be blue.
	_, _, b, _ := ctx.Image().At(100, 100).RGBA()
	if b>>8 < 100 {
		t.Errorf("expected blue shaft, got b=%d", b>>8)
	}
}

func TestArrowAnnotation_HeadLength_ScalesWithWidth(t *testing.T) {
	// For a horizontal arrow (angle=0) with Width=5, headLen = max(10, 5*4) = 20.
	// Barb endpoints from tip (180, 100):
	//   spread = ±π/6 (30°)
	//   hx = 180 - 20*cos(∓π/6) ≈ 163
	//   hy = 100 - 20*sin(∓π/6) = 90 or 110
	// Scan the two barb regions and confirm coloured pixels are present there.
	ctx := gg.NewContext(200, 200)
	ctx.SetRGB(0, 0, 0)
	ctx.Clear()
	a := &ArrowAnnotation{X1: 20, Y1: 100, X2: 180, Y2: 100, Colour: "#0000FF", Width: 5}
	a.Draw(ctx)
	img := ctx.Image()

	// Helper: returns true if any pixel in the scan window has b > 100.
	hasBlue := func(x0, y0, x1, y1 int) bool {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				_, _, b, _ := img.At(x, y).RGBA()
				if b>>8 > 100 {
					return true
				}
			}
		}
		return false
	}

	// Barb toward (163, 110) – scan a ±3 px window around the midpoint (171, 105).
	if !hasBlue(168, 102, 174, 108) {
		t.Error("expected arrowhead barb pixels near (171, 105) but found none")
	}
	// Barb toward (163, 90) – scan a ±3 px window around the midpoint (171, 95).
	if !hasBlue(168, 92, 174, 98) {
		t.Error("expected arrowhead barb pixels near (171, 95) but found none")
	}
}

// ---------------------------------------------------------------------------
// TextAnnotation.Draw
// ---------------------------------------------------------------------------

func TestTextAnnotation_Draw_DoesNotPanic(t *testing.T) {
	ctx := gg.NewContext(200, 100)
	ta := &TextAnnotation{X: 10, Y: 50, Content: "hello", Colour: "#FFFFFF", Size: 14}
	ta.Draw(ctx)
}

// ---------------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------------

func TestSave_WritesFile(t *testing.T) {
	img := makeTestImage(50, 50)
	e := New(img)
	dest := t.TempDir() + "/annotated.png"
	if err := e.Save(dest); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("file does not exist: %v", err)
	}
	if info.Size() == 0 {
		t.Error("saved file is empty")
	}
}

func TestSave_WithAnnotations(t *testing.T) {
	img := makeTestImage(100, 100)
	e := New(img)
	e.AddAnnotation(&RectAnnotation{X1: 10, Y1: 10, X2: 90, Y2: 90, Colour: "#00FF00", Width: 2})
	e.AddAnnotation(&TextAnnotation{X: 20, Y: 50, Content: "saved", Colour: "#FFFFFF", Size: 12})
	dest := t.TempDir() + "/with_annotations.png"
	if err := e.Save(dest); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("file does not exist: %v", err)
	}
}
