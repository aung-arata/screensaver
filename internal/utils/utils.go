// Package utils provides file-save and path utilities for screenshots.
package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultSaveDirectory returns a platform-appropriate default directory
// for saved screenshots (~/Pictures/Screenshots). The directory is
// created if it does not exist.
func DefaultSaveDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}

	dir := filepath.Join(home, "Pictures", "Screenshots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating screenshot directory: %w", err)
	}
	return dir, nil
}

// GenerateFilename returns a timestamped filename inside the given directory.
// If dir is empty, DefaultSaveDirectory is used.
// The format should be "png" or "jpeg"; any other value defaults to "png".
func GenerateFilename(dir, format string) (string, error) {
	if dir == "" {
		d, err := DefaultSaveDirectory()
		if err != nil {
			return "", err
		}
		dir = d
	}

	if format != "png" && format != "jpeg" {
		format = "png"
	}

	ext := format
	if ext == "jpeg" {
		ext = "jpg"
	}

	timestamp := time.Now().Format("20060102_150405")
	name := fmt.Sprintf("screenshot_%s.%s", timestamp, ext)
	return filepath.Join(dir, name), nil
}

// SaveImage saves img to path using the given JPEG quality (1–100).
// Quality is ignored for PNG. Pass 0 to use the default quality (90).
// If path is empty, an auto-generated path is used. The format is inferred
// from the file extension (.png or .jpg/.jpeg).
func SaveImage(img image.Image, path string, quality int) error {
	if path == "" {
		p, err := GenerateFilename("", "png")
		if err != nil {
			return err
		}
		path = p
	}

	if quality <= 0 {
		quality = 90
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
	default:
		return png.Encode(f, img)
	}
}
