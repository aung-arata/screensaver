package hotkey

import "testing"

// ---------------------------------------------------------------------------
// ParseCombo
// ---------------------------------------------------------------------------

func TestParseCombo_CtrlShiftS(t *testing.T) {
	c, err := ParseCombo("ctrl+shift+s")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Ctrl || c.Alt || !c.Shift || c.Key != "S" {
		t.Errorf("got %+v, want Ctrl+Shift+S", c)
	}
}

func TestParseCombo_CtrlAltP(t *testing.T) {
	c, err := ParseCombo("ctrl+alt+p")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Ctrl || !c.Alt || c.Shift || c.Key != "P" {
		t.Errorf("got %+v, want Ctrl+Alt+P", c)
	}
}

func TestParseCombo_CaseInsensitive(t *testing.T) {
	c, err := ParseCombo("CTRL+SHIFT+S")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Ctrl || !c.Shift || c.Key != "S" {
		t.Errorf("got %+v, want Ctrl+Shift+S", c)
	}
}

func TestParseCombo_ControlAlias(t *testing.T) {
	c, err := ParseCombo("control+shift+s")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Ctrl {
		t.Error("expected Ctrl to be set for 'control' alias")
	}
}

func TestParseCombo_SingleKey(t *testing.T) {
	c, err := ParseCombo("s")
	if err != nil {
		t.Fatalf("unexpected error for single key: %v", err)
	}
	if c.Key != "S" {
		t.Errorf("got key %q, want %q", c.Key, "S")
	}
}

func TestParseCombo_EmptyString(t *testing.T) {
	_, err := ParseCombo("")
	if err == nil {
		t.Error("expected error for empty combo")
	}
}

func TestParseCombo_NoKey(t *testing.T) {
	_, err := ParseCombo("ctrl+shift")
	if err == nil {
		t.Error("expected error for combo without key")
	}
}

func TestParseCombo_MultiCharKey(t *testing.T) {
	_, err := ParseCombo("ctrl+shift+ab")
	if err == nil {
		t.Error("expected error for multi-character key")
	}
}

func TestParseCombo_MultipleKeys(t *testing.T) {
	_, err := ParseCombo("ctrl+a+b")
	if err == nil {
		t.Error("expected error for multiple non-modifier keys")
	}
}

func TestParseCombo_WhitespaceAroundParts(t *testing.T) {
	c, err := ParseCombo("ctrl + shift + s")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Ctrl || !c.Shift || c.Key != "S" {
		t.Errorf("got %+v, want Ctrl+Shift+S", c)
	}
}

// ---------------------------------------------------------------------------
// Combo.String
// ---------------------------------------------------------------------------

func TestCombo_String(t *testing.T) {
	tests := []struct {
		combo Combo
		want  string
	}{
		{Combo{Ctrl: true, Shift: true, Key: "S"}, "Ctrl+Shift+S"},
		{Combo{Ctrl: true, Alt: true, Key: "P"}, "Ctrl+Alt+P"},
		{Combo{Alt: true, Key: "A"}, "Alt+A"},
		{Combo{Key: "X"}, "X"},
		{Combo{Ctrl: true, Alt: true, Shift: true, Key: "Q"}, "Ctrl+Alt+Shift+Q"},
	}
	for _, tt := range tests {
		if got := tt.combo.String(); got != tt.want {
			t.Errorf("Combo%+v.String() = %q, want %q", tt.combo, got, tt.want)
		}
	}
}
