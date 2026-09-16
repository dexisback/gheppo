package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
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

	// 1. Prepare visual components
	graph := NewContributionGraph(s.Grid, layout.VisibleWeeks)
	graphConfig := DefaultGraphConfig()
	graphRows := RenderContributionGraph(graph, theme, graphConfig)
	monthHeader := RenderMonthHeader(graph.Weeks, layout.GraphWidth)
	statsRows := RenderStatsPanelWithWidth(s, theme, layout.StatsWidth)

	// 2. Compose into explicit rectangular regions matching Image 3
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
	// If terminal is too narrow to display two columns side by side (< 65 cols), use compact stacked composition
	if layout.TerminalWidth < 65 {
		return composeStacked(s, layout, theme, graphRows, monthHeader, statsRows, wordmarkState)
	}

	var verticalBar string
	if theme.Mode == ColorASCII {
		verticalBar = "|"
	} else {
		verticalBar = "│"
	}

	// === 1. BUILD LEFT RECTANGULAR REGION (16 ROWS) ===
	totalRows := len(statsRows)
	if totalRows < 16 {
		totalRows = 16
	}
	leftLines := make([]string, totalRows)

	// Rows 0..2: Wordmark
	wordmarkWidth := layout.GraphWidth
	if wordmarkWidth > 46 {
		wordmarkWidth = 46
	}
	wordmarkLines, wordmarkVisLen := RenderWordmarkWithStateAndWidth(theme, wordmarkState, wordmarkWidth)
	for r := 0; r < 3; r++ {
		if r < len(wordmarkLines) {
			pad := layout.GraphWidth - wordmarkVisLen
			if pad < 0 {
				pad = 0
			}
			leftLines[r] = wordmarkLines[r] + strings.Repeat(" ", pad)
		} else {
			leftLines[r] = strings.Repeat(" ", layout.GraphWidth)
		}
	}

	// Row 3: Empty
	leftLines[3] = strings.Repeat(" ", layout.GraphWidth)

	// Row 4: Total Contributions (left) + Username (right) in left region
	numStr := formatNumber(s.Total)
	labelStr := "CONTRIBUTIONS"
	if s.Total == 1 {
		labelStr = "CONTRIBUTION"
	}
	var totalFormatted string
	if theme.Mode == ColorASCII {
		totalFormatted = fmt.Sprintf("%s %s", numStr, labelStr)
	} else {
		totalFormatted = fmt.Sprintf("%s%s%s%s %s%s%s",
			theme.Bold, theme.Primary, numStr, theme.Reset,
			theme.Secondary, labelStr, theme.Reset,
		)
	}
	totalVisLen := len(numStr) + 1 + len(labelStr)

	username := "@" + s.Login
	var userFormatted string
	if theme.Mode == ColorASCII {
		userFormatted = username
	} else {
		userFormatted = fmt.Sprintf("%s%s%s%s", theme.Primary, theme.Bold, username, theme.Reset)
	}
	userVisLen := len(username)

	padTotal := layout.GraphWidth - totalVisLen - userVisLen
	if padTotal < 2 {
		padTotal = 2
	}
	leftRow4 := totalFormatted + strings.Repeat(" ", padTotal) + userFormatted
	remPad := layout.GraphWidth - (totalVisLen + padTotal + userVisLen)
	if remPad > 0 {
		leftRow4 += strings.Repeat(" ", remPad)
	}
	leftLines[4] = leftRow4

	// Rows 5..7: Empty lines for vertical breathing room
	for r := 5; r <= 7; r++ {
		leftLines[r] = strings.Repeat(" ", layout.GraphWidth)
	}

	// Row 8: Month Header
	if monthHeader != "" {
		var mhFormatted string
		if theme.Mode == ColorASCII {
			mhFormatted = monthHeader
		} else {
			mhFormatted = theme.Secondary + monthHeader + theme.Reset
		}
		padMH := layout.GraphWidth - utf8.RuneCountInString(monthHeader)
		if padMH < 0 {
			padMH = 0
		}
		leftLines[8] = mhFormatted + strings.Repeat(" ", padMH)
	} else {
		leftLines[8] = strings.Repeat(" ", layout.GraphWidth)
	}

	// Rows 9..15: Graph Rows (7 weekday rows, terminating at row 15)
	for w := 0; w < 7; w++ {
		targetRow := 9 + w
		if targetRow < totalRows {
			if w < len(graphRows) {
				leftLines[targetRow] = graphRows[w]
			} else {
				leftLines[targetRow] = strings.Repeat(" ", layout.GraphWidth)
			}
		}
	}

	leftRegion := strings.Join(leftLines, "\n")

	// === 3. BUILD RIGHT RECTANGULAR REGION ===
	if len(statsRows) > totalRows {
		totalRows = len(statsRows)
	}

	// === 2. BUILD DIVIDER REGION ===
	dividerCell := theme.Border + verticalBar + theme.Reset
	gapDivider := "  " + dividerCell + "  "
	if layout.DividerWidth <= 3 {
		gapDivider = " " + dividerCell + " "
	}
	dividerLines := make([]string, totalRows)
	for i := 0; i < totalRows; i++ {
		dividerLines[i] = gapDivider
	}
	dividerRegion := strings.Join(dividerLines, "\n")

	rightLines := make([]string, totalRows)
	for i := 0; i < totalRows; i++ {
		if i < len(statsRows) {
			rightLines[i] = statsRows[i].Content
		} else {
			rightLines[i] = ""
		}
	}
	rightRegion := strings.Join(rightLines, "\n")

	// === 4. JOIN REGIONS HORIZONTALLY WITH LIP GLOSS ===
	joined := lipgloss.JoinHorizontal(lipgloss.Top, leftRegion, dividerRegion, rightRegion)

	// === 5. APPLY LEFT MARGIN ===
	if layout.LeftMargin > 0 {
		marginStr := strings.Repeat(" ", layout.LeftMargin)
		lines := strings.Split(joined, "\n")
		for i := range lines {
			lines[i] = marginStr + lines[i]
		}
		joined = strings.Join(lines, "\n")
	}

	return "\n" + joined
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

	var sections []string

	// 1. Wordmark
	availW := layout.TerminalWidth - len(margin)
	wLines, _ := RenderWordmarkWithStateAndWidth(theme, wordmarkState, availW)
	sections = append(sections, strings.Join(wLines, "\n"))

	// 2. Year Progress
	year := 2026
	if !s.FetchedAt.IsZero() {
		year = s.FetchedAt.Year()
	}
	progressStr, _ := RenderYearProgress(year, CalculateYearProgress(), 8, theme)
	sections = append(sections, progressStr)

	// 3. Total + Username
	numStr := formatNumber(s.Total)
	labelStr := "CONTRIBUTIONS"
	if s.Total == 1 {
		labelStr = "CONTRIBUTION"
	}
	username := "@" + s.Login
	if theme.Mode == ColorASCII {
		sections = append(sections, fmt.Sprintf("%s %s   %s", numStr, labelStr, username))
	} else {
		sections = append(sections, fmt.Sprintf("%s%s%s%s %s%s%s   %s%s%s%s",
			theme.Bold, theme.Primary, numStr, theme.Reset,
			theme.Secondary, labelStr, theme.Reset,
			theme.Primary, theme.Bold, username, theme.Reset,
		))
	}

	// 4. Month header + Graph
	var graphLines []string
	if monthHeader != "" {
		if theme.Mode == ColorASCII {
			graphLines = append(graphLines, monthHeader)
		} else {
			graphLines = append(graphLines, theme.Secondary+monthHeader+theme.Reset)
		}
	}
	graphLines = append(graphLines, graphRows...)
	sections = append(sections, strings.Join(graphLines, "\n"))

	// 5. Stats
	var statsLines []string
	for _, sr := range statsRows {
		if sr.Content != "" {
			statsLines = append(statsLines, sr.Content)
		}
	}
	sections = append(sections, strings.Join(statsLines, "\n"))

	joined := lipgloss.JoinVertical(lipgloss.Left, sections...)

	if len(margin) > 0 {
		lines := strings.Split(joined, "\n")
		for i := range lines {
			lines[i] = margin + lines[i]
		}
		joined = strings.Join(lines, "\n")
	}

	return "\n" + joined
}
