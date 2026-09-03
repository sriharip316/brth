package ui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var brailleDotMap = [4][2]rune{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

// CircleConfig configures circle rendering parameters.
type CircleConfig struct {
	Style       string
	AspectRatio float64
	Theme       Theme
	PhaseLabel  string
	TimeLabel   string
}

// AvailableStyles lists valid rendering styles.
var AvailableStyles = []string{"braille-ring", "braille-fill", "block", "ascii"}

// NextStyle returns the next rendering style cyclically.
func NextStyle(current string) string {
	current = strings.ToLower(strings.TrimSpace(current))
	for i, s := range AvailableStyles {
		if s == current {
			nextIdx := (i + 1) % len(AvailableStyles)
			return AvailableStyles[nextIdx]
		}
	}
	return AvailableStyles[0]
}

// RenderCircle generates a centered, aspect-ratio corrected circle animation frame.
// width and height are the terminal canvas dimensions for the circle area.
// factor is [0.0, 1.0] (0.0 = min radius, 1.0 = max radius).
func RenderCircle(width, height int, factor float64, cfg CircleConfig) string {
	if width < 20 {
		width = 20
	}
	if height < 10 {
		height = 10
	}

	if cfg.AspectRatio <= 0 {
		cfg.AspectRatio = 2.0
	}

	switch cfg.Style {
	case "braille-fill":
		return renderBraille(width, height, factor, true, cfg)
	case "block":
		return renderBlock(width, height, factor, cfg)
	case "ascii":
		return renderASCII(width, height, factor, cfg)
	case "braille-ring":
		fallthrough
	default:
		return renderBraille(width, height, factor, false, cfg)
	}
}

// renderBraille renders using Braille patterns (U+2800..U+28FF) for 8x sub-cell resolution.
func renderBraille(cellWidth, cellHeight int, factor float64, filled bool, cfg CircleConfig) string {
	dotW := cellWidth * 2
	dotH := cellHeight * 4
	cx := float64(dotW) / 2.0
	cy := float64(dotH) / 2.0

	// Each braille character has 2 dots horizontally and 4 dots vertically.
	// In a terminal where cell height is AspectRatio times cell width (typically ~2.0),
	// the aspect ratio of each individual braille dot is AspectRatio / (4/2) = AspectRatio / 2.0.
	dotAspect := cfg.AspectRatio / 2.0
	if dotAspect <= 0 {
		dotAspect = 1.0
	}

	margin := 3.0
	maxRadius := math.Min(cx-margin, (cy-margin)*dotAspect)
	if maxRadius < 6 {
		maxRadius = 6
	}
	minRadius := maxRadius * 0.35
	if minRadius < 4 {
		minRadius = 4
	}
	if minRadius >= maxRadius {
		minRadius = maxRadius * 0.5
	}

	radius := minRadius + factor*(maxRadius-minRadius)
	thickness := 1.5 + factor*0.5

	// Sub-pixel grid
	dots := make([][]bool, dotH)
	for y := 0; y < dotH; y++ {
		dots[y] = make([]bool, dotW)
		for x := 0; x < dotW; x++ {
			dx := float64(x) - cx
			dy := (float64(y) - cy) * dotAspect
			dist := math.Sqrt(dx*dx + dy*dy)

			if filled {
				if dist <= radius {
					dots[y][x] = true
				}
			} else {
				if math.Abs(dist-radius) <= thickness {
					dots[y][x] = true
				}
			}
		}
	}

	// Prepare cell character grid
	grid := make([][]rune, cellHeight)
	for cyCell := 0; cyCell < cellHeight; cyCell++ {
		grid[cyCell] = make([]rune, cellWidth)
		for cxCell := 0; cxCell < cellWidth; cxCell++ {
			var r rune = 0x2800
			for dy := 0; dy < 4; dy++ {
				py := cyCell*4 + dy
				for dx := 0; dx < 2; dx++ {
					px := cxCell*2 + dx
					if py < dotH && px < dotW && dots[py][px] {
						r |= brailleDotMap[dy][dx]
					}
				}
			}
			if r == 0x2800 {
				r = ' '
			}
			grid[cyCell][cxCell] = r
		}
	}

	// Overlay text in center of grid
	overlayCenterText(grid, cellWidth, cellHeight, cfg.PhaseLabel, cfg.TimeLabel)

	// Style output lines
	style := lipgloss.NewStyle().Foreground(cfg.Theme.Primary)
	var sb strings.Builder
	for y := 0; y < cellHeight; y++ {
		line := string(grid[y])
		sb.WriteString(style.Render(line))
		if y < cellHeight-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// renderBlock renders with shaded block characters (░▒▓█).
func renderBlock(cellWidth, cellHeight int, factor float64, cfg CircleConfig) string {
	cx := float64(cellWidth) / 2.0
	cy := float64(cellHeight) / 2.0

	margin := 1.5
	maxRadius := math.Min(cx-margin*2, (cy-margin)*cfg.AspectRatio)
	if maxRadius < 4 {
		maxRadius = 4
	}
	minRadius := maxRadius * 0.35
	if minRadius < 3 {
		minRadius = 3
	}
	if minRadius >= maxRadius {
		minRadius = maxRadius * 0.5
	}

	radius := minRadius + factor*(maxRadius-minRadius)

	grid := make([][]rune, cellHeight)
	for y := 0; y < cellHeight; y++ {
		grid[y] = make([]rune, cellWidth)
		for x := 0; x < cellWidth; x++ {
			dx := float64(x) - cx
			dy := (float64(y) - cy) * cfg.AspectRatio
			dist := math.Sqrt(dx*dx + dy*dy)

			diff := dist - radius
			if diff < -2.0 {
				grid[y][x] = '█'
			} else if diff < -1.0 {
				grid[y][x] = '▓'
			} else if diff < 0.0 {
				grid[y][x] = '▒'
			} else if diff < 1.0 {
				grid[y][x] = '░'
			} else {
				grid[y][x] = ' '
			}
		}
	}

	overlayCenterText(grid, cellWidth, cellHeight, cfg.PhaseLabel, cfg.TimeLabel)

	style := lipgloss.NewStyle().Foreground(cfg.Theme.Primary)
	var sb strings.Builder
	for y := 0; y < cellHeight; y++ {
		sb.WriteString(style.Render(string(grid[y])))
		if y < cellHeight-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// renderASCII renders a clean ASCII circle.
func renderASCII(cellWidth, cellHeight int, factor float64, cfg CircleConfig) string {
	cx := float64(cellWidth) / 2.0
	cy := float64(cellHeight) / 2.0

	margin := 1.5
	maxRadius := math.Min(cx-margin*2, (cy-margin)*cfg.AspectRatio)
	if maxRadius < 4 {
		maxRadius = 4
	}
	minRadius := maxRadius * 0.35
	if minRadius < 3 {
		minRadius = 3
	}
	if minRadius >= maxRadius {
		minRadius = maxRadius * 0.5
	}

	radius := minRadius + factor*(maxRadius-minRadius)
	thickness := 1.2

	grid := make([][]rune, cellHeight)
	for y := 0; y < cellHeight; y++ {
		grid[y] = make([]rune, cellWidth)
		for x := 0; x < cellWidth; x++ {
			dx := float64(x) - cx
			dy := (float64(y) - cy) * cfg.AspectRatio
			dist := math.Sqrt(dx*dx + dy*dy)

			if math.Abs(dist-radius) <= thickness {
				grid[y][x] = 'o'
			} else {
				grid[y][x] = ' '
			}
		}
	}

	overlayCenterText(grid, cellWidth, cellHeight, cfg.PhaseLabel, cfg.TimeLabel)

	style := lipgloss.NewStyle().Foreground(cfg.Theme.Primary)
	var sb strings.Builder
	for y := 0; y < cellHeight; y++ {
		sb.WriteString(style.Render(string(grid[y])))
		if y < cellHeight-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// overlayCenterText places phase label and countdown timer cleanly in the center of the grid,
// clearing a margin around them for crystal-clear readability.
func overlayCenterText(grid [][]rune, cellWidth, cellHeight int, phaseLabel, timeLabel string) {
	if cellHeight < 3 || cellWidth < 10 {
		return
	}

	midY := cellHeight / 2
	midX := cellWidth / 2

	// Render Phase label at midY - 1 or midY
	if phaseLabel != "" {
		pRunes := []rune(phaseLabel)
		startX := midX - len(pRunes)/2
		targetY := midY - 1
		if timeLabel == "" {
			targetY = midY
		}

		if targetY >= 0 && targetY < cellHeight {
			// Clear padding left and right
			for px := startX - 1; px <= startX+len(pRunes); px++ {
				if px >= 0 && px < cellWidth {
					grid[targetY][px] = ' '
				}
			}
			for i, r := range pRunes {
				pos := startX + i
				if pos >= 0 && pos < cellWidth {
					grid[targetY][pos] = r
				}
			}
		}
	}

	// Render Time label at midY + 1 (or midY)
	if timeLabel != "" {
		tRunes := []rune(timeLabel)
		startX := midX - len(tRunes)/2
		targetY := midY + 1
		if phaseLabel == "" {
			targetY = midY
		}

		if targetY >= 0 && targetY < cellHeight {
			for px := startX - 1; px <= startX+len(tRunes); px++ {
				if px >= 0 && px < cellWidth {
					grid[targetY][px] = ' '
				}
			}
			for i, r := range tRunes {
				pos := startX + i
				if pos >= 0 && pos < cellWidth {
					grid[targetY][pos] = r
				}
			}
		}
	}
}
