package preset

import (
	"testing"
)

func TestBuiltinPresets(t *testing.T) {
	presets := BuiltinPresets()
	if len(presets) == 0 {
		t.Fatalf("expected builtin presets, got empty")
	}

	for _, p := range presets {
		if err := p.Validate(); err != nil {
			t.Errorf("builtin preset %s is invalid: %v", p.Name, err)
		}
		if p.TotalCycleDuration() <= 0 {
			t.Errorf("builtin preset %s has invalid total duration: %f", p.Name, p.TotalCycleDuration())
		}
		if p.TimingPattern() == "" {
			t.Errorf("builtin preset %s has empty timing pattern", p.Name)
		}
	}
}

func TestGetPreset(t *testing.T) {
	p, found := GetPreset("box", nil)
	if !found {
		t.Fatalf("expected to find 'box' preset")
	}
	if p.Inhale != 4.0 || p.HoldIn != 4.0 || p.Exhale != 4.0 || p.HoldOut != 4.0 {
		t.Errorf("unexpected box timings: %+v", p)
	}

	// Test case-insensitivity
	pUpper, foundUpper := GetPreset("BOX", nil)
	if !foundUpper || pUpper.Name != p.Name {
		t.Errorf("expected case-insensitive match for BOX")
	}

	// Test custom preset overriding built-in
	custom := []Preset{
		{
			Name:    "box",
			Inhale:  5.0,
			HoldIn:  5.0,
			Exhale:  5.0,
			HoldOut: 5.0,
		},
		{
			Name:    "zen-custom",
			Inhale:  3.0,
			HoldIn:  0.0,
			Exhale:  3.0,
			HoldOut: 0.0,
		},
	}

	pOverridden, foundOverridden := GetPreset("box", custom)
	if !foundOverridden || pOverridden.Inhale != 5.0 {
		t.Errorf("expected custom override for box to have Inhale=5.0, got %f", pOverridden.Inhale)
	}

	pZen, foundZen := GetPreset("zen-custom", custom)
	if !foundZen || pZen.Inhale != 3.0 {
		t.Errorf("expected to find custom preset zen-custom")
	}

	_, foundMissing := GetPreset("non-existent-xyz", custom)
	if foundMissing {
		t.Errorf("did not expect to find non-existent preset")
	}
}

func TestAllPresets(t *testing.T) {
	custom := []Preset{
		{
			Name:   "custom-test",
			Inhale: 3,
			Exhale: 3,
		},
	}
	all := AllPresets(custom)
	foundCustom := false
	for _, p := range all {
		if p.Name == "custom-test" {
			foundCustom = true
			break
		}
	}
	if !foundCustom {
		t.Errorf("expected AllPresets to include custom preset")
	}
}

func TestPresetValidation(t *testing.T) {
	invalidInhale := Preset{Name: "bad", Inhale: 0, Exhale: 4}
	if err := invalidInhale.Validate(); err == nil {
		t.Errorf("expected error for zero inhale")
	}

	invalidExhale := Preset{Name: "bad", Inhale: 4, Exhale: -1}
	if err := invalidExhale.Validate(); err == nil {
		t.Errorf("expected error for negative exhale")
	}

	invalidHold := Preset{Name: "bad", Inhale: 4, Exhale: 4, HoldIn: -2}
	if err := invalidHold.Validate(); err == nil {
		t.Errorf("expected error for negative hold")
	}

	invalidName := Preset{Name: "bad\x1b[2Jname", Inhale: 4, Exhale: 4}
	if err := invalidName.Validate(); err == nil {
		t.Errorf("expected error for name with control characters")
	}

	emptyName := Preset{Name: "", Inhale: 4, Exhale: 4}
	if err := emptyName.Validate(); err == nil {
		t.Errorf("expected error for empty preset name")
	}

	valid := Preset{Name: "good", Inhale: 4, Exhale: 4, HoldIn: 0, HoldOut: 0}
	if err := valid.Validate(); err != nil {
		t.Errorf("unexpected error for valid preset: %v", err)
	}
}

func TestSanitizeControlChars(t *testing.T) {
	// C0 controls (ESC, BEL), DEL, and C1 controls must all be stripped.
	in := "es\x1bcape\x07de\x7fscri\u009bption"
	want := "escapedescription"
	if got := SanitizeControlChars(in); got != want {
		t.Errorf("SanitizeControlChars(%q) = %q, want %q", in, got, want)
	}
	if got := SanitizeControlChars("clean 4-7-8 text"); got != "clean 4-7-8 text" {
		t.Errorf("SanitizeControlChars altered clean text: %q", got)
	}
}
