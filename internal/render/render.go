package render

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dexisback/gheppo/internal/stats"
)

// Grid renders a complete, responsive contribution card with stats panel using the detected terminal width.
func Grid(s *stats.Summary) string {
	if s == nil {
		return ""
	}
	termWidth := GetTerminalWidth()
	return Render(s, termWidth)
}

// Render calculates layout and renders a complete contribution card for the given terminal width.
func Render(s *stats.Summary, termWidth int) string {
	if s == nil {
		return ""
	}

	mode := DetectColorMode()
	themeMode := DetectTheme()
	theme := GetTheme(mode, themeMode)

	totalWeeks := 0
	if s.Grid != nil {
		totalWeeks = len(s.Grid)
	}

	layout := CalculateLayout(termWidth, totalWeeks)
	return RenderWithTheme(s, layout, theme)
}

// RenderCard builds the full formatted card string according to layout, mode, and theme.
// Retained for backward compatibility.
func RenderCard(s *stats.Summary, layout LayoutConfig, mode ColorMode, themeMode ThemeMode) string {
	if s == nil {
		return ""
	}
	theme := GetTheme(mode, themeMode)
	return RenderWithTheme(s, layout, theme)
}

// RenderWithTheme produces the fully composed terminal card from the given layout and theme.
func RenderWithTheme(s *stats.Summary, layout Layout, theme Theme) string {
	return RenderWithThemeAndState(s, layout, theme, WordmarkState{Progress: 1.0})
}

// RenderWithThemeAndState produces the terminal card from layout, theme, and animated wordmark state.
func RenderWithThemeAndState(s *stats.Summary, layout Layout, theme Theme, wordmarkState WordmarkState) string {
	if s == nil {
		return ""
	}

	// 1. Prepare components
	graph := NewContributionGraph(s.Grid, layout.VisibleWeeks)
	graphConfig := DefaultGraphConfig()
	graphRows := RenderContributionGraph(graph, theme, graphConfig)
	monthHeader := RenderMonthHeader(graph.Weeks, layout.GraphWidth)
	statsRows := RenderStatsPanel(s, theme)

	// 2. Compose into outer card
	return ComposeWithState(s, layout, theme, graphRows, monthHeader, statsRows, wordmarkState)
}

// Compose assembles all rendered visual components inside a single outer frame with fully resolved wordmark.
func Compose(
	s *stats.Summary,
	layout Layout,
	theme Theme,
	graphRows []string,
	monthHeader string,
	statsRows []StatsRow,
) string {
	return ComposeWithState(s, layout, theme, graphRows, monthHeader, statsRows, WordmarkState{Progress: 1.0})
}

