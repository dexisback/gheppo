package render

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dexisback/gheppo/internal/stats"
	"golang.org/x/term"
)

// ShouldAnimate reports whether interactive terminal reveal animation is suitable.
func ShouldAnimate() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("CI") != "" {
		return false
	}
	if os.Getenv("GHEPPO_NO_ANIMATION") != "" {
		return false
	}
	termEnv := strings.ToLower(os.Getenv("TERM"))
	if termEnv == "dumb" || termEnv == "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Animate plays a progressive 120ms graph reveal on stdout in interactive terminals.
func Animate(w io.Writer, s *stats.Summary) {
	if s == nil {
		return
	}

	if !ShouldAnimate() {
		fmt.Fprintln(w, Grid(s))
		return
	}

	mode := DetectColorMode()
	theme := DetectTheme()
	termWidth := GetTerminalWidth()

	totalWeeks := 0
	if s.Grid != nil {
		totalWeeks = len(s.Grid)
	}

	layout := ComputeLayout(termWidth, totalWeeks)

	// Step count for progressive sweep
	steps := 5
	stepDelay := 20 * time.Millisecond

	// Render progressive frames
	for step := 1; step <= steps; step++ {
		revealedWeeks := (layout.VisibleWeeks * step) / steps
		if revealedWeeks < 1 {
			revealedWeeks = 1
		}

		// Create a temporary summary copy with progressive cell reveal
		stepSummary := createStepSummary(s, layout.VisibleWeeks, revealedWeeks)
		cardOutput := RenderCard(stepSummary, layout, mode, theme)

		lines := strings.Split(cardOutput, "\n")
		lineCount := len(lines)

		if step == 1 {
			// First frame: print normally
			fmt.Fprint(w, cardOutput)
		} else {
			// Subsequent frames: move cursor up to beginning of card and rewrite
			fmt.Fprintf(w, "\033[%dA\r%s", lineCount-1, cardOutput)
		}

		time.Sleep(stepDelay)
	}

	// Move to next line below the card
	fmt.Fprintln(w)
}

// createStepSummary creates a copy of Summary where columns beyond revealedLimit are dimmed to 0.
func createStepSummary(s *stats.Summary, visibleCount, revealedLimit int) *stats.Summary {
	if s == nil || s.Grid == nil {
		return s
	}

	total := len(s.Grid)
	start := total - visibleCount
	if start < 0 {
		start = 0
	}

	clonedGrid := make([][]stats.Cell, len(s.Grid))
	for i, week := range s.Grid {
		clonedGrid[i] = make([]stats.Cell, len(week))
		copy(clonedGrid[i], week)
	}

	// Dim columns beyond revealedLimit in the visible slice
	for vIdx := revealedLimit; vIdx < visibleCount; vIdx++ {
		wIdx := start + vIdx
		if wIdx >= 0 && wIdx < len(clonedGrid) {
			for cIdx := range clonedGrid[wIdx] {
				if !clonedGrid[wIdx][cIdx].Empty {
					clonedGrid[wIdx][cIdx].Bucket = 0
				}
			}
		}
	}

	return &stats.Summary{
		Login:         s.Login,
		Total:         s.Total,
		CurrentStreak: s.CurrentStreak,
		LongestStreak: s.LongestStreak,
		FetchedAt:     s.FetchedAt,
		Grid:          clonedGrid,
	}
}
