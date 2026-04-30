// Package config loads, saves, and merges user configuration for screensaver.
//
// The config file is located at:
//   - Windows:  %APPDATA%\screensaver\config.yaml
//   - Fallback: ~/.screensaver/config.yaml
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Config holds all user-configurable settings for screensaver.
type Config struct {
	Hotkey  string `yaml:"hotkey"`   // e.g. "ctrl+shift+s"
	SaveDir string `yaml:"save_dir"` // e.g. `C:\Users\user\Pictures\Screenshots`
	Format  string `yaml:"format"`   // "png" or "jpeg"  (default: "png")
	Quality int    `yaml:"quality"`  // JPEG quality 1-100 (default: 90, ignored for PNG)
}

// DefaultConfig returns built-in defaults.
func DefaultConfig() Config {
	return Config{
		Hotkey:  "ctrl+shift+s",
		SaveDir: "",
		Format:  "png",
		Quality: 90,
	}
}

// ConfigPath returns the platform-appropriate path to the config file.
// On Windows: %APPDATA%\screensaver\config.yaml
// Fallback:   ~/.screensaver/config.yaml
func ConfigPath() (string, error) {
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata != "" {
			return filepath.Join(appdata, "screensaver", "config.yaml"), nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}
	return filepath.Join(home, ".screensaver", "config.yaml"), nil
}

// Load reads the config file from ConfigPath.
// If the file does not exist, returns DefaultConfig() with no error.
// If the file exists but is malformed, returns an error.
func Load() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), fmt.Errorf("resolving config path: %w", err)
	}
	return LoadFrom(path)
}

// LoadFrom reads the config file from the given path.
// If the file does not exist, returns DefaultConfig() with no error.
// If the file exists but is malformed, returns an error.
func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), fmt.Errorf("reading config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), fmt.Errorf("parsing config file: %w", err)
	}
	return cfg, nil
}

// Save writes cfg to ConfigPath, creating the directory if needed.
func Save(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	return SaveTo(cfg, path)
}

// SaveTo writes cfg to the given path, creating the directory if needed.
func SaveTo(cfg Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}

// Merge applies non-zero CLI override values onto a base Config.
// A string override is applied only when it is non-empty ("").
// An int override is applied only when it is non-zero (0).
func Merge(base Config, overrides Config) Config {
	result := base
	if overrides.Hotkey != "" {
		result.Hotkey = overrides.Hotkey
	}
	if overrides.SaveDir != "" {
		result.SaveDir = overrides.SaveDir
	}
	if overrides.Format != "" {
		result.Format = overrides.Format
	}
	if overrides.Quality != 0 {
		result.Quality = overrides.Quality
	}
	return result
}

// EncodeYAML encodes cfg as YAML bytes.
func EncodeYAML(cfg Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
