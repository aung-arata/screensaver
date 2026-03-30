package overlay

import (
	"image"
	"image/color"
	"testing"
)

func makeRGBA(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = c.R
		img.Pix[i+1] = c.G
		img.Pix[i+2] = c.B
		img.Pix[i+3] = c.A
	}
	return img
}

// ---------------------------------------------------------------------------
// DimImage
// ---------------------------------------------------------------------------

func TestDimImage_FullBlack(t *testing.T) {
	src := makeRGBA(10, 10, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	dst := DimImage(src, 0)
	for i := 0; i < len(dst.Pix); i += 4 {
		if dst.Pix[i+0] != 0 || dst.Pix[i+1] != 0 || dst.Pix[i+2] != 0 {
			t.Fatal("expected all-black output with factor 0")
		}
		if dst.Pix[i+3] != 255 {
			t.Fatal("expected alpha preserved")
		}
	}
}

func TestDimImage_FullBright(t *testing.T) {
	src := makeRGBA(10, 10, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	dst := DimImage(src, 255)
	for i := 0; i < len(dst.Pix); i += 4 {
		if dst.Pix[i+0] != 200 || dst.Pix[i+1] != 100 || dst.Pix[i+2] != 50 {
			t.Fatalf("expected unchanged output with factor 255, got (%d,%d,%d)",
				dst.Pix[i+0], dst.Pix[i+1], dst.Pix[i+2])
		}
	}
}

func TestDimImage_Half(t *testing.T) {
	src := makeRGBA(4, 4, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	dst := DimImage(src, 128)
	// 200*128/255 ≈ 100, 100*128/255 ≈ 50, 50*128/255 ≈ 25
	for i := 0; i < len(dst.Pix); i += 4 {
		if dst.Pix[i+0] < 99 || dst.Pix[i+0] > 101 {
			t.Fatalf("R: got %d, expected ~100", dst.Pix[i+0])
		}
		if dst.Pix[i+1] < 49 || dst.Pix[i+1] > 51 {
			t.Fatalf("G: got %d, expected ~50", dst.Pix[i+1])
		}
		if dst.Pix[i+2] < 24 || dst.Pix[i+2] > 26 {
			t.Fatalf("B: got %d, expected ~25", dst.Pix[i+2])
		}
	}
}

func TestDimImage_PreservesAlpha(t *testing.T) {
	src := makeRGBA(4, 4, color.RGBA{R: 255, G: 255, B: 255, A: 128})
	dst := DimImage(src, 128)
	for i := 0; i < len(dst.Pix); i += 4 {
		if dst.Pix[i+3] != 128 {
			t.Fatalf("alpha: got %d, expected 128", dst.Pix[i+3])
		}
	}
}

func TestDimImage_DoesNotMutateSrc(t *testing.T) {
	src := makeRGBA(4, 4, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	_ = DimImage(src, 64)
	for i := 0; i < len(src.Pix); i += 4 {
		if src.Pix[i+0] != 200 || src.Pix[i+1] != 100 || src.Pix[i+2] != 50 {
			t.Fatal("DimImage mutated the source image")
		}
	}
}

// ---------------------------------------------------------------------------
// ComposeSelection
// ---------------------------------------------------------------------------

func TestComposeSelection_RestoresBrightPixels(t *testing.T) {
	src := makeRGBA(20, 20, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	dst := DimImage(src, 128) // dimmed copy

	sel := image.Rect(5, 5, 15, 15)
	ComposeSelection(dst, src, sel)

	// Inside the selection pixels should match src.
	for y := sel.Min.Y; y < sel.Max.Y; y++ {
		for x := sel.Min.X; x < sel.Max.X; x++ {
			off := dst.PixOffset(x, y)
			if dst.Pix[off+0] != 200 || dst.Pix[off+1] != 100 || dst.Pix[off+2] != 50 {
				t.Fatalf("pixel (%d,%d) inside selection not restored: (%d,%d,%d)",
					x, y, dst.Pix[off+0], dst.Pix[off+1], dst.Pix[off+2])
			}
		}
	}

	// Outside the selection pixels should still be dimmed.
	off := dst.PixOffset(0, 0)
	if dst.Pix[off+0] == 200 {
		t.Error("pixel outside selection should not be restored")
	}
}

func TestComposeSelection_EmptyRect(t *testing.T) {
	src := makeRGBA(10, 10, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	dst := DimImage(src, 128)
	origPix := make([]byte, len(dst.Pix))
	copy(origPix, dst.Pix)

	ComposeSelection(dst, src, image.ZR) // empty selection

	for i := range dst.Pix {
		if dst.Pix[i] != origPix[i] {
			t.Fatal("ComposeSelection with empty rect modified the image")
		}
	}
}

func TestComposeSelection_SelectionOutOfBounds(t *testing.T) {
	src := makeRGBA(10, 10, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	dst := DimImage(src, 128)

	// Selection extends beyond image bounds — should be clamped.
	sel := image.Rect(-5, -5, 15, 15)
	ComposeSelection(dst, src, sel)

	// Pixel at (0,0) is within the clamped selection.
	off := dst.PixOffset(0, 0)
	if dst.Pix[off+0] != 200 {
		t.Errorf("pixel (0,0) should be restored; got R=%d", dst.Pix[off+0])
	}
}
