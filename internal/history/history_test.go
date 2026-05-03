package history

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// createTestPNG writes a minimal PNG image of size w×h to path and returns path.
func createTestPNG(t *testing.T, path string, w, h int) string {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test png: %v", err)
	}
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	return path
}

// useTestDir redirects all history I/O to dir by setting APPDATA.
// The original value is restored via t.Cleanup.
func useTestDir(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("APPDATA", dir)
}

// ---------------------------------------------------------------------------
// Load — missing file returns empty history
// ---------------------------------------------------------------------------

func TestLoad_MissingFile(t *testing.T) {
	useTestDir(t, t.TempDir())

	h, err := Load()
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if len(h.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(h.Entries))
	}
}

// ---------------------------------------------------------------------------
// Add → Recent round-trip
// ---------------------------------------------------------------------------

func TestAdd_Recent_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	useTestDir(t, dir)

	imgDir := t.TempDir()
	paths := []string{
		createTestPNG(t, filepath.Join(imgDir, "a.png"), 100, 50),
		createTestPNG(t, filepath.Join(imgDir, "b.png"), 200, 100),
		createTestPNG(t, filepath.Join(imgDir, "c.png"), 300, 150),
	}

	for _, p := range paths {
		if err := Add(p, "png"); err != nil {
			t.Fatalf("Add(%q): %v", p, err)
		}
	}

	entries, err := Recent(2)
	if err != nil {
		t.Fatalf("Recent(2): %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Newest-first: last added path should be first.
	if entries[0].Path != paths[2] {
		t.Errorf("entries[0].Path = %q, want %q", entries[0].Path, paths[2])
	}
	if entries[1].Path != paths[1] {
		t.Errorf("entries[1].Path = %q, want %q", entries[1].Path, paths[1])
	}
	// Width/Height should be populated.
	if entries[0].Width != 300 || entries[0].Height != 150 {
		t.Errorf("entries[0] dims = %d×%d, want 300×150", entries[0].Width, entries[0].Height)
	}
	// SizeBytes should be non-zero.
	if entries[0].SizeBytes == 0 {
		t.Errorf("entries[0].SizeBytes = 0, want > 0")
	}
	// Format should be recorded.
	if entries[0].Format != "png" {
		t.Errorf("entries[0].Format = %q, want %q", entries[0].Format, "png")
	}
}

// ---------------------------------------------------------------------------
// Trim to 200 entries
// ---------------------------------------------------------------------------

func TestAdd_TrimTo200(t *testing.T) {
	dir := t.TempDir()
	useTestDir(t, dir)

	imgDir := t.TempDir()
	// Use a single image file for all 201 adds (path can repeat — we just test trimming).
	imgPath := createTestPNG(t, filepath.Join(imgDir, "x.png"), 10, 10)

	for i := 0; i < 201; i++ {
		if err := Add(imgPath, "png"); err != nil {
			t.Fatalf("Add iteration %d: %v", i, err)
		}
	}

	h, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(h.Entries) != maxEntries {
		t.Errorf("expected %d entries after trim, got %d", maxEntries, len(h.Entries))
	}
}

// ---------------------------------------------------------------------------
// Clear empties the index
// ---------------------------------------------------------------------------

func TestClear(t *testing.T) {
	dir := t.TempDir()
	useTestDir(t, dir)

	imgDir := t.TempDir()
	imgPath := createTestPNG(t, filepath.Join(imgDir, "x.png"), 10, 10)

	if err := Add(imgPath, "png"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if err := Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	h, err := Load()
	if err != nil {
		t.Fatalf("Load after Clear: %v", err)
	}
	if len(h.Entries) != 0 {
		t.Errorf("expected 0 entries after Clear, got %d", len(h.Entries))
	}

	// File should still exist.
	path, _ := IndexPath()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("history file should exist after Clear: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Recent(0) returns all entries
// ---------------------------------------------------------------------------

func TestRecent_ZeroReturnsAll(t *testing.T) {
	dir := t.TempDir()
	useTestDir(t, dir)

	imgDir := t.TempDir()
	imgPath := createTestPNG(t, filepath.Join(imgDir, "x.png"), 10, 10)

	for i := 0; i < 5; i++ {
		if err := Add(imgPath, "png"); err != nil {
			t.Fatalf("Add iteration %d: %v", i, err)
		}
	}

	entries, err := Recent(0)
	if err != nil {
		t.Fatalf("Recent(0): %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("expected 5 entries from Recent(0), got %d", len(entries))
	}
}

// ---------------------------------------------------------------------------
// Recent(n) on empty history returns empty slice (no error)
// ---------------------------------------------------------------------------

func TestRecent_EmptyHistory(t *testing.T) {
	useTestDir(t, t.TempDir())

	entries, err := Recent(10)
	if err != nil {
		t.Fatalf("Recent on empty history: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
