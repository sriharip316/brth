package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/pflag"

	"github.com/srihari/brth/internal/config"
	"github.com/srihari/brth/internal/preset"
	"github.com/srihari/brth/internal/ui"
)

var version = "dev"

func getVersion() string {
	if version != "dev" && version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			return bi.Main.Version
		}
	}
	return version
}

func main() {
	var (
		flagPreset     string
		flagInhale     float64
		flagHoldIn     float64
		flagExhale     float64
		flagHoldOut    float64
		flagCycles     int
		flagConfig     string
		flagTheme      string
		flagStyle      string
		flagFPS        int
		flagSound      bool
		flagList       bool
		flagInitConfig bool
		flagVersion    bool
		flagHelp       bool
	)

	flags := pflag.NewFlagSet("brth", pflag.ContinueOnError)
	flags.StringVarP(&flagPreset, "preset", "p", "", "Name of the breathing preset to use")
	flags.Float64VarP(&flagInhale, "inhale", "i", 0, "Custom inhale duration in seconds")
	flags.Float64Var(&flagHoldIn, "hold-in", 0, "Custom hold-in duration in seconds")
	flags.Float64Var(&flagHoldIn, "hold", 0, "Alias for --hold-in")
	flags.Float64VarP(&flagExhale, "exhale", "e", 0, "Custom exhale duration in seconds")
	flags.Float64Var(&flagHoldOut, "hold-out", 0, "Custom hold-out duration in seconds")
	flags.IntVarP(&flagCycles, "cycles", "c", 0, "Total number of breathing cycles (0 for continuous)")
	flags.StringVar(&flagConfig, "config", "", "Path to custom configuration YAML file")
	flags.StringVar(&flagTheme, "theme", "", "Color theme (calm, sunset, forest, nord, ocean, mono)")
	flags.StringVar(&flagStyle, "style", "", "Circle rendering style (braille-ring, braille-fill, block, ascii)")
	flags.IntVar(&flagFPS, "fps", 0, "Animation framerate (15-60)")
	flags.BoolVarP(&flagSound, "sound", "s", false, "Enable terminal bell on phase changes")
	flags.BoolVarP(&flagList, "list", "l", false, "List all available presets (built-in and config)")
	flags.BoolVar(&flagInitConfig, "init-config", false, "Create default ~/.brth.yaml configuration file")
	flags.BoolVarP(&flagVersion, "version", "v", false, "Display brth version")
	flags.BoolVarP(&flagHelp, "help", "h", false, "Display help message")

	flags.Usage = func() {
		printUsage()
	}

	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == pflag.ErrHelp {
			printUsage()
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if flagHelp {
		printUsage()
		os.Exit(0)
	}

	if flagVersion {
		fmt.Printf("brth version %s\n", getVersion())
		os.Exit(0)
	}

	// 1. Handle --init-config
	if flagInitConfig {
		targetPath := flagConfig
		if targetPath == "" {
			var err error
			targetPath, err = config.DefaultConfigPath()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		if err := config.GenerateDefaultConfigFile(targetPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Created configuration file: %s\n", targetPath)
		os.Exit(0)
	}

	// 2. Load configuration (~/.brth.yaml or custom)
	cfg, err := config.Load(flagConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		cfg = config.NewDefaultConfig()
	}

	// Apply CLI overrides to configuration
	if flagTheme != "" {
		cfg.Theme = flagTheme
	}
	if flagStyle != "" {
		cfg.Style = flagStyle
	}
	if flagFPS > 0 {
		cfg.FPS = flagFPS
	}
	if flagSound {
		cfg.Sound = true
	}

	allPresets := preset.AllPresets(cfg.Presets)

	// 3. Handle --list
	if flagList {
		printPresetList(allPresets, cfg.Presets)
		os.Exit(0)
	}

	// 4. Resolve Target Preset
	var activePreset preset.Preset

	// Case A: Custom duration flags specified via CLI
	hasCustomTiming := flags.Changed("inhale") || flags.Changed("exhale") || flags.Changed("hold") || flags.Changed("hold-in") || flags.Changed("hold-out")
	if hasCustomTiming {
		if flagInhale <= 0 {
			flagInhale = 4.0
		}
		if flagExhale <= 0 {
			flagExhale = 4.0
		}
		activePreset = preset.Preset{
			Name:        "custom",
			Description: "Command-line custom breathing cycle",
			Inhale:      flagInhale,
			HoldIn:      flagHoldIn,
			Exhale:      flagExhale,
			HoldOut:     flagHoldOut,
			Cycles:      flagCycles,
		}
		if err := activePreset.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid custom timings: %v\n", err)
			os.Exit(1)
		}
		// Prepend custom preset so it is immediately active and available
		allPresets = append([]preset.Preset{activePreset}, allPresets...)
	} else {
		// Case B: Preset name from positional argument or flag
		presetName := flagPreset
		if presetName == "" && len(flags.Args()) > 0 {
			presetName = flags.Args()[0]
		}
		if presetName == "" {
			presetName = cfg.DefaultPreset
		}

		selected, found := preset.GetPreset(presetName, cfg.Presets)
		if !found {
			fmt.Fprintf(os.Stderr, "Unknown preset %q.\nRun 'brth --list' to see available presets.\n", presetName)
			os.Exit(1)
		}
		activePreset = selected
	}

	// 5. Launch Bubble Tea TUI
	model := ui.NewModel(cfg, activePreset, allPresets, flagCycles)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running brth: %v\n", err)
		os.Exit(1)
	}
}

func printPresetList(all []preset.Preset, custom []preset.Preset) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00d2ff"))
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#38bdf8")).Width(14)
	patternStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a855f7")).Width(14)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e0f2fe"))
	customBadge := lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Render(" [custom]")

	isCustom := func(name string) bool {
		for _, cp := range custom {
			if strings.EqualFold(cp.Name, name) {
				return true
			}
		}
		return false
	}

	fmt.Println()
	fmt.Println(titleStyle.Render("Available Breathing Presets:"))
	fmt.Println()

	for _, p := range all {
		badge := ""
		if isCustom(p.Name) {
			badge = customBadge
		}
		fmt.Printf("  %s %s %s%s\n",
			nameStyle.Render(p.Name),
			patternStyle.Render("["+p.TimingPattern()+"]"),
			descStyle.Render(p.Description),
			badge,
		)
	}

	fmt.Println()
	fmt.Println("Usage: brth [preset-name]  (e.g., 'brth 4-7-8' or 'brth coherent')")
	fmt.Println()
}

