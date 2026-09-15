package render

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dexisback/gheppo/internal/stats"
	"golang.org/x/term"
)

// Default terminal width when detection fails.
const (
	DefaultTerminalWidth = 80
	MinCardWidth         = 46
	MaxCardWidth         = 140
)

// GetTerminalWidth detects the current terminal width in columns.
func GetTerminalWidth() int {
	if colsStr := os.Getenv("COLUMNS"); colsStr != "" {
		if cols, err := strconv.Atoi(colsStr); err == nil && cols > 0 {
			return cols
		}
	}

	if term.IsTerminal(int(os.Stdout.Fd())) {
		if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
			return width
		}
	}

	return DefaultTerminalWidth
}

// LayoutConfig encapsulates all computed geometry for the current render pass.
type LayoutConfig struct {
	TerminalWidth int
	CardWidth     int
	LeftMargin    int
	VisibleWeeks  int
	GraphWidth    int
	StatsWidth    int
}

// ComputeLayout calculates the responsive layout parameters.
// The graph should occupy ~70-75% of width, stats ~25-30%.
func ComputeLayout(termWidth, totalWeeks int) LayoutConfig {
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	// Determine card width - use most of terminal but respect max
	cardWidth := termWidth - 4
	if cardWidth > MaxCardWidth {
		cardWidth = MaxCardWidth
	}
	if cardWidth < MinCardWidth {
		cardWidth = MinCardWidth
	}

	// Horizontal centering margin for wide terminals
	leftMargin := 0
	if termWidth > cardWidth {
		leftMargin = (termWidth - cardWidth) / 2
	}

	// Calculate inside content width (cardWidth - 4 for borders and padding)
	insideWidth := cardWidth - 4
	if insideWidth < 30 {
		insideWidth = 30
	}

	// Reserve space for divider: " │ " = 3 characters
	availableForContent := insideWidth - 3
	if availableForContent < 20 {
		availableForContent = 20
	}

	// Graph gets ~70-75% of available content space
	graphWidth := (availableForContent * 70) / 100
	statsWidth := availableForContent - graphWidth

	// Ensure stats panel is readable (minimum ~20 chars)
	if statsWidth < 20 {
		statsWidth = 20
		graphWidth = availableForContent - statsWidth
	}

	// Each week occupies 2 characters ("■ ")
	visibleWeeks := graphWidth / 2
	if totalWeeks > 0 && visibleWeeks > totalWeeks {
		visibleWeeks = totalWeeks
	}
	if visibleWeeks < 4 {
		visibleWeeks = 4
	}

	// Adjust graphWidth to actual used width
	graphWidth = visibleWeeks * 2

	return LayoutConfig{
		TerminalWidth: termWidth,
		CardWidth:     cardWidth,
		LeftMargin:    leftMargin,
		VisibleWeeks:  visibleWeeks,
		GraphWidth:    graphWidth,
		StatsWidth:    statsWidth,
	}
}

// BuildMonthHeader creates the aligned month header string matching visible weeks.
func BuildMonthHeader(weeks [][]stats.Cell, graphWidth int) string {
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
