package render

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dexisback/gheppo/internal/stats"
)

// Grid renders a complete, responsive contribution card with stats panel.
func Grid(s *stats.Summary) string {
	if s == nil {
		return ""
	}

	mode := DetectColorMode()
	theme := DetectTheme()
	termWidth := GetTerminalWidth()

	totalWeeks := 0
	if s.Grid != nil {
		totalWeeks = len(s.Grid)
	}

	layout := ComputeLayout(termWidth, totalWeeks)
	return RenderCard(s, layout, mode, theme)
}

// RenderCard builds the full formatted card string according to layout and styling.
func RenderCard(s *stats.Summary, layout LayoutConfig, mode ColorMode, theme ThemeMode) string {
	if s == nil {
		return ""
	}

	tokens := GetTokens(mode, theme)

	// Border characters
	var (
		cornerTL, cornerTR, cornerBL, cornerBR string
		horizontalBar, verticalBar             string
	)

	if mode == ColorASCII {
		cornerTL, cornerTR, cornerBL, cornerBR = "+", "+", "+", "+"
		horizontalBar, verticalBar = "-", "|"
	} else {
		// Use simple box drawing - clean, not rounded
		cornerTL, cornerTR, cornerBL, cornerBR = "┌", "┐", "└", "┘"
		horizontalBar, verticalBar = "─", "│"
	}

	innerWidth := layout.CardWidth - 4 // space between "│ " and " │"
	if innerWidth < 20 {
		innerWidth = 20
	}

	// Sliced weeks for responsive graph
	var visibleWeeks [][]stats.Cell
	if len(s.Grid) > 0 {
		start := len(s.Grid) - layout.VisibleWeeks
		if start < 0 {
			start = 0
		}
		visibleWeeks = s.Grid[start:]
	}

	var lines []string

	// Helper to add framed line
	addLine := func(content string, visibleLen int) {
		pad := innerWidth - visibleLen
		if pad < 0 {
			pad = 0
		}
		var b strings.Builder
		if layout.LeftMargin > 0 {
			b.WriteString(strings.Repeat(" ", layout.LeftMargin))
		}
		b.WriteString(tokens.Border)
		b.WriteString(verticalBar)
		b.WriteString(" ")
		b.WriteString(tokens.Reset)
		b.WriteString(content)
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(tokens.Border)
		b.WriteString(" ")
		b.WriteString(verticalBar)
		b.WriteString(tokens.Reset)
		lines = append(lines, b.String())
	}

	addEmptyLine := func() {
		addLine("", 0)
	}

	// === TOP BORDER ===
	topBorder := fmt.Sprintf("%s%s%s%s%s",
		tokens.Border,
		cornerTL,
		strings.Repeat(horizontalBar, layout.CardWidth-2),
		cornerTR,
		tokens.Reset,
	)
	if layout.LeftMargin > 0 {
		topBorder = strings.Repeat(" ", layout.LeftMargin) + topBorder
	}
	lines = append(lines, topBorder)

	addEmptyLine()

	// === WORDMARK + YEAR ===
	wordmark := GetWordmark(mode)
	year := time.Now().Year()
	if !s.FetchedAt.IsZero() {
		year = s.FetchedAt.Year()
	}
	yearStr := fmt.Sprintf("%d", year)

	headerVisLen := utf8.RuneCountInString(wordmark) + utf8.RuneCountInString(yearStr)
	if headerVisLen <= innerWidth {
		spacing := innerWidth - headerVisLen
		headerContent := fmt.Sprintf("%s%s%s%s%s%s%s",
			tokens.Primary, wordmark, tokens.Reset,
			strings.Repeat(" ", spacing),
			tokens.Secondary, yearStr, tokens.Reset,
		)
		addLine(headerContent, innerWidth)
	} else {
		addLine(fmt.Sprintf("%s%s%s", tokens.Primary, wordmark, tokens.Reset), utf8.RuneCountInString(wordmark))
		addLine(fmt.Sprintf("%s%s%s", tokens.Secondary, yearStr, tokens.Reset), utf8.RuneCountInString(yearStr))
	}

	// === TOTAL CONTRIBUTIONS + YEAR PROGRESS BAR ===
	var totalStr string
	if s.Total == 1 {
		totalStr = "1 CONTRIBUTION"
	} else {
		totalStr = fmt.Sprintf("%s CONTRIBUTIONS", formatNumber(s.Total))
	}

	// Year progress percentage
	yearProgress := calculateYearProgress()
	progressBar := renderYearProgressBar(yearProgress, 15, mode, theme)
	progressStr := fmt.Sprintf("%s %d%%", progressBar, yearProgress)

	// Account for ANSI codes not being visible
	actualProgressLen := 15 + 1 + len(fmt.Sprintf("%d%%", yearProgress))

	if utf8.RuneCountInString(totalStr)+actualProgressLen <= innerWidth {
		spacing := innerWidth - utf8.RuneCountInString(totalStr) - actualProgressLen
		subHeaderContent := fmt.Sprintf("%s%s%s%s%s%s",
			tokens.Primary, totalStr, tokens.Reset,
			strings.Repeat(" ", spacing),
			progressStr, tokens.Reset,
		)
		addLine(subHeaderContent, innerWidth)
	} else {
		addLine(fmt.Sprintf("%s%s%s", tokens.Primary, totalStr, tokens.Reset), utf8.RuneCountInString(totalStr))
		addLine(progressStr+tokens.Reset, actualProgressLen)
	}

	addEmptyLine()

	// === MONTH LABELS ===
	monthHeader := BuildMonthHeader(visibleWeeks, layout.GraphWidth)
	if monthHeader != "" {
		coloredMonth := fmt.Sprintf("%s%s%s", tokens.Secondary, monthHeader, tokens.Reset)
		addLine(coloredMonth, utf8.RuneCountInString(monthHeader))
	}

	addEmptyLine()

	// === TWO-COLUMN LAYOUT: GRAPH | STATS ===
	// We need to render the graph and stats side by side with a vertical divider

	// Build contribution graph rows (7 rows)
	graphRows := make([]string, 7)
	for weekday := 0; weekday < 7; weekday++ {
		var row strings.Builder
		for _, week := range visibleWeeks {
			if weekday >= len(week) {
				row.WriteString("  ")
				continue
			}

			cell := week[weekday]
			if cell.Empty {
				row.WriteString("  ")
				continue
			}

			row.WriteString(CellColor(mode, theme, cell.Bucket))
			row.WriteByte(' ')
		}
		graphRows[weekday] = row.String()
	}

	// Build stats rows
	statsRows := buildStatsRows(s, tokens, mode)

	// Combine graph and stats with vertical divider
	maxRows := 7
	if len(statsRows) > maxRows {
		maxRows = len(statsRows)
	}

	divider := tokens.Border + verticalBar + tokens.Reset

	for i := 0; i < maxRows; i++ {
		var graphPart string
		if i < len(graphRows) {
			graphPart = graphRows[i]
		} else {
			graphPart = strings.Repeat(" ", layout.GraphWidth)
		}

		var statsPart string
		if i < len(statsRows) {
			statsPart = statsRows[i].content
		} else {
			statsPart = ""
		}

		// Calculate spacing
		graphVisLen := layout.GraphWidth
		statsVisLen := 0
		if i < len(statsRows) {
			statsVisLen = statsRows[i].visibleLen
		}

		// Total available width minus graph minus divider (3 chars: " │ ")
		statsAreaWidth := innerWidth - graphVisLen - 3
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
		tokens.Border,
		cornerBL,
		strings.Repeat(horizontalBar, layout.CardWidth-2),
		cornerBR,
		tokens.Reset,
	)
	if layout.LeftMargin > 0 {
		bottomBorder = strings.Repeat(" ", layout.LeftMargin) + bottomBorder
	}
	lines = append(lines, bottomBorder)

	return strings.Join(lines, "\n")
}

