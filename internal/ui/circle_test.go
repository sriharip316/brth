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

func TestCircleContainedAndRound(t *testing.T) {
	cfg := CircleConfig{
		Style:       "braille-ring",
		AspectRatio: 2.0,
		Theme:       GetTheme("calm"),
	}

	for _, factor := range []float64{0.0, 0.5, 1.0} {
		out := RenderCircle(36, 14, factor, cfg)
		lines := strings.Split(out, "\n")
		if len(lines) != 14 {
			t.Errorf("expected 14 lines, got %d", len(lines))
		}

		// At factor 1.0 (max expansion), ensure top and bottom arcs exist
		// (i.e. not clipped off into vertical brackets)
		if factor == 1.0 {
			// Row 1 or 2 should have characters for the top arc
			topFound := false
			for r := 0; r <= 2; r++ {
				if strings.TrimSpace(lines[r]) != "" {
					topFound = true
					break
				}
			}
			if !topFound {
				t.Errorf("expected top arc within top rows for factor 1.0")
			}

			// Row 11, 12, or 13 should have characters for the bottom arc
			bottomFound := false
			for r := 11; r < 14; r++ {
				if strings.TrimSpace(lines[r]) != "" {
					bottomFound = true
					break
				}
			}
			if !bottomFound {
				t.Errorf("expected bottom arc within bottom rows for factor 1.0")
			}
		}
	}
}
