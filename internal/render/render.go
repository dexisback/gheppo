package render

import (
	"fmt"
	"strings"
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
	statsRows := RenderStatsPanelWithWidth(s, theme, layout.StatsWidth)

	// 2. Compose into clean dashboard matching Image 3
	return ComposeWithState(s, layout, theme, graphRows, monthHeader, statsRows, wordmarkState)
}

// Compose assembles all rendered visual components inside a single frameless card with fully resolved wordmark.
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

// ComposeWithState assembles all rendered components with external wordmark animation progress state matching Image 3.
func ComposeWithState(
	s *stats.Summary,
	layout Layout,
	theme Theme,
	graphRows []string,
	monthHeader string,
	statsRows []StatsRow,
	wordmarkState WordmarkState,
) string {
	var verticalBar string
	if theme.Mode == ColorASCII {
		verticalBar = "|"
	} else {
		verticalBar = "│"
	}

	margin := "  "
	if layout.LeftMargin >= 0 {
		margin = strings.Repeat(" ", layout.LeftMargin)
	}

	// If terminal is too narrow to display two columns side by side (< 65 cols), use compact stacked composition
	if layout.TerminalWidth < 65 {
		return composeStacked(s, layout, theme, graphRows, monthHeader, statsRows, wordmarkState)
	}

	// === 1. PREPARE 17 LEFT-SIDE ROWS ===
	leftRows := make([]string, 17)
	leftVisLens := make([]int, 17)

	// Wordmark on rows 0, 1, 2
	wordmarkWidth := layout.GraphWidth
	if wordmarkWidth > 46 {
		wordmarkWidth = 46
	}
	wordmarkLines, wordmarkVisLen := RenderWordmarkWithStateAndWidth(theme, wordmarkState, wordmarkWidth)
	for r := 0; r < 3; r++ {
		if r < len(wordmarkLines) {
			leftRows[r] = wordmarkLines[r]
			leftVisLens[r] = wordmarkVisLen
		}
	}

	// Row 3: Empty
	leftRows[3] = ""
	leftVisLens[3] = 0

	// Row 4: Total Contributions
	numStr := formatNumber(s.Total)
	labelStr := "CONTRIBUTIONS"
	if s.Total == 1 {
		labelStr = "CONTRIBUTION"
	}
	if theme.Mode == ColorASCII {
		leftRows[4] = fmt.Sprintf("%s %s", numStr, labelStr)
	} else {
		leftRows[4] = fmt.Sprintf("%s%s%s%s %s%s%s",
			theme.Bold, theme.Primary, numStr, theme.Reset,
			theme.Secondary, labelStr, theme.Reset,
		)
	}
	leftVisLens[4] = len(numStr) + 1 + len(labelStr)

	// Row 5: Empty
	leftRows[5] = ""
	leftVisLens[5] = 0

	// Row 6: Month Header
	if monthHeader != "" {
		if theme.Mode == ColorASCII {
			leftRows[6] = monthHeader
		} else {
			leftRows[6] = theme.Secondary + monthHeader + theme.Reset
		}
		leftVisLens[6] = utf8.RuneCountInString(monthHeader)
	} else {
		leftRows[6] = ""
		leftVisLens[6] = 0
	}

	// Rows 7..13: 7 Graph Rows (Sun..Sat)
	for w := 0; w < 7; w++ {
		if w < len(graphRows) {
			leftRows[7+w] = graphRows[w]
			leftVisLens[7+w] = layout.GraphWidth
		}
	}

	// Rows 14..16: Empty
	leftRows[14] = ""
	leftVisLens[14] = 0
	leftRows[15] = ""
	leftVisLens[15] = 0
	leftRows[16] = ""
	leftVisLens[16] = 0

	// === 2. ASSEMBLE PARALLEL ROWS WITH VERTICAL DIVIDER ===
	totalRows := 17
	if len(statsRows) > totalRows {
		totalRows = len(statsRows)
	}

	divider := theme.Border + verticalBar + theme.Reset
	gapDivider := "  " + divider + "  "
	if layout.DividerWidth <= 3 {
		gapDivider = " " + divider + " "
	}

	var lines []string
	for i := 0; i < totalRows; i++ {
		var leftContent string
		var leftVis int
		if i < len(leftRows) {
			leftContent = leftRows[i]
			leftVis = leftVisLens[i]
		}
		padLeft := layout.GraphWidth - leftVis
		if padLeft < 0 {
			padLeft = 0
		}
		leftPart := leftContent + strings.Repeat(" ", padLeft)

		var rightPart string
		if i < len(statsRows) {
			rightPart = statsRows[i].Content
		}

		line := margin + leftPart + gapDivider + rightPart
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func composeStacked(
	s *stats.Summary,
	layout Layout,
	theme Theme,
	graphRows []string,
	monthHeader string,
	statsRows []StatsRow,
	wordmarkState WordmarkState,
) string {
	margin := " "
	if layout.LeftMargin > 0 {
		margin = strings.Repeat(" ", layout.LeftMargin)
	}

	var lines []string
	addLine := func(c string) {
		lines = append(lines, margin+c)
	}

	// 1. Wordmark
	availW := layout.TerminalWidth - len(margin)
	wLines, _ := RenderWordmarkWithStateAndWidth(theme, wordmarkState, availW)
	for _, l := range wLines {
		addLine(l)
	}

	// 2. Year Progress
	year := 2026
	if !s.FetchedAt.IsZero() {
		year = s.FetchedAt.Year()
	}
	progressStr, _ := RenderYearProgress(year, CalculateYearProgress(), 8, theme)
	addLine(progressStr)

	// 3. Total
	numStr := formatNumber(s.Total)
	labelStr := "CONTRIBUTIONS"
	if s.Total == 1 {
		labelStr = "CONTRIBUTION"
	}
	if theme.Mode == ColorASCII {
		addLine(fmt.Sprintf("%s %s", numStr, labelStr))
	} else {
		addLine(fmt.Sprintf("%s%s%s%s %s%s%s", theme.Bold, theme.Primary, numStr, theme.Reset, theme.Secondary, labelStr, theme.Reset))
	}

	// 4. Month header + Graph
	if monthHeader != "" {
		if theme.Mode == ColorASCII {
			addLine(monthHeader)
		} else {
			addLine(theme.Secondary + monthHeader + theme.Reset)
		}
	}
	for _, gr := range graphRows {
		addLine(gr)
	}

	// 5. Stats
	for _, sr := range statsRows {
		if sr.Content != "" {
			addLine(sr.Content)
		}
	}

	return strings.Join(lines, "\n")
}
