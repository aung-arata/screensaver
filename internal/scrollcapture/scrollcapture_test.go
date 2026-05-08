package scrollcapture

import (
	"image"
	"testing"
)

func TestFindBestVerticalOverlap(t *testing.T) {
	prev := makeGradientFrame(120, 220, 0)
	curr := makeGradientFrame(120, 220, 140)

	got, ok := findBestVerticalOverlap(prev, curr)
	if !ok {
		t.Fatal("expected overlap detection to succeed")
	}
	if got != 80 {
		t.Fatalf("expected overlap=80, got %d", got)
	}
}

func TestFindBestVerticalOverlap_UnrelatedFrames(t *testing.T) {
	prev := makeSolidFrame(120, 220, 5, 15, 25)
	curr := makeSolidFrame(120, 220, 240, 230, 220)

	if _, ok := findBestVerticalOverlap(prev, curr); ok {
		t.Fatal("expected overlap detection to fail for unrelated frames")
	}
}

func TestFramesAreNearIdentical(t *testing.T) {
	a := makeGradientFrame(100, 140, 10)
	b := cloneRGBA(a)

	if !framesAreNearIdentical(a, b) {
		t.Fatal("expected identical frames to be considered near-identical")
	}
}

func TestCropRowsAndStitchFrames(t *testing.T) {
	first := makeGradientFrame(80, 120, 0)
	second := makeGradientFrame(80, 120, 70)
	part := cropRows(second, 50)

	got := stitchFrames([]*image.RGBA{first, part})
	if got.Bounds().Dx() != 80 {
		t.Fatalf("expected width=80, got %d", got.Bounds().Dx())
	}
	if got.Bounds().Dy() != 190 {
		t.Fatalf("expected height=190, got %d", got.Bounds().Dy())
	}
}

func makeGradientFrame(w, h, globalStartY int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		globalY := globalStartY + y
		v := uint8(globalY % 256)
		for x := 0; x < w; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i+0] = v
			img.Pix[i+1] = uint8((int(v) + x) % 256)
			img.Pix[i+2] = uint8((globalY * 3) % 256)
			img.Pix[i+3] = 255
		}
	}
	return img
}

func makeSolidFrame(w, h int, r, g, b byte) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i+0] = r
			img.Pix[i+1] = g
			img.Pix[i+2] = b
			img.Pix[i+3] = 255
		}
	}
	return img
}
