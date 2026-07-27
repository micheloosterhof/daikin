// ABOUTME: Tests for translating between human-friendly mode/fan/swing names
// ABOUTME: and the adapter's numeric codes.
package daikin

import "testing"

func TestModeCode(t *testing.T) {
	cases := map[string]string{"auto": "0", "dry": "2", "cool": "3", "heat": "4", "fan": "6"}
	for name, code := range cases {
		got, err := ModeCode(name)
		if err != nil || got != code {
			t.Errorf("ModeCode(%q) = %q, %v; want %q", name, got, err, code)
		}
	}
	if _, err := ModeCode("turbo"); err == nil {
		t.Error("ModeCode(turbo) should error")
	}
}

func TestModeName(t *testing.T) {
	cases := map[string]string{"0": "auto", "1": "auto", "7": "auto", "2": "dry", "3": "cool", "4": "heat", "6": "fan"}
	for code, name := range cases {
		if got := ModeName(code); got != name {
			t.Errorf("ModeName(%q) = %q, want %q", code, got, name)
		}
	}
}

func TestFanRateCode(t *testing.T) {
	cases := map[string]string{"auto": "A", "quiet": "B", "1": "3", "3": "5", "5": "7"}
	for name, code := range cases {
		got, err := FanRateCode(name)
		if err != nil || got != code {
			t.Errorf("FanRateCode(%q) = %q, %v; want %q", name, got, err, code)
		}
	}
	if _, err := FanRateCode("11"); err == nil {
		t.Error("FanRateCode(11) should error")
	}
}

func TestFanRateName(t *testing.T) {
	cases := map[string]string{"A": "auto", "B": "quiet", "3": "1", "7": "5"}
	for code, name := range cases {
		if got := FanRateName(code); got != name {
			t.Errorf("FanRateName(%q) = %q, want %q", code, got, name)
		}
	}
}

func TestSwingCode(t *testing.T) {
	cases := map[string]string{"off": "0", "vertical": "1", "horizontal": "2", "both": "3"}
	for name, code := range cases {
		got, err := SwingCode(name)
		if err != nil || got != code {
			t.Errorf("SwingCode(%q) = %q, %v; want %q", name, got, err, code)
		}
	}
	if _, err := SwingCode("diagonal"); err == nil {
		t.Error("SwingCode(diagonal) should error")
	}
}

func TestSwingName(t *testing.T) {
	cases := map[string]string{"0": "off", "1": "vertical", "2": "horizontal", "3": "both"}
	for code, name := range cases {
		if got := SwingName(code); got != name {
			t.Errorf("SwingName(%q) = %q, want %q", code, got, name)
		}
	}
}
