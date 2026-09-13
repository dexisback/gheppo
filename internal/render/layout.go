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
	MaxCardWidth         = 112 // 52 weeks * 2 + 8 padding/borders
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
	TerminalWidth  int
	CardWidth      int
	LeftMargin     int
	VisibleWeeks   int
	GridWidth      int
	UseMicroBanner bool
	IsCompact      bool
}

// ComputeLayout calculates the responsive layout parameters.
func ComputeLayout(termWidth, totalWeeks int) LayoutConfig {
	if termWidth <= 0 {
		termWidth = DefaultTerminalWidth
	}

	// Determine card width
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

	// Calculate inside content width (cardWidth - 2 for borders - 4 for internal padding)
	insideWidth := cardWidth - 6
	if insideWidth < 30 {
		insideWidth = 30
	}

	// Each week occupies 2 characters ("■ ")
	visibleWeeks := insideWidth / 2
	if totalWeeks > 0 && visibleWeeks > totalWeeks {
		visibleWeeks = totalWeeks
	}
	if visibleWeeks < 4 {
		visibleWeeks = 4
	}

	gridWidth := visibleWeeks * 2

	useMicroBanner := cardWidth >= 56
	isCompact := cardWidth < 64

	return LayoutConfig{
		TerminalWidth:  termWidth,
		CardWidth:      cardWidth,
		LeftMargin:     leftMargin,
		VisibleWeeks:   visibleWeeks,
		GridWidth:      gridWidth,
		UseMicroBanner: useMicroBanner,
		IsCompact:      isCompact,
	}
}

// BuildMonthHeader creates the aligned month header string matching visible weeks.
func BuildMonthHeader(weeks [][]stats.Cell, gridWidth int) string {
	if len(weeks) == 0 || gridWidth <= 0 {
		return ""
	}

	monthNames := [...]string{
		"JAN", "FEB", "MAR", "APR", "MAY", "JUN",
		"JUL", "AUG", "SEP", "OCT", "NOV", "DEC",
	}

	chars := make([]byte, gridWidth)
	for i := range chars {
		chars[i] = ' '
	}

	lastMonth := time.Month(0)
	lastPlacedCol := -10

	for wIdx, week := range weeks {
		col := wIdx * 2
		if col+3 > gridWidth {
			break
		}

		for _, cell := range week {
			if cell.Empty || cell.Date.IsZero() {
				continue
			}

			currMonth := cell.Date.Month()
			if currMonth != lastMonth {
				// Avoid collision with previous month label (ensure at least 2 blank spaces between 3-char labels)
				if col >= lastPlacedCol+5 {
					label := monthNames[currMonth-1]
					copy(chars[col:], label)
					lastPlacedCol = col
					lastMonth = currMonth
				}
				break
			}
		}
	}

	return strings.TrimRight(string(chars), " ")
}
