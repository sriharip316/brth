package preset

import (
	"fmt"
	"regexp"
	"strings"
)

// Preset represents a defined breathing pattern cycle.
type Preset struct {
	Name        string  `yaml:"name" json:"name"`
	Description string  `yaml:"description" json:"description"`
	Inhale      float64 `yaml:"inhale" json:"inhale"`
	HoldIn      float64 `yaml:"hold_in" json:"hold_in"`
	Exhale      float64 `yaml:"exhale" json:"exhale"`
	HoldOut     float64 `yaml:"hold_out" json:"hold_out"`
	Cycles      int     `yaml:"cycles,omitempty" json:"cycles,omitempty"`
}

// BuiltinPresets returns the standard curated breathing presets.
func BuiltinPresets() []Preset {
	return []Preset{
		{
			Name:        "box",
			Description: "Box Breathing - Equal duration for focus, stress relief, and mental clarity",
			Inhale:      4.0,
			HoldIn:      4.0,
			Exhale:      4.0,
			HoldOut:     4.0,
			Cycles:      0,
		},
		{
			Name:        "4-7-8",
			Description: "4-7-8 Relaxing Breath - Calms nervous system, eases anxiety, promotes sleep",
			Inhale:      4.0,
			HoldIn:      7.0,
			Exhale:      8.0,
			HoldOut:     0.0,
			Cycles:      0,
		},
		{
			Name:        "coherent",
			Description: "Coherent / Resonant Breathing - Maximizes Heart Rate Variability (HRV) (~5.5 bpm)",
			Inhale:      5.5,
			HoldIn:      0.0,
			Exhale:      5.5,
			HoldOut:     0.0,
			Cycles:      0,
		},
		{
			Name:        "calm",
			Description: "Calm & Recover - Extended exhale to stimulate parasympathetic vagal tone",
			Inhale:      4.0,
			HoldIn:      2.0,
			Exhale:      6.0,
			HoldOut:     2.0,
			Cycles:      0,
		},
		{
			Name:        "awake",
			Description: "Awake & Energize - Quick exhalation to sharpen focus and boost energy",
			Inhale:      4.0,
			HoldIn:      0.0,
			Exhale:      2.0,
			HoldOut:     0.0,
			Cycles:      0,
		},
		{
			Name:        "triangle",
			Description: "Triangle Breathing - 3-phase pranayama for breath control and steady focus",
			Inhale:      4.0,
			HoldIn:      4.0,
			Exhale:      4.0,
			HoldOut:     0.0,
			Cycles:      0,
		},
		{
			Name:        "pranayama",
			Description: "1:4:2 Pranayama - Deep retention practice for lung capacity and calm",
			Inhale:      4.0,
			HoldIn:      16.0,
			Exhale:      8.0,
			HoldOut:     0.0,
			Cycles:      0,
		},
	}
}

// AllPresets returns combined built-in and custom presets.
// Custom presets override built-ins with the same name.
func AllPresets(customPresets []Preset) []Preset {
	builtins := BuiltinPresets()
	presetMap := make(map[string]Preset, len(builtins)+len(customPresets))
	order := make([]string, 0, len(builtins)+len(customPresets))

	for _, p := range builtins {
		key := strings.ToLower(p.Name)
		presetMap[key] = p
		order = append(order, key)
	}

	for _, p := range customPresets {
		key := strings.ToLower(p.Name)
		if _, exists := presetMap[key]; !exists {
			order = append(order, key)
		}
		presetMap[key] = p
	}

	result := make([]Preset, 0, len(order))
	for _, key := range order {
		result = append(result, presetMap[key])
	}
	return result
}

// GetPreset searches for a preset by name (case-insensitive) in custom and built-in presets.
func GetPreset(name string, customPresets []Preset) (Preset, bool) {
	needle := strings.ToLower(strings.TrimSpace(name))
	for _, p := range customPresets {
		if strings.ToLower(p.Name) == needle {
			return p, true
		}
	}
	for _, p := range BuiltinPresets() {
		if strings.ToLower(p.Name) == needle {
			return p, true
		}
	}
	return Preset{}, false
}

// TotalCycleDuration calculates the sum of all phase durations.
func (p Preset) TotalCycleDuration() float64 {
	return p.Inhale + p.HoldIn + p.Exhale + p.HoldOut
}

// presetNameRe constrains preset names to a safe identifier charset. Names
// come from user config files and are echoed to the terminal, so they must
// not be able to carry terminal escape sequences.
var presetNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _.\-]{0,31}$`)

// SanitizeControlChars strips terminal control runes (C0 and C1, including
// ESC) from config-supplied text so it cannot inject escape sequences when
// rendered.
func SanitizeControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || (r >= 0x7f && r <= 0x9f) {
			return -1
		}
		return r
	}, s)
}

// Validate verifies that the preset has sensible durations.
func (p Preset) Validate() error {
	if !presetNameRe.MatchString(p.Name) {
		return fmt.Errorf("preset name must be 1-32 characters of letters, digits, spaces, '.', '_', or '-'")
	}
	if p.Inhale <= 0 {
		return fmt.Errorf("inhale duration must be greater than 0")
	}
	if p.Exhale <= 0 {
		return fmt.Errorf("exhale duration must be greater than 0")
	}
	if p.HoldIn < 0 || p.HoldOut < 0 {
		return fmt.Errorf("hold durations cannot be negative")
	}
	return nil
}

// TimingPattern returns a concise representation like "4-4-4-4" or "4-7-8-0".
func (p Preset) TimingPattern() string {
	formatFloat := func(f float64) string {
		if f == float64(int(f)) {
			return fmt.Sprintf("%d", int(f))
		}
		return fmt.Sprintf("%.1f", f)
	}
	return fmt.Sprintf("%s-%s-%s-%s",
		formatFloat(p.Inhale),
		formatFloat(p.HoldIn),
		formatFloat(p.Exhale),
		formatFloat(p.HoldOut),
	)
}

// Summary returns a human-readable summary of the preset phases.
func (p Preset) Summary() string {
	parts := []string{fmt.Sprintf("%.1fs Inhale", p.Inhale)}
	if p.HoldIn > 0 {
		parts = append(parts, fmt.Sprintf("%.1fs Hold", p.HoldIn))
	}
	parts = append(parts, fmt.Sprintf("%.1fs Exhale", p.Exhale))
	if p.HoldOut > 0 {
		parts = append(parts, fmt.Sprintf("%.1fs Hold", p.HoldOut))
	}
	return fmt.Sprintf("%s (Total: %.1fs)", strings.Join(parts, " • "), p.TotalCycleDuration())
}