type statsRow struct {
	content    string
	visibleLen int
}

func buildStatsRows(s *stats.Summary, tokens PaletteTokens, mode ColorMode) []statsRow {
	rows := []statsRow{}

	// === ACTIVITY SECTION ===
	rows = append(rows, statsRow{
		content:    tokens.Muted + "ACTIVITY" + tokens.Reset,
		visibleLen: 8,
	})
	rows = append(rows, statsRow{content: "", visibleLen: 0})

	// Longest streak
	streakStr := fmt.Sprintf("%s", formatNumber(s.LongestStreak))
	streakLabel := "LONGEST STREAK"
	if s.LongestStreak == 1 {
		streakLabel = "DAY"
	} else {
		streakLabel = "DAYS"
	}
	rows = append(rows, statsRow{
		content:    tokens.ActivityHigh + tokens.Bold + streakStr + tokens.Reset + " " + tokens.Secondary + streakLabel + tokens.Reset,
		visibleLen: len(streakStr) + 1 + len(streakLabel),
	})
	rows = append(rows, statsRow{
		content:    tokens.Secondary + "LONGEST STREAK" + tokens.Reset,
		visibleLen: 14,
	})
	rows = append(rows, statsRow{content: "", visibleLen: 0})

	// Best day
	if !s.BestDay.IsZero() && s.BestDayCount > 0 {
		bestDayStr := s.BestDay.Format("JAN 02")
		bestDayStr = strings.ToUpper(bestDayStr)
		bestStr := fmt.Sprintf("%s · %s", bestDayStr, formatNumber(s.BestDayCount))
		rows = append(rows, statsRow{
			content:    tokens.Primary + tokens.Bold + bestStr + tokens.Reset,
			visibleLen: len(bestStr),
		})
		rows = append(rows, statsRow{
			content:    tokens.Secondary + "BEST DAY" + tokens.Reset,
			visibleLen: 8,
		})
		rows = append(rows, statsRow{content: "", visibleLen: 0})
	}

	// Daily average
	if s.DailyAverage > 0 {
		avgStr := fmt.Sprintf("%s / DAY", formatFloat(s.DailyAverage))
		rows = append(rows, statsRow{
			content:    tokens.Primary + avgStr + tokens.Reset,
			visibleLen: len(avgStr),
		})
		rows = append(rows, statsRow{
			content:    tokens.Secondary + "DAILY AVERAGE" + tokens.Reset,
			visibleLen: 13,
		})
		rows = append(rows, statsRow{content: "", visibleLen: 0})
	}

	// === PROFILE SECTION ===
	rows = append(rows, statsRow{
		content:    tokens.Muted + "PROFILE" + tokens.Reset,
		visibleLen: 7,
	})
	rows = append(rows, statsRow{content: "", visibleLen: 0})

	// Username
	username := "@" + s.Login
	rows = append(rows, statsRow{
		content:    tokens.Primary + tokens.Bold + username + tokens.Reset,
		visibleLen: len(username),
	})
	rows = append(rows, statsRow{content: "", visibleLen: 0})

	// Followers
	if s.Followers > 0 {
		followersStr := fmt.Sprintf("%s FOLLOWERS", formatNumber(s.Followers))
		rows = append(rows, statsRow{
			content:    tokens.Primary + followersStr + tokens.Reset,
			visibleLen: len(followersStr),
		})
	}

	// Following
	if s.Following > 0 {
		followingStr := fmt.Sprintf("%s FOLLOWING", formatNumber(s.Following))
		rows = append(rows, statsRow{
			content:    tokens.Secondary + followingStr + tokens.Reset,
			visibleLen: len(followingStr),
		})
	}

	// Repositories
	if s.Repos > 0 {
		reposStr := fmt.Sprintf("%s REPOSITORIES", formatNumber(s.Repos))
		rows = append(rows, statsRow{
			content:    tokens.Secondary + reposStr + tokens.Reset,
			visibleLen: len(reposStr),
		})
	}

	// Total stars
	if s.TotalStars > 0 {
		starsStr := fmt.Sprintf("★ %s TOTAL STARS", formatNumber(s.TotalStars))
		rows = append(rows, statsRow{
			content:    tokens.ActivityHigh + starsStr + tokens.Reset,
			visibleLen: len(starsStr),
		})
	}

	return rows
}

