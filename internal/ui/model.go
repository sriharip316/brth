package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sriharip316/brth/internal/breath"
	"github.com/sriharip316/brth/internal/config"
	"github.com/sriharip316/brth/internal/preset"
)

type tickMsg time.Time

func tickCmd(fps int) tea.Cmd {
	if fps <= 0 {
		fps = 30
	}
	interval := time.Second / time.Duration(fps)
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// bellCmd rings the terminal bell. It is returned as a tea.Cmd so the write
// is sequenced by the event loop instead of racing the renderer goroutine.
func bellCmd() tea.Msg {
	fmt.Print("\a")
	return nil
}

// Model represents the Bubble Tea UI state.
type Model struct {
	Engine         *breath.Engine
	Presets        []preset.Preset
	PresetIndex    int
	Config         *config.Config
	Theme          Theme
	Style          string
	Width          int
	Height         int
	ShowHelp       bool
	SoundEnabled   bool
	Quitting       bool
	CyclesOverride int
}

// NewModel initializes the Bubble Tea model with configuration and initial preset.
func NewModel(cfg *config.Config, initialPreset preset.Preset, allPresets []preset.Preset, cyclesOverride int) Model {
	initialIndex := 0
	for i, p := range allPresets {
		if strings.EqualFold(p.Name, initialPreset.Name) {
			initialIndex = i
			break
		}
	}

	theme := GetTheme(cfg.Theme)
	engine := breath.NewEngine(initialPreset, cfg.PrepSeconds, cyclesOverride)

	return Model{
		Engine:         engine,
		Presets:        allPresets,
		PresetIndex:    initialIndex,
		Config:         cfg,
		Theme:          theme,
		Style:          cfg.Style,
		SoundEnabled:   cfg.Sound,
		CyclesOverride: cyclesOverride,
	}
}

// Init begins the animation tick loop.
func (m Model) Init() tea.Cmd {
	return tickCmd(m.Config.FPS)
}

// Update handles keyboard and timer events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tickMsg:
		if m.Quitting {
			return m, nil
		}
		transitioned := m.Engine.Update(time.Time(msg))
		if transitioned && m.SoundEnabled {
			return m, tea.Batch(tickCmd(m.Config.FPS), bellCmd)
		}
		return m, tickCmd(m.Config.FPS)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// The help modal documents Esc as its dismiss key.
			if m.ShowHelp {
				m.ShowHelp = false
				return m, nil
			}
			m.Quitting = true
			return m, tea.Quit

		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case " ":
			m.Engine.TogglePause()
			return m, nil

		case "p", "tab":
			if len(m.Presets) > 1 {
				m.PresetIndex = (m.PresetIndex + 1) % len(m.Presets)
				m.Engine.SetPreset(m.Presets[m.PresetIndex], m.CyclesOverride)
				m.Engine.Restart()
			}
			return m, nil

		case "P", "shift+tab":
			if len(m.Presets) > 1 {
				m.PresetIndex = (m.PresetIndex - 1 + len(m.Presets)) % len(m.Presets)
				m.Engine.SetPreset(m.Presets[m.PresetIndex], m.CyclesOverride)
				m.Engine.Restart()
			}
			return m, nil

		case "+", "=":
			// Speed up (shorten duration)
			m.Engine.AdjustRate(-0.1)
			return m, nil

		case "-", "_":
			// Slow down (lengthen duration)
			m.Engine.AdjustRate(0.1)
			return m, nil

		case "t":
			m.Theme = NextTheme(m.Theme.Name)
			return m, nil

		case "s":
			m.Style = NextStyle(m.Style)
			return m, nil

		case "r":
			m.Engine.Restart()
			return m, nil

		case "?", "h":
			m.ShowHelp = !m.ShowHelp
			return m, nil
		}
	}

	return m, nil
}

// View renders the terminal user interface.
func (m Model) View() string {
	if m.Quitting {
		return "\n  ✨ Thank you for taking time to breathe. Have a mindful day!\n\n"
	}

	if m.Width < 36 || m.Height < 14 {
		return lipgloss.Place(
			m.Width, m.Height,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Foreground(m.Theme.Secondary).Render("Terminal too small\nPlease enlarge window"),
		)
	}

	if m.ShowHelp {
		return m.renderHelpModal()
	}

	var sections []string

	// 1. Header Box
	header := m.renderHeader()
	sections = append(sections, header)

	// 2. Animated Circle Centerpiece
	circleArea := m.renderCircleArea()
	sections = append(sections, circleArea)

	// 3. Guidance Text
	guidance := m.renderGuidance()
	sections = append(sections, guidance)

	// 4. Progress Gauge
	progress := m.renderProgressGauge()
	sections = append(sections, progress)

	// 5. Status Bar
	statusBar := m.renderStatusBar()
	sections = append(sections, statusBar)

	// 6. Navigation Footer
	footer := m.renderFooter()
	sections = append(sections, footer)

	mainContent := lipgloss.JoinVertical(lipgloss.Center, sections...)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, mainContent)
}

func (m Model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.Theme.Primary)

	patternStyle := lipgloss.NewStyle().
		Foreground(m.Theme.TextMuted)

	title := fmt.Sprintf("✦  %s  ✦", m.Engine.Preset.Name)
	pattern := fmt.Sprintf("[%s]", m.Engine.Preset.TimingPattern())

	return lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
		patternStyle.Render(pattern),
	)
}