func printUsage() {
	banner := `
  _          _   _     
 | |__  _ __| |_| |__  
 | '_ \| '__| __| '_ \ 
 | |_) | |  | |_| | | |
 |_.__/|_|   \__|_| |_|  Terminal Breathwork Companion
`
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00d2ff")).Render(banner))
	fmt.Println("Usage:")
	fmt.Println("  brth [preset] [flags]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  brth                  # Launch default preset (box breathing)")
	fmt.Println("  brth 4-7-8            # Launch 4-7-8 relaxing sleep breath")
	fmt.Println("  brth coherent         # Launch resonant 5.5s HRV breath")
	fmt.Println("  brth -i 4 -e 6 -c 10  # Custom 4s inhale, 6s exhale, 10 cycles")
	fmt.Println("  brth --list           # List all built-in and user presets")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -p, --preset string   Name of the breathing preset to use")
	fmt.Println("  -i, --inhale float    Custom inhale duration in seconds")
	fmt.Println("      --hold-in float   Custom hold-after-inhale duration in seconds")
	fmt.Println("  -e, --exhale float    Custom exhale duration in seconds")
	fmt.Println("      --hold-out float  Custom hold-after-exhale duration in seconds")
	fmt.Println("  -c, --cycles int      Total breathing cycles (default 0 = continuous)")
	fmt.Println("  -l, --list            List all available presets")
	fmt.Println("      --config string   Custom config file path (default ~/.brth.yaml)")
	fmt.Println("      --init-config     Create default ~/.brth.yaml configuration file")
	fmt.Println("      --theme string    Color theme: calm, sunset, forest, nord, ocean, mono")
	fmt.Println("      --style string    Visual style: braille-ring, braille-fill, block, ascii")
	fmt.Println("      --fps int         Target animation FPS (15-60, default 30)")
	fmt.Println("  -s, --sound           Enable terminal bell on phase changes")
	fmt.Println("  -v, --version         Show version information")
	fmt.Println("  -h, --help            Show help")
	fmt.Println()
	fmt.Println("Interactive Controls (while running):")
	fmt.Println("  Space                 Pause / Resume")
	fmt.Println("  p / Tab               Next preset")
	fmt.Println("  P / Shift+Tab         Previous preset")
	fmt.Println("  + / =                 Speed up pace (+10%)")
	fmt.Println("  -                     Slow down pace (-10%)")
	fmt.Println("  t                     Cycle color theme")
	fmt.Println("  s                     Cycle visual circle style")
	fmt.Println("  r                     Restart session from cycle 1")
	fmt.Println("  ? / h                 Toggle keyboard shortcuts help")
	fmt.Println("  q / Esc / Ctrl+C      Quit")
	fmt.Println()
}
