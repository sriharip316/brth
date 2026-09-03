# Agent Guide: brth 🌬️

Operating manual and engineering guidelines for AI coding assistants and autonomous agents working on `brth`.

---

## 1. Project Overview & Philosophy

`brth` is an aesthetically refined, minimalist terminal breathwork companion written in Go. It utilizes Charmbracelet's [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm architecture TUI runtime) and [Lip Gloss](https://github.com/charmbracelet/lipgloss) (style definitions) to provide smooth, organic lung expansion and contraction animations directly inside the terminal.

### Core Philosophy
- **Zero Runtime Dependencies**: The project compiles to a single standalone Go binary without external system libraries.
- **Evidence-First Engineering**: Verify facts by inspecting the code, running tests, and checking terminal rendering rather than making assumptions.
- **Decoupled Architecture**: Strict separation between the mathematical breath engine, the configuration system, and the terminal presentation layer.
- **Aesthetic Integrity**: Visuals must remain smooth, calm, and readable across standard 80×24 terminals up to high-DPI full-screen layouts.

---

## 2. Directory Structure & Architecture

```
brth/
├── main.go                     # Entrypoint: CLI parsing, config merge, preset resolution, TUI init
├── internal/
│   ├── breath/                 # Pure breathwork state machine & biomechanics
│   │   ├── engine.go           # Engine struct, Phase enum, RateScale, EaseInOutSine, radius factor
│   │   └── engine_test.go      # State progression & math verification tests
│   ├── preset/                 # Preset domain model & built-in definitions
│   │   ├── preset.go           # Preset struct, BuiltinPresets(), AllPresets(), validation
│   │   └── preset_test.go      # Preset validation & lookup tests
│   ├── config/                 # User configuration handling (~/.brth.yaml)
│   │   ├── config.go           # Config struct, Load(), GenerateDefaultConfigFile(), fallbacks
│   │   └── config_test.go      # YAML parsing & default generation tests
│   └── ui/                     # Presentation layer (Bubble Tea & Lip Gloss)
│       ├── model.go            # Bubble Tea Model (Init, Update, View), key handlers, tick loops
│       ├── circle.go           # Canvas drawing: Braille (U+2800..U+28FF), block, ASCII renderers
│       ├── circle_test.go      # Canvas dimensions, style rendering & bounding tests
│       └── theme.go            # Color palettes (calm, sunset, forest, nord, ocean, mono)
├── .brth.example.yaml          # Sample YAML configuration file
├── go.mod                      # Go module definition (Go 1.22+)
├── go.sum                      # Dependency checksums
└── README.md                   # User-facing documentation and usage manual
```

---

## 3. Package Responsibilities & Boundaries

To preserve maintainability, agents must respect these package boundaries:

### `internal/breath`
- **Responsibility**: Manages phase states (`PhasePrep`, `PhaseInhale`, `PhaseHoldIn`, `PhaseExhale`, `PhaseHoldOut`, `PhaseDone`), elapsed durations, cycle counters, pacing (`RateScale`), and easing mathematics.
- **Rule**: Must **NEVER** import `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, or any UI packages. All engine updates are driven by deterministic timestamps (`Update(now time.Time) bool`).

### `internal/preset`
- **Responsibility**: Defines breathing rhythm data (`Inhale`, `HoldIn`, `Exhale`, `HoldOut`, `Cycles`), validates patterns, and manages built-in presets (`box`, `4-7-8`, `coherent`, `calm`, `awake`, `triangle`, `pranayama`).
- **Rule**: Pure data structures and pure functions. Contains no state or mutation logic.

### `internal/config`
- **Responsibility**: Parses and writes YAML configuration (`~/.brth.yaml`), merges user presets, and provides sensible defaults (default preset: `box`, FPS: `30`, theme: `calm`, style: `braille-ring`, aspect ratio: `2.0`).
- **Rule**: Validates parsed presets with `preset.Validate()`. Does not launch processes or render UI.

### `internal/ui`
- **Responsibility**: Implements the Bubble Tea `Model` interface (`Init`, `Update`, `View`), handles keyboard input, renders the animated circle canvas, displays progress gauges, formats time, and manages color themes.
- **Rule**: Rendering functions (`renderCircleArea`, `renderHeader`, `renderStatusBar`, etc.) must be stateless calculations based solely on the current state of `Model` and `Engine`.

### `main.go`
- **Responsibility**: CLI flag parsing using `github.com/spf13/pflag`, CLI flags overriding file configuration, handling standalone actions (`--list`, `--init-config`, `--version`, `--help`), and starting the Bubble Tea program (`tea.WithAltScreen()`).

---

## 4. TUI Invariants & Implementation Standards

When modifying or extending UI components, adhere to the following rules:

1. **Terminal Minimum Dimensions**:
   - The UI enforces a minimum window dimension check (`Width < 36 || Height < 14`). When below this threshold, it displays a compact warning requesting the user to enlarge their window. Never bypass this check without careful consideration.
2. **Sub-Cell Braille Canvas**:
   - The primary renderer (`braille-ring` and `braille-fill`) uses 8-dot Unicode Braille characters (`0x2800`–`0x28FF`). Each cell contains a 2×4 dot grid.
   - Any modifications to coordinate calculations must account for the font aspect ratio (`cfg.AspectRatio / 2.0` per dot) to ensure circles remain circular rather than vertically squashed ellipses.
3. **Clean Center Text Overlay**:
   - `overlayCenterText` clears surrounding character space in the rune grid to prevent circle dots from colliding with or obscuring the phase text and countdown timer.
4. **No Blocking Operations in Bubble Tea Loop**:
   - Never execute blocking network requests, file I/O, or `time.Sleep` inside `Update()` or `View()`. All animation frames must flow through `tickCmd(fps)`.
5. **Terminal Cleanliness**:
   - The app runs in an alternate screen buffer (`tea.WithAltScreen()`) to ensure no terminal scrollback pollution upon exit. Always ensure graceful restoration upon receiving `q`, `Esc`, or `Ctrl+C`.

---

## 5. Developer Workflows & Commands

Always use the following commands for development, testing, and verification:

### Build and Run
```bash
# Build standalone binary in the current directory
go build -o brth .

# Run directly
go run .

# Run with custom parameters
go run . coherent --fps 60 --theme forest
```

### Test Suite
```bash
# Run all unit tests
go test ./...

# Run all tests with verbose output and coverage
go test -v -cover ./...

# Run tests for a specific package
go test -v ./internal/breath
go test -v ./internal/config
go test -v ./internal/preset
go test -v ./internal/ui
```

### Formatting and Code Hygiene
```bash
# Verify formatting
gofmt -l .

# Auto-format Go files
gofmt -w -s .

# Static analysis
go vet ./...

# Clean up module dependencies
go mod tidy
```

---

## 6. Testing Guidelines

When introducing changes:
- **Engine Logic**: Any modification to phase timing, transitions, or easing must have accompanying tests in `internal/breath/engine_test.go`. Ensure edge cases like 0-second phases (e.g. `coherent` breathing having no holds) are tested.
- **Config & Presets**: Custom YAML structures, validation boundaries (e.g. zero/negative durations), and preset merging must be verified in `internal/config/config_test.go` and `internal/preset/preset_test.go`.
- **UI & Circle Rendering**: Canvas math, boundary bounds, and styling fallbacks must be covered in `internal/ui/circle_test.go`.

---

## 7. Safety Protocols & Agent Guardrails

1. **Git Execution Policy**:
   - **NEVER** execute `git commit` or `git push`. Leave version control actions entirely to the user.
2. **Verification Before Task Completion**:
   - Always run `go test ./...` and `go build` to verify that changes compile and pass all tests before declaring a task finished.
   - Clean up any generated binaries (e.g. `rm -f brth`) if built solely for testing.
3. **Preserve Documentation Integrity**:
   - Preserve all existing comments, docstrings, and documentation unless specifically instructed to modify them.
   - If CLI flags or default configurations change, update `main.go`, `.brth.example.yaml`, `README.md`, and this `AGENTS.md` simultaneously.