func (m Model) renderCircleArea() string {
	// Allocate available vertical space to circle
	circleHeight := m.Height - 11
	if circleHeight > 22 {
		circleHeight = 22
	} else if circleHeight < minCircleHeight {
		circleHeight = minCircleHeight
	}

	circleWidth := int(float64(circleHeight)*m.Config.AspectRatio) + 4
	if circleWidth > m.Width-4 {
		circleWidth = m.Width - 4
	}
	if circleWidth < 20 {
		circleWidth = 20
	}

	factor := m.Engine.CurrentRadiusFactor()

	phaseLabel := m.Engine.CurrentPhase.String()
	timeRemaining := m.Engine.PhaseTimeRemaining()
	// Round the countdown up so it ticks naturally into the transition
	// instead of dwelling on 0.0s for the final fraction of a second.
	timeLabel := fmt.Sprintf("%.1fs", math.Ceil(timeRemaining*10)/10)

	if m.Engine.Paused {
		phaseLabel = "PAUSED"
	}

	circleCfg := CircleConfig{
		Style:       m.Style,
		AspectRatio: m.Config.AspectRatio,
		Theme:       m.Theme,
		PhaseLabel:  phaseLabel,
		TimeLabel:   timeLabel,
	}

	return RenderCircle(circleWidth, circleHeight, factor, circleCfg)
}

func (m Model) renderGuidance() string {
	guidance := m.Engine.CurrentPhase.Guidance()
	if m.Engine.Paused {
		guidance = "Press Space to resume breathing"
	}

	style := lipgloss.NewStyle().
		Italic(true).
		Foreground(m.Theme.TextMuted)

	return style.Render(guidance)
}

func (m Model) renderProgressGauge() string {
	progress := m.Engine.PhaseProgress()
	if m.Engine.CurrentPhase == breath.PhaseDone {
		progress = 1.0
	}

	barWidth := 32
	if m.Width < 50 {
		barWidth = 20
	}

	filledLen := int(progress * float64(barWidth))
	if filledLen > barWidth {
		filledLen = barWidth
	}
	emptyLen := barWidth - filledLen

	filledStyle := lipgloss.NewStyle().Foreground(m.Theme.PhaseColor(m.Engine.CurrentPhase))
	emptyStyle := lipgloss.NewStyle().Foreground(m.Theme.BgMuted)

	bar := filledStyle.Render(strings.Repeat("━", filledLen)) + emptyStyle.Render(strings.Repeat("─", emptyLen))
	dur := m.Engine.CurrentPhaseDuration().Seconds()
	elapsed := m.Engine.PhaseElapsed.Seconds()
	if elapsed > dur {
		elapsed = dur
	}

	label := fmt.Sprintf("%.1fs / %.1fs", elapsed, dur)
	labelStyle := lipgloss.NewStyle().Foreground(m.Theme.TextMuted)

	return fmt.Sprintf("%s  %s", bar, labelStyle.Render(label))
}

func (m Model) renderStatusBar() string {
	var cycleStr string
	if m.Engine.TargetCycles > 0 {
		cycleStr = fmt.Sprintf("Cycle %d / %d", m.Engine.CurrentCycle, m.Engine.TargetCycles)
	} else {
		cycleStr = fmt.Sprintf("Cycle %d", m.Engine.CurrentCycle)
	}

	totalMin := int(m.Engine.TotalElapsed.Minutes())
	totalSec := int(m.Engine.TotalElapsed.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", totalMin, totalSec)

	pacePct := int((1.0 / m.Engine.RateScale) * 100)
	paceStr := fmt.Sprintf("%d%% pace", pacePct)

	dimStyle := lipgloss.NewStyle().Foreground(m.Theme.TextMuted)
	activeStyle := lipgloss.NewStyle().Foreground(m.Theme.Secondary)

	dot := dimStyle.Render("  •  ")
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		activeStyle.Render(cycleStr),
		dot,
		dimStyle.Render(timeStr),
		dot,
		dimStyle.Render(paceStr),
		dot,
		dimStyle.Render(fmt.Sprintf("%s (%s)", m.Theme.Name, m.Style)),
	)
}

func (m Model) renderFooter() string {
	keyStyle := lipgloss.NewStyle().Foreground(m.Theme.Primary).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(m.Theme.TextMuted)
	sep := descStyle.Render("   ")

	item := func(key, desc string) string {
		return keyStyle.Render(key) + " " + descStyle.Render(desc)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		item("[Space]", "Pause"),
		sep,
		item("[p]", "Preset"),
		sep,
		item("[+/-]", "Pace"),
		sep,
		item("[t]", "Theme"),
		sep,
		item("[s]", "Style"),
		sep,
		item("[?]", "Help"),
		sep,
		item("[q]", "Quit"),
	)
}

func (m Model) renderHelpModal() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.Theme.Primary).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.Theme.Accent).
		Width(16)

	descStyle := lipgloss.NewStyle().
		Foreground(m.Theme.Text)

	row := func(key, desc string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top, keyStyle.Render(key), descStyle.Render(desc))
	}

	rows := []string{
		titleStyle.Render("Keyboard Shortcuts"),
		"",
		row("Space", "Pause or resume session"),
		row("p / Tab", "Switch to next preset"),
		row("P / Shift+Tab", "Switch to previous preset"),
		row("+ / =", "Speed up pace (shorten phase durations)"),
		row("- / _", "Slow down pace (lengthen phase durations)"),
		row("t", "Cycle color theme (calm, sunset, forest, nord, etc.)"),
		row("s", "Cycle visual circle style (ring, fill, block, ascii)"),
		row("r", "Restart session from cycle 1"),
		row("? / h", "Toggle this help menu"),
		row("q / Esc", "Quit and return to terminal"),
		"",
		lipgloss.NewStyle().Foreground(m.Theme.TextMuted).Render("Press ? or Esc to close help"),
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.Theme.Primary).
		Padding(1, 3).
		Background(m.Theme.BgMuted)

	modal := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modal)
}
