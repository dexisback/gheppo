package render

import (
	"strings"
	"unicode/utf8"
)

// WordmarkPoint represents a single geometric point or block in the wordmark canvas.
type WordmarkPoint struct {
	Row       int     // 0, 1, 2
	Col       int     // 0-indexed column
	Glyph     string  // UTF-8 display glyph
	AsciiAlt  string  // ASCII fallback glyph
	Intensity int     // 0..4 (palette level: 0=dimmest/empty, 1..4=activity levels)
	Threshold float64 // Progress threshold [0.0, 1.0] when this point activates
}

// WordmarkCanvas holds the geometric model of the custom GHEPPO wordmark.
type WordmarkCanvas struct {
	Width  int
	Height int
	Points []WordmarkPoint
}

// WordmarkState captures animation or visual phase state for wordmark rendering.
type WordmarkState struct {
	Progress float64 // 0.0 to 1.0
}

const (
	defaultWordmarkWidth = 46
	maxTraceShift        = 14
)

// BuildWordmarkCanvas constructs the default full-width (46 cols) geometric representation of the
// procedural trace and the custom block-letter GHEPPO wordmark matching Image 3.
func BuildWordmarkCanvas() WordmarkCanvas {
	return BuildWordmarkCanvasWithWidth(defaultWordmarkWidth)
}

// BuildWordmarkCanvasWithWidth constructs the geometric wordmark adapted to fit within maxWidth.
func BuildWordmarkCanvasWithWidth(maxWidth int) WordmarkCanvas {
	if maxWidth <= 0 || maxWidth >= defaultWordmarkWidth {
		maxWidth = defaultWordmarkWidth
	}

	shift := 0
	if maxWidth < defaultWordmarkWidth {
		shift = defaultWordmarkWidth - maxWidth
		if shift > maxTraceShift {
			shift = maxTraceShift // Preserve all lettering intact
		}
	}

	var rawPts []WordmarkPoint
	add := func(row, col int, glyph, ascii string, intensity int, threshold float64) {
		rawPts = append(rawPts, WordmarkPoint{
			Row:       row,
			Col:       col,
			Glyph:     glyph,
			AsciiAlt:  ascii,
			Intensity: intensity,
			Threshold: threshold,
		})
	}

	// =========================================================================
	// 1. TRACE PATH & CONNECTOR (Cols 0..13) — Progress: [0.00, 0.35]
	// =========================================================================
	// Row 0 satellite dots & top bracket
	add(0, 4, "▪", "-", 1, 0.08)
	add(0, 8, "▪", "-", 2, 0.18)
	add(0, 13, "▄", "-", 3, 0.32)

	// Row 1 main horizontal data path leading directly into G
	add(1, 0, "▪", ".", 1, 0.00)
	add(1, 2, "·", ".", 1, 0.05)
	add(1, 4, "▪", "-", 2, 0.10)
	add(1, 6, "▪", "-", 2, 0.15)
	add(1, 8, "▪", "-", 2, 0.20)
	add(1, 10, "█", "=", 3, 0.25)
	add(1, 12, "█", "=", 3, 0.30)

	// Row 2 satellite dots & bottom bracket
	add(2, 2, "▪", "-", 1, 0.06)
	add(2, 6, "▪", "-", 2, 0.16)
	add(2, 13, "▀", "-", 3, 0.32)

	// =========================================================================
	// 2. LETTER G (Cols 14..18, Width 5) — Progress: [0.35, 0.48]
	// Row 0: █████
	// Row 1: █  ██
	// Row 2: █████
	// =========================================================================
	gOffset := 14
	for c := 0; c < 5; c++ {
		t := 0.35 + float64(c)*0.026
		add(0, gOffset+c, "█", "#", 3, t)
		if c == 0 || c >= 3 {
			add(1, gOffset+c, "█", "#", 3, t)
		}
		add(2, gOffset+c, "█", "#", 3, t)
	}

	// Space at 19

	// =========================================================================
	// 3. LETTER H (Cols 20..24, Width 5) — Progress: [0.48, 0.58]
	// Row 0: █   █
	// Row 1: █████
	// Row 2: █   █
	// =========================================================================
	hOffset := 20
	for c := 0; c < 5; c++ {
		t := 0.48 + float64(c)*0.02
		if c == 0 || c == 4 {
			add(0, hOffset+c, "█", "#", 3, t)
			add(2, hOffset+c, "█", "#", 3, t)
		}
		add(1, hOffset+c, "█", "#", 3, t)
	}

	// Space at 25

	// =========================================================================
	// 4. LETTER E (Cols 26..29, Width 4) — Progress: [0.58, 0.68]
	// Row 0: ████
	// Row 1: ███ 
	// Row 2: ████
	// =========================================================================
	eOffset := 26
	for c := 0; c < 4; c++ {
		t := 0.58 + float64(c)*0.025
		add(0, eOffset+c, "█", "#", 3, t)
		if c < 3 {
			add(1, eOffset+c, "█", "#", 3, t)
		}
		add(2, eOffset+c, "█", "#", 3, t)
	}

	// Space at 30

	// =========================================================================
	// 5. LETTER P1 (Cols 31..34, Width 4) — Progress: [0.68, 0.78]
	// Row 0: ████
	// Row 1: █  █
	// Row 2: █   
	// =========================================================================
	p1Offset := 31
	for c := 0; c < 4; c++ {
		t := 0.68 + float64(c)*0.025
		add(0, p1Offset+c, "█", "#", 3, t)
		if c == 0 || c == 3 {
			add(1, p1Offset+c, "█", "#", 3, t)
		}
		if c == 0 {
			add(2, p1Offset+c, "█", "#", 3, t)
		}
	}

	// Space at 35

	// =========================================================================
	// 6. LETTER P2 (Cols 36..39, Width 4) — Progress: [0.78, 0.88]
	// Row 0: ████
	// Row 1: █  █
	// Row 2: █   
	// =========================================================================
	p2Offset := 36
	for c := 0; c < 4; c++ {
		t := 0.78 + float64(c)*0.025
		add(0, p2Offset+c, "█", "#", 3, t)
		if c == 0 || c == 3 {
			add(1, p2Offset+c, "█", "#", 3, t)
		}
		if c == 0 {
			add(2, p2Offset+c, "█", "#", 3, t)
		}
	}

	// Space at 40

	// =========================================================================
	// 7. LETTER O (Cols 41..45, Width 5) — Progress: [0.88, 1.00]
	// Row 0: █████
	// Row 1: █   █
	// Row 2: █████
	// =========================================================================
	oOffset := 41
	for c := 0; c < 5; c++ {
		t := 0.88 + float64(c)*0.024
		add(0, oOffset+c, "█", "#", 3, t)
		if c == 0 || c == 4 {
			add(1, oOffset+c, "█", "#", 3, t)
		}
		add(2, oOffset+c, "█", "#", 3, t)
	}

	finalWidth := defaultWordmarkWidth - shift
	var shiftedPts []WordmarkPoint
	for _, pt := range rawPts {
		if pt.Col >= shift {
			newPt := pt
			newPt.Col = pt.Col - shift
			shiftedPts = append(shiftedPts, newPt)
		}
	}

	return WordmarkCanvas{
		Width:  finalWidth,
		Height: 3,
		Points: shiftedPts,
	}
}

