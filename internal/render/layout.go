package render

import (
	"os"
	"strconv"
	"golang.org/x/term"
)

// Default terminal width and bounds.
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

// Layout encapsulates all computed geometry and component dimensions for a render pass.
type Layout struct {
	TerminalWidth int
	CardWidth     int
	InnerWidth    int
	LeftMargin    int

	// Left Region (Contribution Graph area)
	LeftWidth    int
	VisibleWeeks int
	GraphWidth   int
	GraphHeight  int // Number of weekday rows (7)

	// Divider
	DividerWidth int // " │ " spacing (3 characters)

	// Right Region (Stats area)
	RightWidth int
	StatsWidth int
}

// LayoutConfig is an alias for Layout to maintain backwards compatibility.
type LayoutConfig = Layout

// CalculateLayout calculates responsive layout parameters from terminal width and total available weeks.
// Proportions allocate ~70-75% of content space to the graph and ~25-30% to stats.
func CalculateLayout(termWidth, totalWeeks int) Layout {
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

	// Calculate inside content width (cardWidth - 4 for borders and padding: "│ " and " │")
	innerWidth := cardWidth - 4
	if innerWidth < 30 {
		innerWidth = 30
	}

	// Reserve space for divider: " │ " = 3 characters
	dividerWidth := 3
	availableForContent := innerWidth - dividerWidth
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

	return Layout{
		TerminalWidth: termWidth,
		CardWidth:     cardWidth,
		InnerWidth:    innerWidth,
		LeftMargin:    leftMargin,
		LeftWidth:     graphWidth,
		VisibleWeeks:  visibleWeeks,
		GraphWidth:    graphWidth,
		GraphHeight:   7,
		DividerWidth:  dividerWidth,
		RightWidth:    statsWidth,
		StatsWidth:    statsWidth,
	}
}

// ComputeLayout is retained for backward compatibility.
func ComputeLayout(termWidth, totalWeeks int) LayoutConfig {
	return CalculateLayout(termWidth, totalWeeks)
}
