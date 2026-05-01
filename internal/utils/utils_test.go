package utils

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
// DefaultSaveDirectory
// ---------------------------------------------------------------------------

func TestDefaultSaveDirectory_ReturnsPath(t *testing.T) {
	dir, err := DefaultSaveDirectory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty directory path")
	}
}

func TestDefaultSaveDirectory_CreatesDir(t *testing.T) {
	dir, err := DefaultSaveDirectory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory")
	}
}

// ---------------------------------------------------------------------------
// GenerateFilename
// ---------------------------------------------------------------------------

func TestGenerateFilename_PNGExtension(t *testing.T) {
	path, err := GenerateFilename(t.TempDir(), "png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(path, ".png") {
		t.Errorf("expected .png suffix, got %s", path)
	}
}

func TestGenerateFilename_JPGExtension(t *testing.T) {
	path, err := GenerateFilename(t.TempDir(), "jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(path, ".jpg") {
		t.Errorf("expected .jpg suffix, got %s", path)
	}
}

func TestGenerateFilename_ContainsScreenshot(t *testing.T) {
	path, err := GenerateFilename(t.TempDir(), "png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	base := filepath.Base(path)
	if !strings.Contains(base, "screenshot") {
		t.Errorf("expected filename to contain 'screenshot', got %s", base)
	}
}

func TestGenerateFilename_UsesProvidedDirectory(t *testing.T) {
	dir := t.TempDir()
	path, err := GenerateFilename(dir, "png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("expected parent dir %s, got %s", dir, filepath.Dir(path))
	}
}

// ---------------------------------------------------------------------------
// SaveImage
// ---------------------------------------------------------------------------

func TestSaveImage_ExplicitPath(t *testing.T) {
	img := makeTestImage(100, 80)
	dest := filepath.Join(t.TempDir(), "test.png")

	if err := SaveImage(img, dest, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(dest); err != nil {
		t.Errorf("file does not exist: %v", err)
	}
}

func TestSaveImage_JPEGFormat(t *testing.T) {
	img := makeTestImage(100, 80)
	dest := filepath.Join(t.TempDir(), "out.jpg")

	if err := SaveImage(img, dest, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(dest); err != nil {
		t.Errorf("file does not exist: %v", err)
	}
}

func TestSaveImage_AutoPath(t *testing.T) {
	img := makeTestImage(50, 50)

	// Use an empty path — should auto-generate.
	err := SaveImage(img, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
