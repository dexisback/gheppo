package render

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

// Animate plays a progressive trace-draw and graph reveal on stdout in interactive terminals using Bubble Tea.
func Animate(w io.Writer, s *stats.Summary) {
	if s == nil {
		return
	}

	if !ShouldAnimate() {
		fmt.Fprintln(w, Grid(s))
		return
	}

	m := NewModel(s)
	p := tea.NewProgram(m, tea.WithOutput(w))
	if _, err := p.Run(); err != nil {
		// Fallback to static output on error
		fmt.Fprintln(w, Grid(s))
		return
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
