// Package history provides a persistent JSON index of captured screenshots.
//
// The index is stored at:
//   - %APPDATA%\screensaver\history.json  (Windows, or any platform when APPDATA is set)
//   - ~/.screensaver/history.json         (fallback)
//
// No build tags — compiles on all platforms.
package history

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"time"
)

const maxEntries = 200

// Entry represents a single captured screenshot in the history index.
type Entry struct {
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Format    string    `json:"format"`    // "png" or "jpeg"
	SizeBytes int64     `json:"size_bytes"`
}

// History holds the full capture history index.
type History struct {
	Entries []Entry `json:"entries"`
}

// IndexPath returns the path to the history JSON file.
//
// If the APPDATA environment variable is set (on any platform), it is used.
// This allows test environments on non-Windows platforms to control the path.
//
// On Windows:  %APPDATA%\screensaver\history.json
// Fallback:    ~/.screensaver/history.json
func IndexPath() (string, error) {
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		return filepath.Join(appdata, "screensaver", "history.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, ".screensaver", "history.json"), nil
}

// Load reads the history index from disk.
// Returns an empty History (not an error) if the file does not exist.
func Load() (History, error) {
	path, err := IndexPath()
	if err != nil {
		return History{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return History{Entries: []Entry{}}, nil
		}
		return History{}, fmt.Errorf("reading history file: %w", err)
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return History{}, fmt.Errorf("parsing history file: %w", err)
	}
	if h.Entries == nil {
		h.Entries = []Entry{}
	}
	return h, nil
}

// Save writes the history index to disk, creating parent directories if needed.
func Save(h History) error {
	path, err := IndexPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating history directory: %w", err)
	}
	if h.Entries == nil {
		h.Entries = []Entry{}
	}
	data, err := json.Marshal(h)
	if err != nil {
		return fmt.Errorf("encoding history: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing history file: %w", err)
	}
	return nil
}

// Add appends a new entry to the history index, trims to maxEntries (200),
// and saves to disk. It fills Width, Height, and SizeBytes from the file on disk.
func Add(path string, format string) error {
	h, err := Load()
	if err != nil {
		return err
	}

	entry := Entry{
		Path:      path,
		Timestamp: time.Now(),
		Format:    format,
	}

	// Fill file size.
	if fi, statErr := os.Stat(path); statErr == nil {
		entry.SizeBytes = fi.Size()
	}

	// Fill image dimensions using image.DecodeConfig.
	if f, openErr := os.Open(path); openErr == nil {
		if cfg, _, decErr := image.DecodeConfig(f); decErr == nil {
			entry.Width = cfg.Width
			entry.Height = cfg.Height
		}
		f.Close()
	}

	h.Entries = append(h.Entries, entry)

	// Trim to maxEntries, keeping the newest entries (tail).
	if len(h.Entries) > maxEntries {
		h.Entries = h.Entries[len(h.Entries)-maxEntries:]
	}

	return Save(h)
}

// Recent returns the n most recent entries, newest first.
// If n <= 0, all entries are returned.
func Recent(n int) ([]Entry, error) {
	h, err := Load()
	if err != nil {
		return nil, err
	}

	// Entries are stored oldest-first; reverse for newest-first output.
	entries := h.Entries
	reversed := make([]Entry, len(entries))
	for i, e := range entries {
		reversed[len(entries)-1-i] = e
	}

	if n > 0 && n < len(reversed) {
		return reversed[:n], nil
	}
	return reversed, nil
}

// Clear removes all entries from the history index.
// The file is kept on disk but written with an empty entries array.
func Clear() error {
	return Save(History{Entries: []Entry{}})
}