// GetWordmark returns the complete wordmark string for backward compatibility.
func GetWordmark(mode ColorMode) string {
	theme := GetTheme(mode, ThemeDark)
	lines, _ := RenderWordmarkWithState(theme, WordmarkState{Progress: 1.0})
	return strings.Join(lines, "\n")
}

// RenderWordmark returns the fully resolved 3-line slice and visible width for static display.
func RenderWordmark(theme Theme) ([]string, int) {
	return RenderWordmarkWithState(theme, WordmarkState{Progress: 1.0})
}

// RenderWordmarkWithState renders the wordmark with an external animation progress state.
func RenderWordmarkWithState(theme Theme, state WordmarkState) ([]string, int) {
	return RenderWordmarkWithStateAndWidth(theme, state, defaultWordmarkWidth)
}

// RenderWordmarkWithStateAndWidth renders the wordmark with animation state and width constraint.
func RenderWordmarkWithStateAndWidth(theme Theme, state WordmarkState, maxWidth int) ([]string, int) {
	canvas := BuildWordmarkCanvasWithWidth(maxWidth)
	progress := state.Progress
	if progress < 0.0 {
		progress = 0.0
	}
	if progress > 1.0 {
		progress = 1.0
	}

	// Build color lookup from theme tokens
	colors := map[int]string{
		0: theme.EmptyCell,
		1: theme.ContributionLevel1,
		2: theme.ContributionLevel2,
		3: theme.ContributionLevel3,
		4: theme.ContributionLevel4,
	}

	// Create 2D grid
	grid := make([][]string, canvas.Height)
	for r := 0; r < canvas.Height; r++ {
		grid[r] = make([]string, canvas.Width)
		for c := 0; c < canvas.Width; c++ {
			grid[r][c] = " "
		}
	}

	for _, pt := range canvas.Points {
		if progress >= pt.Threshold {
			if theme.Mode == ColorASCII {
				grid[pt.Row][pt.Col] = pt.AsciiAlt
			} else {
				intensity := pt.Intensity
				// Glow the leading edge during animation sweep
				if progress < 1.0 && progress-pt.Threshold < 0.06 {
					intensity = 4
				}
				color := colors[intensity]
				grid[pt.Row][pt.Col] = color + pt.Glyph + theme.Reset
			}
		}
	}

	lines := make([]string, canvas.Height)
	for r := 0; r < canvas.Height; r++ {
		lines[r] = strings.Join(grid[r], "")
	}

	return lines, canvas.Width
}

// WordmarkVisibleWidth returns the column width of the custom wordmark canvas.
func WordmarkVisibleWidth() int {
	return defaultWordmarkWidth
}

// RuneCountWithoutAnsi measures visible terminal columns ignoring ANSI escapes.
func RuneCountWithoutAnsi(s string) int {
	inEscape := false
	count := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		count += utf8.RuneCountInString(string(r))
	}
	return count
}
