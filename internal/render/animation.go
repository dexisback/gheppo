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

// Animate plays a progressive trace-draw and graph reveal on stdout in interactive terminals.
func Animate(w io.Writer, s *stats.Summary) {
	if s == nil {
		return
	}

	if !ShouldAnimate() {
		fmt.Fprintln(w, Grid(s))
		return
	}

	mode := DetectColorMode()
	themeMode := DetectTheme()
	theme := GetTheme(mode, themeMode)
	termWidth := GetTerminalWidth()

	totalWeeks := 0
	if s.Grid != nil {
		totalWeeks = len(s.Grid)
	}

	layout := CalculateLayout(termWidth, totalWeeks)

	// Sweep across 8 progressive steps (~140ms total duration)
	steps := 8
	stepDelay := 18 * time.Millisecond
	prevLineCount := 0

	for step := 1; step <= steps; step++ {
		progress := float64(step) / float64(steps)
		revealedWeeks := int(float64(layout.VisibleWeeks) * progress)
		if revealedWeeks < 1 {
			revealedWeeks = 1
		}

		// Progressive cell reveal in graph along with progressive trace-draw in wordmark
		stepSummary := createStepSummary(s, layout.VisibleWeeks, revealedWeeks)
		cardOutput := RenderWithThemeAndState(stepSummary, layout, theme, WordmarkState{Progress: progress})

		lines := strings.Split(cardOutput, "\n")
		currentLineCount := len(lines)

		if step == 1 {
			// First frame: print normally without trailing newline
			fmt.Fprint(w, cardOutput)
		} else {
			// Subsequent frames: move cursor up exactly (prevLineCount-1) lines, clear screen below, rewrite
			upCount := prevLineCount - 1
			if upCount < 1 {
				upCount = 1
			}
			fmt.Fprintf(w, "\033[%dA\r\033[J%s", upCount, cardOutput)
		}

		prevLineCount = currentLineCount
		time.Sleep(stepDelay)
	}

	// Move to next line below the card when animation finishes
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
		Followers:     s.Followers,
		Following:     s.Following,
		Repos:         s.Repos,
		TotalStars:    s.TotalStars,
		BestDay:       s.BestDay,
		BestDayCount:  s.BestDayCount,
		DailyAverage:  s.DailyAverage,
		Grid:          clonedGrid,
	}
}
