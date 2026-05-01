package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// DefaultConfig
// ---------------------------------------------------------------------------

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Hotkey != "ctrl+shift+s" {
		t.Errorf("expected hotkey 'ctrl+shift+s', got %q", cfg.Hotkey)
	}
	if cfg.SaveDir != "" {
		t.Errorf("expected empty SaveDir, got %q", cfg.SaveDir)
	}
	if cfg.Format != "png" {
		t.Errorf("expected format 'png', got %q", cfg.Format)
	}
	if cfg.Quality != 90 {
		t.Errorf("expected quality 90, got %d", cfg.Quality)
	}
}

// ---------------------------------------------------------------------------
// Merge
// ---------------------------------------------------------------------------

func TestMerge_OverrideWins(t *testing.T) {
	base := DefaultConfig()
	overrides := Config{
		Hotkey:  "ctrl+alt+s",
		SaveDir: "/tmp/shots",
		Format:  "jpeg",
		Quality: 75,
	}
	result := Merge(base, overrides)
	if result.Hotkey != "ctrl+alt+s" {
		t.Errorf("expected hotkey override, got %q", result.Hotkey)
	}
	if result.SaveDir != "/tmp/shots" {
		t.Errorf("expected SaveDir override, got %q", result.SaveDir)
	}
	if result.Format != "jpeg" {
		t.Errorf("expected format override, got %q", result.Format)
	}
	if result.Quality != 75 {
		t.Errorf("expected quality override, got %d", result.Quality)
	}
}

func TestMerge_ZeroValuesDoNotOverwrite(t *testing.T) {
	base := DefaultConfig()
	overrides := Config{} // all zero values
	result := Merge(base, overrides)
	if result.Hotkey != base.Hotkey {
		t.Errorf("expected base hotkey %q, got %q", base.Hotkey, result.Hotkey)
	}
	if result.Format != base.Format {
		t.Errorf("expected base format %q, got %q", base.Format, result.Format)
	}
	if result.Quality != base.Quality {
		t.Errorf("expected base quality %d, got %d", base.Quality, result.Quality)
	}
}

func TestMerge_PartialOverride(t *testing.T) {
	base := DefaultConfig()
	overrides := Config{Format: "jpeg"} // only format changed
	result := Merge(base, overrides)
	if result.Hotkey != base.Hotkey {
		t.Errorf("expected base hotkey %q, got %q", base.Hotkey, result.Hotkey)
	}
	if result.Format != "jpeg" {
		t.Errorf("expected jpeg format, got %q", result.Format)
	}
	if result.Quality != base.Quality {
		t.Errorf("expected base quality %d, got %d", base.Quality, result.Quality)
	}
}

// ---------------------------------------------------------------------------
// Load
// ---------------------------------------------------------------------------

func TestLoad_ReturnsDefaultsWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.yaml")
	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := DefaultConfig()
	if cfg.Hotkey != def.Hotkey || cfg.Format != def.Format || cfg.Quality != def.Quality {
		t.Errorf("expected defaults, got %+v", cfg)
	}
}

func TestLoad_ErrorOnMalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(":\tinvalid:\tyaml:::"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	_, err := LoadFrom(path)
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}
}

// ---------------------------------------------------------------------------
// Save + Load round-trip
// ---------------------------------------------------------------------------

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := Config{
		Hotkey:  "ctrl+f1",
		SaveDir: "/my/shots",
		Format:  "jpeg",
		Quality: 80,
	}

	if err := SaveTo(original, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", loaded, original)
	}
}

func TestSaveTo_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yaml")

	cfg := DefaultConfig()
	if err := SaveTo(cfg, path); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}
