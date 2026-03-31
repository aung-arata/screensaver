package hotkey

import (
	"fmt"
	"strings"
)

// Combo represents a parsed global hotkey key combination, split into
// modifier flags and a single printable key.
type Combo struct {
	Ctrl  bool
	Alt   bool
	Shift bool
	Key   string // single uppercase letter, e.g. "S"
}

// ParseCombo parses a human-readable combo string such as "ctrl+shift+s"
// into its constituent modifier flags and key.
//
// Recognised modifier names (case-insensitive): ctrl, control, alt, shift.
// The non-modifier part must be a single ASCII letter.
func ParseCombo(raw string) (Combo, error) {
	parts := strings.Split(strings.ToLower(raw), "+")
	if len(parts) == 0 {
		return Combo{}, fmt.Errorf("hotkey: empty combo")
	}

	var c Combo
	var key string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		switch p {
		case "ctrl", "control":
			c.Ctrl = true
		case "alt":
			c.Alt = true
		case "shift":
			c.Shift = true
		default:
			if key != "" {
				return Combo{}, fmt.Errorf("hotkey: multiple non-modifier keys in combo %q", raw)
			}
			key = p
		}
	}

	if key == "" || len(key) != 1 {
		return Combo{}, fmt.Errorf("hotkey: unsupported key %q in combo %q", key, raw)
	}

	c.Key = strings.ToUpper(key)
	return c, nil
}

// String returns a human-readable representation such as "Ctrl+Shift+S".
func (c Combo) String() string {
	var parts []string
	if c.Ctrl {
		parts = append(parts, "Ctrl")
	}
	if c.Alt {
		parts = append(parts, "Alt")
	}
	if c.Shift {
		parts = append(parts, "Shift")
	}
	parts = append(parts, c.Key)
	return strings.Join(parts, "+")
}
