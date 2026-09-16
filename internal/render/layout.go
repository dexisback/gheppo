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

	statsWidth := 26
	dividerWidth := 5
	if termWidth < 78 {
		dividerWidth = 3
	}

	// Budget horizontal margins for breathing room
	marginBudget := 8
	if termWidth < 75 {
		marginBudget = 4
	} else if termWidth >= 100 {
		marginBudget = (termWidth * 15) / 100
		if marginBudget < 10 {
			marginBudget = 10
		}
	}

	availForGraph := termWidth - marginBudget - statsWidth - dividerWidth
	if availForGraph < 20 {
		availForGraph = 20
	}

	visibleWeeks := availForGraph / 2
	if visibleWeeks > 52 {
		visibleWeeks = 52
	}
	if visibleWeeks < 16 && termWidth >= 65 {
		visibleWeeks = 16
	}
	if totalWeeks > 0 && visibleWeeks > totalWeeks {
		visibleWeeks = totalWeeks
	}
	if visibleWeeks < 4 {
		visibleWeeks = 4
	}

	graphWidth := visibleWeeks * 2
	totalContent := graphWidth + dividerWidth + statsWidth

	leftMargin := 0
	if termWidth > totalContent {
		leftMargin = (termWidth - totalContent) / 2
	}
	if leftMargin < 0 {
		leftMargin = 0
	}

	cardWidth := totalContent + leftMargin*2
	if cardWidth > termWidth {
		cardWidth = termWidth
	}
	if cardWidth < MinCardWidth {
		cardWidth = MinCardWidth
	}

	return Layout{
		TerminalWidth: termWidth,
		CardWidth:     cardWidth,
		InnerWidth:    totalContent,
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