func calculateYearProgress() int {
	now := time.Now()
	year := now.Year()
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
	endOfYear := time.Date(year+1, 1, 1, 0, 0, 0, 0, now.Location())

	totalDuration := endOfYear.Sub(startOfYear)
	elapsed := now.Sub(startOfYear)

	percentage := int((float64(elapsed) / float64(totalDuration)) * 100)
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	return percentage
}

func renderYearProgressBar(percentage, width int, mode ColorMode, theme ThemeMode) string {
	if width < 1 {
		width = 10
	}

	filled := (percentage * width) / 100
	if filled > width {
		filled = width
	}

	unfilled := width - filled

	tokens := GetTokens(mode, theme)

	var bar strings.Builder

	// Use block characters for the bar
	fillChar := "█"
	emptyChar := "░"

	if mode == ColorASCII {
		fillChar = "#"
		emptyChar = "-"
		// No colors in ASCII mode
		for i := 0; i < filled; i++ {
			bar.WriteString(fillChar)
		}
		for i := 0; i < unfilled; i++ {
			bar.WriteString(emptyChar)
		}
		return bar.String()
	}

	// Colored version
	bar.WriteString(tokens.ActivityHigh)
	for i := 0; i < filled; i++ {
		bar.WriteString(fillChar)
	}
	bar.WriteString(tokens.Reset)
	bar.WriteString(tokens.Muted)
	for i := 0; i < unfilled; i++ {
		bar.WriteString(emptyChar)
	}
	bar.WriteString(tokens.Reset)

	return bar.String()
}
