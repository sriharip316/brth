package ui

import (
	"strings"
	"testing"
)

func TestRenderCircleStyles(t *testing.T) {
	styles := []string{"braille-ring", "braille-fill", "block", "ascii"}
	theme := GetTheme("calm")

	for _, style := range styles {
		t.Run(style, func(t *testing.T) {
			cfg := CircleConfig{
				Style:       style,
				AspectRatio: 2.0,
				Theme:       theme,
				PhaseLabel:  "INHALE",
				TimeLabel:   "3.2s",
			}

			out := RenderCircle(40, 16, 0.5, cfg)
			if out == "" {
				t.Fatalf("expected non-empty rendered circle for style %s", style)
			}

			// Ensure phase and time label are present
			if !strings.Contains(out, "INHALE") {
				t.Errorf("rendered output missing phase label for style %s", style)
			}
			if !strings.Contains(out, "3.2s") {
				t.Errorf("rendered output missing time label for style %s", style)
			}
		})
	}
}

func TestRenderCircleEdgeDimensions(t *testing.T) {
	cfg := CircleConfig{
		Style:       "braille-ring",
		AspectRatio: 2.0,
		Theme:       GetTheme("calm"),
	}

	// Should not panic on tiny or zero dimensions
	_ = RenderCircle(5, 5, 0.0, cfg)
	_ = RenderCircle(100, 50, 1.0, cfg)
	_ = RenderCircle(30, 10, -0.5, cfg)
	_ = RenderCircle(30, 10, 1.5, cfg)
}

func TestNextThemeAndStyle(t *testing.T) {
	nextTheme := NextTheme("calm")
	if nextTheme.Name == "calm" {
		t.Errorf("expected NextTheme to advance from calm")
	}

	nextStyle := NextStyle("braille-ring")
	if nextStyle == "braille-ring" {
		t.Errorf("expected NextStyle to advance from braille-ring")
	}
}