// ComposeWithState assembles all rendered components with external wordmark animation progress state.
func ComposeWithState(
	s *stats.Summary,
	layout Layout,
	theme Theme,
	graphRows []string,
	monthHeader string,
	statsRows []StatsRow,
	wordmarkState WordmarkState,
) string {
	var (
		cornerTL, cornerTR, cornerBL, cornerBR string
		horizontalBar, verticalBar             string
	)

	if theme.Mode == ColorASCII {
		cornerTL, cornerTR, cornerBL, cornerBR = "+", "+", "+", "+"
		horizontalBar, verticalBar = "-", "|"
	} else {
		cornerTL, cornerTR, cornerBL, cornerBR = "┌", "┐", "└", "┘"
		horizontalBar, verticalBar = "─", "│"
	}

	innerWidth := layout.InnerWidth
	if innerWidth < 20 {
		innerWidth = 20
	}

	var lines []string

	addLine := func(content string, visibleLen int) {
		lines = append(lines, formatFramedLine(content, visibleLen, innerWidth, layout.LeftMargin, theme, verticalBar))
	}

	addEmptyLine := func() {
		addLine("", 0)
	}

	// === TOP BORDER ===
	topBorder := fmt.Sprintf("%s%s%s%s%s",
		theme.Border,
		cornerTL,
		strings.Repeat(horizontalBar, layout.CardWidth-2),
		cornerTR,
		theme.Reset,
	)
	if layout.LeftMargin > 0 {
		topBorder = strings.Repeat(" ", layout.LeftMargin) + topBorder
	}
	lines = append(lines, topBorder)

	addEmptyLine()

	// === HEADER: CUSTOM WORDMARK + YEAR PROGRESS BAR ===
	wordmarkLines, wordmarkVisLen := RenderWordmarkWithStateAndWidth(theme, wordmarkState, innerWidth)

	year := time.Now().Year()
	if !s.FetchedAt.IsZero() {
		year = s.FetchedAt.Year()
	}
	yearProgress := CalculateYearProgress()

	// Progress bar width adapts to right stats panel width
	barWidth := layout.StatsWidth - 9
	if barWidth < 5 {
		barWidth = 5
	}
	if barWidth > 25 {
		barWidth = 25
	}

	progressStr, progressVisLen := RenderYearProgress(year, yearProgress, barWidth, theme)

	// Combine Wordmark Row 0 with Year Progress on the right if space allows
	if wordmarkVisLen+progressVisLen+2 <= innerWidth {
		spacing := innerWidth - wordmarkVisLen - progressVisLen
		headerRow0 := fmt.Sprintf("%s%s%s",
			wordmarkLines[0],
			strings.Repeat(" ", spacing),
			progressStr,
		)
		addLine(headerRow0, innerWidth)
	} else {
		addLine(wordmarkLines[0], wordmarkVisLen)
	}

	// Wordmark Row 1 & Row 2
	if len(wordmarkLines) > 1 {
		addLine(wordmarkLines[1], wordmarkVisLen)
	}
	if len(wordmarkLines) > 2 {
		addLine(wordmarkLines[2], wordmarkVisLen)
	}

	// If progress bar could not fit inline on Row 0, output on its own line
	if wordmarkVisLen+progressVisLen+2 > innerWidth {
		addLine(progressStr, progressVisLen)
	}

	addEmptyLine()

	// === SUBHEADER: TOTAL CONTRIBUTIONS ===
	var numStr, labelStr string
	numStr = formatNumber(s.Total)
	if s.Total == 1 {
		labelStr = "CONTRIBUTION"
	} else {
		labelStr = "CONTRIBUTIONS"
	}

	var totalContent string
	if theme.Mode == ColorASCII {
		totalContent = fmt.Sprintf("%s %s", numStr, labelStr)
	} else {
		totalContent = fmt.Sprintf("%s%s%s%s %s%s%s",
			theme.Bold, theme.Primary, numStr, theme.Reset,
			theme.Secondary, labelStr, theme.Reset,
		)
	}
	totalVisLen := utf8.RuneCountInString(numStr) + 1 + utf8.RuneCountInString(labelStr)
	addLine(totalContent, totalVisLen)

	addEmptyLine()

	// === MONTH LABELS ===
	if monthHeader != "" {
		coloredMonth := fmt.Sprintf("%s%s%s", theme.Secondary, monthHeader, theme.Reset)
		addLine(coloredMonth, utf8.RuneCountInString(monthHeader))
	}

	// === TWO-COLUMN LAYOUT: GRAPH | STATS ===
	maxRows := layout.GraphHeight
	if len(statsRows) > maxRows {
		maxRows = len(statsRows)
	}

	divider := theme.Border + verticalBar + theme.Reset

	for i := 0; i < maxRows; i++ {
		var graphPart string
		if i < len(graphRows) {
			graphPart = graphRows[i]
		} else {
			graphPart = strings.Repeat(" ", layout.GraphWidth)
		}

		var statsPart string
		statsVisLen := 0
		if i < len(statsRows) {
			statsPart = statsRows[i].Content
			statsVisLen = statsRows[i].VisibleLen
		}

		statsAreaWidth := innerWidth - layout.GraphWidth - layout.DividerWidth
		if statsAreaWidth < 0 {
			statsAreaWidth = 0
		}

		statsPadding := statsAreaWidth - statsVisLen
		if statsPadding < 0 {
			statsPadding = 0
		}

		combinedContent := graphPart + " " + divider + " " + statsPart + strings.Repeat(" ", statsPadding)
		addLine(combinedContent, innerWidth)
	}

	addEmptyLine()

	// === BOTTOM BORDER ===
	bottomBorder := fmt.Sprintf("%s%s%s%s%s",
		theme.Border,
		cornerBL,
		strings.Repeat(horizontalBar, layout.CardWidth-2),
		cornerBR,
		theme.Reset,
	)
	if layout.LeftMargin > 0 {
		bottomBorder = strings.Repeat(" ", layout.LeftMargin) + bottomBorder
	}
	lines = append(lines, bottomBorder)

	return strings.Join(lines, "\n")
}

// formatFramedLine formats a single row within the outer box border with left and right margins/padding.
func formatFramedLine(content string, visibleLen int, innerWidth int, leftMargin int, theme Theme, verticalBar string) string {
	pad := innerWidth - visibleLen
	if pad < 0 {
		pad = 0
	}
	var b strings.Builder
	if leftMargin > 0 {
		b.WriteString(strings.Repeat(" ", leftMargin))
	}
	b.WriteString(theme.Border)
	b.WriteString(verticalBar)
	b.WriteString(" ")
	b.WriteString(theme.Reset)
	b.WriteString(content)
	b.WriteString(strings.Repeat(" ", pad))
	b.WriteString(theme.Border)
	b.WriteString(" ")
	b.WriteString(verticalBar)
	b.WriteString(theme.Reset)
	return b.String()
}
