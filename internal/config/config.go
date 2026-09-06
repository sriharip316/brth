package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sriharip316/brth/internal/preset"
	"gopkg.in/yaml.v3"
)

// Animation FPS bounds, enforced identically on every input path
// (config file and CLI flags). Docs advertise this same range.
const (
	DefaultFPS = 30
	MinFPS     = 15
	MaxFPS     = 60
)

// ClampFPS constrains fps to the supported animation range, substituting the
// default for unset (zero or negative) values.
func ClampFPS(fps int) int {
	switch {
	case fps <= 0:
		return DefaultFPS
	case fps < MinFPS:
		return MinFPS
	case fps > MaxFPS:
		return MaxFPS
	default:
		return fps
	}
}

// Config represents the application configuration.
type Config struct {
	DefaultPreset string          `yaml:"default_preset"`
	Theme         string          `yaml:"theme"`
	Style         string          `yaml:"style"`
	FPS           int             `yaml:"fps"`
	AspectRatio   float64         `yaml:"aspect_ratio"`
	Sound         bool            `yaml:"sound"`
	PrepSeconds   int             `yaml:"prep_seconds"`
	Presets       []preset.Preset `yaml:"presets"`
}

// NewDefaultConfig returns a configuration with recommended defaults.
func NewDefaultConfig() *Config {
	return &Config{
		DefaultPreset: "box",
		Theme:         "calm",
		Style:         "braille-ring",
		FPS:           DefaultFPS,
		AspectRatio:   2.0,
		Sound:         false,
		PrepSeconds:   3,
		Presets:       []preset.Preset{},
	}
}

// DefaultConfigPath returns the default path to ~/.brth.yaml.
func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user home directory: %w", err)
	}
	return filepath.Join(home, ".brth.yaml"), nil
}

// Load loads configuration from the given path. If path is empty, DefaultConfigPath() is used.
// If the default config file does not exist, a default configuration is returned without error.
// If a non-empty custom path does not exist, an error is returned.
func Load(path string) (*Config, error) {
	cfg := NewDefaultConfig()

	isDefault := false
	if path == "" {
		isDefault = true
		defaultPath, err := DefaultConfigPath()
		if err != nil {
			return cfg, nil // gracefully fallback to defaults
		}
		path = defaultPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if isDefault {
				return cfg, nil
			}
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config %s: %w", path, err)
	}

	// Apply fallbacks for missing/zero fields
	if cfg.DefaultPreset == "" {
		cfg.DefaultPreset = "box"
	}
	if cfg.Theme == "" {
		cfg.Theme = "calm"
	}
	if cfg.Style == "" {
		cfg.Style = "braille-ring"
	}
	cfg.FPS = ClampFPS(cfg.FPS)
	if cfg.AspectRatio <= 0 {
		cfg.AspectRatio = 2.0
	}
	if cfg.PrepSeconds < 0 {
		cfg.PrepSeconds = 0
	}

	// Strip terminal control runes from config-supplied strings so an
	// untrusted config file cannot inject escape sequences when rendered.
	cfg.DefaultPreset = preset.SanitizeControlChars(cfg.DefaultPreset)
	cfg.Theme = preset.SanitizeControlChars(cfg.Theme)
	cfg.Style = preset.SanitizeControlChars(cfg.Style)
	for i := range cfg.Presets {
		cfg.Presets[i].Name = preset.SanitizeControlChars(cfg.Presets[i].Name)
		cfg.Presets[i].Description = preset.SanitizeControlChars(cfg.Presets[i].Description)
	}

	// Validate custom presets
	for i, p := range cfg.Presets {
		if err := p.Validate(); err != nil {
			return nil, fmt.Errorf("invalid custom preset [%d] %q: %w", i, p.Name, err)
		}
	}

	return cfg, nil
}

// SampleConfigContent returns commented YAML suitable for ~/.brth.yaml.
func SampleConfigContent() string {
	return `# ~/.brth.yaml - Configuration for brth CLI breathing app

# Default preset to start if not specified on the command line
default_preset: "box"

# Color theme: calm, sunset, forest, nord, ocean, mono
theme: "calm"

# Animation visual style: braille-ring, braille-fill, block
style: "braille-ring"

# Animation refresh rate in frames per second (15 to 60)
fps: 30

# Character aspect ratio correction (height / width ratio of terminal font)
# Default 2.0 works well for most terminal fonts
aspect_ratio: 2.0

# Sound notification (terminal bell) on phase changes
sound: false

# Preparation countdown seconds before starting cycle (0 to disable)
prep_seconds: 3

# Custom user presets
# Durations are in seconds (can be fractional, e.g. 5.5)
# Cycles set to 0 runs continuously until stopped
presets:
  - name: "my-zen"
    description: "Deep relaxation rhythm"
    inhale: 4.0
    hold_in: 4.0
    exhale: 6.0
    hold_out: 2.0
    cycles: 10

  - name: "quick-reset"
    description: "3-minute stress reset"
    inhale: 4.0
    hold_in: 0.0
    exhale: 6.0
    hold_out: 0.0
    cycles: 18
`
}

// GenerateDefaultConfigFile writes a documented default ~/.brth.yaml.
func GenerateDefaultConfigFile(path string) (err error) {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Avoid overwriting existing configuration. O_EXCL fails atomically if the
	// file already exists, closing the stat-then-write race window.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("config file already exists at %s", path)
		}
		return fmt.Errorf("failed to create config file %s: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err := io.WriteString(f, SampleConfigContent()); err != nil {
		return fmt.Errorf("failed to write default config to %s: %w", path, err)
	}

	return nil
}
