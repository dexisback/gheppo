package render

import (
	"strings"
	"time"

	"github.com/dexisback/gheppo/internal/stats"
)

// GraphConfig specifies geometry and styling parameters for rendering the contribution graph.
type GraphConfig struct {
	CellWidth    int    // Number of characters per cell (default 1)
	CellGap      int    // Space between cells in columns (default 1)
	BlockGlyph   string // Character used for colored cells (default "■")
	DaysInWeek   int    // Number of days in a week / rows (default 7)
	MonthSpacing int    // Minimum character space between month labels (default 5)
}

// DefaultGraphConfig returns the standard graph configuration.
func DefaultGraphConfig() GraphConfig {
	return GraphConfig{
		CellWidth:    1,
		CellGap:      1,
		BlockGlyph:   "■",
		DaysInWeek:   7,
		MonthSpacing: 5,
	}
}

// ContributionGraph encapsulates sliced weeks ready for rendering.
type ContributionGraph struct {
	Weeks        [][]stats.Cell
	VisibleWeeks int
	Width        int
	Height       int
}

// NewContributionGraph slices the full grid to visible weeks based on layout.
func NewContributionGraph(grid [][]stats.Cell, visibleWeeks int) ContributionGraph {
	var visible [][]stats.Cell
	if len(grid) > 0 {
		start := len(grid) - visibleWeeks
		if start < 0 {
			start = 0
		}
		visible = grid[start:]
	}
	return ContributionGraph{
		Weeks:        visible,
		VisibleWeeks: visibleWeeks,
		Width:        visibleWeeks * 2,
		Height:       7,
	}
}

// RenderContributionGraph renders the contribution grid rows (7 weekday rows) as formatted strings.
//
// Matching GitHub's own graph, every weekday column is a complete 7-box
// rectangle: days that fall outside the calendar (leading padding before
// the first real day and future days of the current week) render as
// "no contribution" cells instead of being left blank.
func RenderContributionGraph(graph ContributionGraph, theme Theme, cfg GraphConfig) []string {
	if cfg.DaysInWeek <= 0 {
		cfg.DaysInWeek = 7
	}

	rows := make([]string, cfg.DaysInWeek)
	for weekday := 0; weekday < cfg.DaysInWeek; weekday++ {
		var row strings.Builder
		for _, week := range graph.Weeks {
			if weekday >= len(week) {
				row.WriteString(theme.CellColor(0))
				row.WriteByte(' ')
				continue
			}

			cell := week[weekday]
			if cell.Empty {
				row.WriteString(theme.CellColor(0))
				row.WriteByte(' ')
				continue
			}

			row.WriteString(theme.CellColor(cell.Bucket))
			row.WriteByte(' ')
		}
		rows[weekday] = row.String()
	}
	return rows
}

// RenderMonthHeader creates an aligned month header string matching visible weeks.
func RenderMonthHeader(weeks [][]stats.Cell, graphWidth int) string {
	if len(weeks) == 0 || graphWidth <= 0 {
		return ""
	}

	monthNames := [...]string{
		"JAN", "FEB", "MAR", "APR", "MAY", "JUN",
		"JUL", "AUG", "SEP", "OCT", "NOV", "DEC",
	}

	chars := make([]byte, graphWidth)
	for i := range chars {
		chars[i] = ' '
	}

	lastMonth := time.Month(0)
	lastPlacedCol := -10

	for wIdx, week := range weeks {
		col := wIdx * 2
		if col+3 > graphWidth {
			break
		}

		for _, cell := range week {
			if cell.Empty || cell.Date.IsZero() {
				continue
			}

			currMonth := cell.Date.Month()
			if currMonth != lastMonth {
				// Avoid collision with previous month label (ensure spacing between 3-char labels)
				if col >= lastPlacedCol+5 {
					label := monthNames[currMonth-1]
					if col+len(label) <= graphWidth {
						copy(chars[col:], label)
						lastPlacedCol = col
						lastMonth = currMonth
					}
				}
				break
			}
		}
	}

	return strings.TrimRight(string(chars), " ")
}

// BuildMonthHeader is retained for backward compatibility.
func BuildMonthHeader(weeks [][]stats.Cell, graphWidth int) string {
	return RenderMonthHeader(weeks, graphWidth)
}
