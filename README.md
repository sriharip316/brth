# brth 🌬️

> Clean, minimalist terminal breathing companion with expanding & contracting animations.

`brth` is an aesthetically refined CLI breathwork application designed to help you relax, focus, or sleep right inside your terminal.

```
                              ✦  box  ✦
                              [4-4-4-4]

                            ⣀⣀⣀⣀⣀⣀⡀
                        ⢀⣴⣾⣿⠿⠟⠛⠛⠛⠿⢿⣿⣶⣄
                      ⢀⣾⣿⠟⠉        ⠈⠙⢿⣿⣆
                     ⣰⣿⡟⠁             ⠙⣿⣷⡀
                    ⢠⣿⡟                ⠘⣿⣧
                    ⢸⣿⠁     INHALE      ⢹⣿
                    ⢸⣿                  ⢸⣿
                    ⠸⣿⣇      2.8s      ⢀⣿⡿
                     ⢻⣿⣆              ⢀⣾⣿⠃
                      ⠙⣿⣷⣄          ⢀⣴⣿⡟⠁
                       ⠈⠙⢿⣿⣶⣤⣄⣀⣀⣀⣤⣴⣾⣿⠟⠉
                          ⠈⠉⠛⠛⠛⠛⠛⠛⠋⠉

                Breathe in deeply through your nose...
            ━━━━━━━━━━━─────────────────────  2.8s / 4.0s
           Cycle 1 / 4  •  00:15  •  100% pace  •  calm (braille-ring)
[Space] Pause   [p] Preset   [+/-] Pace   [t] Theme   [s] Style   [?] Help   [q] Quit
```

---

## Features

- **Expanding & Contracting Animations:** Smooth circle visualization rendered using sub-cell Unicode Braille patterns (`U+2800`–`U+28FF`) with 8× coordinate density, or block shading fallback (`░▒▓█`).
- **Organic Lung Biomechanics:** Uses `easeInOutSine` easing curves so lung expansion and contraction feel natural rather than mechanical.
- **Built-in Science-Backed Presets:** Includes Box Breathing, 4-7-8 Relaxing Breath, Resonant/Coherence Breathing, Awake/Energize, Calm & Recover, and Triangle Breathing.
- **Fully Configurable:** Easily configure durations through CLI flags or `$HOME/.brth.yaml`.
- **Custom Presets:** Define your own custom breathing exercises in `$HOME/.brth.yaml`.
- **Interactive Controls:** Switch presets live, pause/resume, scale pace, switch color themes and visual styles without leaving the session.
- **Zero Dependencies at Runtime:** Compiles into a single standalone binary.

---

## Installation

### From Source

Ensure you have [Go](https://golang.org) 1.22+ installed:

```bash
git clone https://github.com/srihari/brth.git
cd brth
go build -o brth .
```

Move the binary into your `$PATH` (e.g. `/usr/local/bin` or `~/.local/bin`):

```bash
sudo mv brth /usr/local/bin/
```

---

## Quick Start

Launch with default settings (Box Breathing, 4-4-4-4):

```bash
brth
```

Launch a specific preset:

```bash
brth 4-7-8
brth coherent
brth awake
brth calm
```

List all available built-in and user-configured presets:

```bash
brth --list
```

Run custom timings on the fly:

```bash
# 4s Inhale, 2s Hold, 6s Exhale, 2s Hold, for 10 cycles
brth -i 4 --hold 2 -e 6 --hold-out 2 -c 10
```

Generate the default `$HOME/.brth.yaml` configuration file:

```bash
brth --init-config
```

---

## Interactive Hotkeys

While `brth` is running, you can control the experience directly with your keyboard:

| Key | Action |
|---|---|
| `Space` | Pause / Resume session |
| `p` / `Tab` | Switch to next preset |
| `P` / `Shift+Tab` | Switch to previous preset |
| `+` / `=` | Increase breathing speed (shorten durations by 10%) |
| `-` | Decrease breathing speed (lengthen durations by 10%) |
| `t` | Cycle color theme (`calm` → `sunset` → `forest` → `nord` → `ocean` → `mono`) |
| `s` | Cycle visual style (`braille-ring` → `braille-fill` → `block` → `ascii`) |
| `r` | Restart session from cycle 1 |
| `?` / `h` | Toggle keyboard shortcuts help overlay |
| `q` / `Esc` / `Ctrl+C` | Quit and restore terminal |

---

## Built-in Presets

| Name | Pattern | Total Cycle | Description |
|---|---|---|---|
| `box` | 4-4-4-4 | 16s | Box Breathing - Equal durations for focus and stress mitigation |
| `4-7-8` | 4-7-8-0 | 19s | Dr. Weil's relaxing sleep breath to down-regulate nervous system |
| `coherent` | 5.5-0-5.5-0 | 11s | Resonant breathing to optimize Heart Rate Variability (HRV) |
| `calm` | 4-2-6-2 | 14s | Extended exhale to stimulate the parasympathetic vagus nerve |
| `awake` | 4-0-2-0 | 6s | Energizing breath with short exhale to boost alertness |
| `triangle` | 4-4-4-0 | 12s | Classic yogic triangle breathing for concentration |
| `pranayama` | 4-16-8-0 | 28s | Traditional 1:4:2 retention practice for breath capacity |

---

## Configuration (`$HOME/.brth.yaml`)

Create `$HOME/.brth.yaml` automatically via `brth --init-config` or manually:

```yaml
# Default preset to start
default_preset: "box"

# Color theme: calm, sunset, forest, nord, ocean, mono
theme: "calm"

# Visual style: braille-ring, braille-fill, block, ascii
style: "braille-ring"

# Target animation framerate (15-60)
fps: 30

# Character aspect ratio correction (width / height of terminal font)
aspect_ratio: 2.0

# Terminal bell on phase change
sound: false

# Preparation countdown seconds before starting (0 to disable)
prep_seconds: 3

# Custom user-defined presets
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
```

---

## Command-line Options

```
Usage:
  brth [preset] [flags]

Flags:
  -p, --preset string   Name of the breathing preset to use
  -i, --inhale float    Custom inhale duration in seconds
      --hold-in float   Custom hold-after-inhale duration in seconds
  -e, --exhale float    Custom exhale duration in seconds
      --hold-out float  Custom hold-after-exhale duration in seconds
  -c, --cycles int      Total breathing cycles (default 0 = continuous)
  -l, --list            List all available presets
      --config string   Custom config file path (default ~/.brth.yaml)
      --init-config     Create default ~/.brth.yaml configuration file
      --theme string    Color theme: calm, sunset, forest, nord, ocean, mono
      --style string    Visual style: braille-ring, braille-fill, block, ascii
      --fps int         Target animation FPS (15-60, default 30)
  -s, --sound           Enable terminal bell on phase changes
  -v, --version         Show version information
  -h, --help            Show help
```

---

## License

MIT
