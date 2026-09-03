package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sriharip316/brth/internal/breath"
)

// Theme defines colors for the application UI.
type Theme struct {
	Name      string
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	BgMuted   lipgloss.Color
	Text      lipgloss.Color
	TextMuted lipgloss.Color
	InhaleCol lipgloss.Color
	HoldCol   lipgloss.Color
	ExhaleCol lipgloss.Color
}

var Themes = map[string]Theme{
	"calm": {
		Name:      "calm",
		Primary:   lipgloss.Color("#00d2ff"),
		Secondary: lipgloss.Color("#3a7bd5"),
		Accent:    lipgloss.Color("#38bdf8"),
		BgMuted:   lipgloss.Color("#1e293b"),
		Text:      lipgloss.Color("#f0fdf4"),
		TextMuted: lipgloss.Color("#64748b"),
		InhaleCol: lipgloss.Color("#38bdf8"),
		HoldCol:   lipgloss.Color("#a855f7"),
		ExhaleCol: lipgloss.Color("#2dd4bf"),
	},
	"sunset": {
		Name:      "sunset",
		Primary:   lipgloss.Color("#ff758c"),
		Secondary: lipgloss.Color("#ff7eb3"),
		Accent:    lipgloss.Color("#f43f5e"),
		BgMuted:   lipgloss.Color("#271a2b"),
		Text:      lipgloss.Color("#fff1f2"),
		TextMuted: lipgloss.Color("#71717a"),
		InhaleCol: lipgloss.Color("#fb7185"),
		HoldCol:   lipgloss.Color("#e879f9"),
		ExhaleCol: lipgloss.Color("#fb923c"),
	},
	"forest": {
		Name:      "forest",
		Primary:   lipgloss.Color("#10b981"),
		Secondary: lipgloss.Color("#059669"),
		Accent:    lipgloss.Color("#34d399"),
		BgMuted:   lipgloss.Color("#064e3b"),
		Text:      lipgloss.Color("#ecfdf5"),
		TextMuted: lipgloss.Color("#6b7280"),
		InhaleCol: lipgloss.Color("#34d399"),
		HoldCol:   lipgloss.Color("#fbbf24"),
		ExhaleCol: lipgloss.Color("#22c55e"),
	},
	"nord": {
		Name:      "nord",
		Primary:   lipgloss.Color("#88c0d0"),
		Secondary: lipgloss.Color("#81a1c1"),
		Accent:    lipgloss.Color("#8fbcbb"),
		BgMuted:   lipgloss.Color("#2e3440"),
		Text:      lipgloss.Color("#eceff4"),
		TextMuted: lipgloss.Color("#4c566a"),
		InhaleCol: lipgloss.Color("#88c0d0"),
		HoldCol:   lipgloss.Color("#b48ead"),
		ExhaleCol: lipgloss.Color("#a3be8c"),
	},
	"ocean": {
		Name:      "ocean",
		Primary:   lipgloss.Color("#0284c7"),
		Secondary: lipgloss.Color("#0369a1"),
		Accent:    lipgloss.Color("#0ea5e9"),
		BgMuted:   lipgloss.Color("#0f172a"),
		Text:      lipgloss.Color("#f8fafc"),
		TextMuted: lipgloss.Color("#475569"),
		InhaleCol: lipgloss.Color("#38bdf8"),
		HoldCol:   lipgloss.Color("#818cf8"),
		ExhaleCol: lipgloss.Color("#06b6d4"),
	},
	"mono": {
		Name:      "mono",
		Primary:   lipgloss.Color("#f3f4f6"),
		Secondary: lipgloss.Color("#9ca3af"),
		Accent:    lipgloss.Color("#e5e7eb"),
		BgMuted:   lipgloss.Color("#1f2937"),
		Text:      lipgloss.Color("#ffffff"),
		TextMuted: lipgloss.Color("#6b7280"),
		InhaleCol: lipgloss.Color("#ffffff"),
		HoldCol:   lipgloss.Color("#d1d5db"),
		ExhaleCol: lipgloss.Color("#9ca3af"),
	},
}

// AvailableThemeNames returns an ordered list of available theme names.
var AvailableThemeNames = []string{"calm", "sunset", "forest", "nord", "ocean", "mono"}

// GetTheme returns the theme matching name, or falls back to "calm".
func GetTheme(name string) Theme {
	if t, exists := Themes[strings.ToLower(strings.TrimSpace(name))]; exists {
		return t
	}
	return Themes["calm"]
}

// NextTheme returns the next theme in cyclic order.
func NextTheme(currentName string) Theme {
	currentName = strings.ToLower(strings.TrimSpace(currentName))
	for i, name := range AvailableThemeNames {
		if name == currentName {
			nextIdx := (i + 1) % len(AvailableThemeNames)
			return Themes[AvailableThemeNames[nextIdx]]
		}
	}
	return Themes[AvailableThemeNames[0]]
}

// PhaseColor returns the accent color tailored to the current breathing phase.
func (t Theme) PhaseColor(phase breath.Phase) lipgloss.Color {
	switch phase {
	case breath.PhaseInhale:
		return t.InhaleCol
	case breath.PhaseHoldIn:
		return t.HoldCol
	case breath.PhaseExhale:
		return t.ExhaleCol
	case breath.PhaseHoldOut:
		return t.TextMuted
	case breath.PhasePrep:
		return t.Accent
	case breath.PhaseDone:
		return t.Primary
	default:
		return t.Text
	}
}
