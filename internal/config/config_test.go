package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg.DefaultPreset != "box" {
		t.Errorf("expected default preset box, got %s", cfg.DefaultPreset)
	}
	if cfg.FPS != 30 {
		t.Errorf("expected default FPS 30, got %d", cfg.FPS)
	}
	if cfg.Theme != "calm" {
		t.Errorf("expected default theme calm, got %s", cfg.Theme)
	}
}

func TestLoadNonExistentDefault(t *testing.T) {
	// A non-existent default path should gracefully return default config
	tempDir := t.TempDir()
	fakeDefault := filepath.Join(tempDir, "nonexistent.yaml")

	cfg, err := Load(fakeDefault)
	if err == nil {
		t.Fatalf("expected error for explicit non-existent file")
	}
	if cfg != nil {
		t.Fatalf("expected nil config on error")
	}
}

func TestLoadCustomYaml(t *testing.T) {
	yamlContent := `
default_preset: "coherent"
theme: "forest"
style: "block"
fps: 60
aspect_ratio: 1.8
sound: true
prep_seconds: 5
presets:
  - name: "custom-1"
    description: "Custom description"
    inhale: 3.5
    hold_in: 1.0
    exhale: 4.5
    hold_out: 0.5
    cycles: 5
`
	tmpFile := filepath.Join(t.TempDir(), "test.yaml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.DefaultPreset != "coherent" {
		t.Errorf("expected default_preset coherent, got %s", cfg.DefaultPreset)
	}
	if cfg.Theme != "forest" {
		t.Errorf("expected theme forest, got %s", cfg.Theme)
	}
	if cfg.Style != "block" {
		t.Errorf("expected style block, got %s", cfg.Style)
	}
	if cfg.FPS != 60 {
		t.Errorf("expected fps 60, got %d", cfg.FPS)
	}
	if cfg.AspectRatio != 1.8 {
		t.Errorf("expected aspect_ratio 1.8, got %f", cfg.AspectRatio)
	}
	if !cfg.Sound {
		t.Errorf("expected sound to be true")
	}
	if cfg.PrepSeconds != 5 {
		t.Errorf("expected prep_seconds 5, got %d", cfg.PrepSeconds)
	}
	if len(cfg.Presets) != 1 {
		t.Fatalf("expected 1 custom preset, got %d", len(cfg.Presets))
	}
	if cfg.Presets[0].Name != "custom-1" || cfg.Presets[0].Inhale != 3.5 {
		t.Errorf("unexpected preset data: %+v", cfg.Presets[0])
	}
}

func TestClampFPS(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, 30}, {-5, 30}, {1, 15}, {10, 15}, {15, 15},
		{30, 30}, {60, 60}, {61, 60}, {120, 60}, {5000, 60},
	}
	for _, c := range cases {
		if got := ClampFPS(c.in); got != c.want {
			t.Errorf("ClampFPS(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestLoadSanitizesControlChars(t *testing.T) {
	yamlContent := `
theme: "cal\x1bm"
style: "bra\x07ille-ring"
presets:
  - name: "es\x1bcape"
    description: "de\x1bscription"
    inhale: 4
    exhale: 4
`
	tmpFile := filepath.Join(t.TempDir(), "test.yaml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Theme != "calm" {
		t.Errorf("expected sanitized theme 'calm', got %q", cfg.Theme)
	}
	if cfg.Style != "braille-ring" {
		t.Errorf("expected sanitized style 'braille-ring', got %q", cfg.Style)
	}
	if len(cfg.Presets) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(cfg.Presets))
	}
	if cfg.Presets[0].Name != "escape" {
		t.Errorf("expected sanitized name 'escape', got %q", cfg.Presets[0].Name)
	}
	if cfg.Presets[0].Description != "description" {
		t.Errorf("expected sanitized description 'description', got %q", cfg.Presets[0].Description)
	}
}

func TestGenerateDefaultConfigFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), ".brth.yaml")

	if err := GenerateDefaultConfigFile(tmpFile); err != nil {
		t.Fatalf("failed to generate default config: %v", err)
	}

	// Should error if generating again over existing file
	if err := GenerateDefaultConfigFile(tmpFile); err == nil {
		t.Fatalf("expected error when config file already exists")
	}

	// Should successfully load the generated config
	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("failed to load generated config: %v", err)
	}
	if cfg.DefaultPreset != "box" {
		t.Errorf("expected box, got %s", cfg.DefaultPreset)
	}
	if len(cfg.Presets) != 2 {
		t.Errorf("expected 2 sample presets in generated config, got %d", len(cfg.Presets))
	}
}
