package render

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dexisback/gheppo/internal/stats"
)

// Grid renders a complete, responsive retro-terminal contribution card.
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
		cornerTL, cornerTR, cornerBL, cornerBR = "┌", "┐", "└", "┘"
		horizontalBar, verticalBar = "─", "│"
	}

	innerWidth := layout.CardWidth - 4 // inside width between "│ " and " │"
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

	// 1. Top border
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

	// 2. Padding
	addEmptyLine()

	// 3. Header: Wordmark (left) + Year (right)
	wordmark := GetWordmark(mode)
	year := time.Now().Year()
	if !s.FetchedAt.IsZero() {
		year = s.FetchedAt.Year()
	}
	yearStr := fmt.Sprintf("%d", year)

	headerVisLen := utf8.RuneCountInString(wordmark) + utf8.RuneCountInString(yearStr)
	if headerVisLen <= innerWidth {
		spacing := innerWidth - headerVisLen
		headerContent := fmt.Sprintf("%s%s%s%s%s%s%s%s",
			tokens.Wordmark, tokens.Bold, wordmark, tokens.Reset,
			strings.Repeat(" ", spacing),
			tokens.Secondary, tokens.Bold, yearStr,
		)
		addLine(headerContent+tokens.Reset, innerWidth)
	} else {
		addLine(fmt.Sprintf("%s%s%s%s", tokens.Wordmark, tokens.Bold, wordmark, tokens.Reset), utf8.RuneCountInString(wordmark))
		addLine(fmt.Sprintf("%s%s%s%s", tokens.Secondary, tokens.Bold, yearStr, tokens.Reset), utf8.RuneCountInString(yearStr))
	}

	// 4. Subheader: Total contributions (left) + User anchor (right)
	var totalStr string
	if s.Total == 1 {
		totalStr = "1 CONTRIBUTION"
	} else {
		totalStr = fmt.Sprintf("%s CONTRIBUTIONS", formatNumber(s.Total))
	}

	username := s.Login
	if username == "" {
		username = "user"
	}
	identityStr := "@" + username

	subHeaderVisLen := utf8.RuneCountInString(totalStr) + utf8.RuneCountInString(identityStr)
	if subHeaderVisLen <= innerWidth {
		spacing := innerWidth - subHeaderVisLen
		subHeaderContent := fmt.Sprintf("%s%s%s%s%s%s%s%s",
			tokens.Primary, tokens.Bold, totalStr, tokens.Reset,
			strings.Repeat(" ", spacing),
			tokens.Muted, tokens.Bold, identityStr,
		)
		addLine(subHeaderContent+tokens.Reset, innerWidth)
	} else {
		addLine(fmt.Sprintf("%s%s%s%s", tokens.Primary, tokens.Bold, totalStr, tokens.Reset), utf8.RuneCountInString(totalStr))
		addLine(fmt.Sprintf("%s%s%s%s", tokens.Muted, tokens.Bold, identityStr, tokens.Reset), utf8.RuneCountInString(identityStr))
	}

	addEmptyLine()

	// 5. Month Labels
	monthHeader := BuildMonthHeader(visibleWeeks, layout.GridWidth)
	if monthHeader != "" {
		coloredMonth := fmt.Sprintf("%s%s%s", tokens.Muted, monthHeader, tokens.Reset)
		addLine(coloredMonth, utf8.RuneCountInString(monthHeader))
	}

	// 6. Contribution Grid (7 rows, no weekday labels)
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

		gridRowStr := row.String()
		addLine(gridRowStr, layout.GridWidth)
	}

	addEmptyLine()

	// 7. Stats Footer (Current Streak + Longest Streak)
	streakNumStr := formatNumber(s.CurrentStreak)
	longestNumStr := formatNumber(s.LongestStreak)

	currentStreakText := fmt.Sprintf("%s DAY STREAK", streakNumStr)
	longestStreakText := fmt.Sprintf("%s LONGEST", longestNumStr)

	statsVisLen := len(currentStreakText) + len(longestStreakText)
	if statsVisLen <= innerWidth {
		spacing := innerWidth - statsVisLen
		leftPart := fmt.Sprintf("%s%s%s%s %s%s%s",
			tokens.Accent, tokens.Bold, streakNumStr, tokens.Reset,
			tokens.Secondary, "DAY STREAK", tokens.Reset,
		)
		rightPart := fmt.Sprintf("%s%s%s%s %s%s%s",
			tokens.Accent, tokens.Bold, longestNumStr, tokens.Reset,
			tokens.Secondary, "LONGEST", tokens.Reset,
		)
		statsContent := leftPart + strings.Repeat(" ", spacing) + rightPart
		addLine(statsContent, innerWidth)
	} else {
		// Stacked stats for narrow cards
		stat1 := fmt.Sprintf("%s%s%s%s %s%s%s", tokens.Accent, tokens.Bold, streakNumStr, tokens.Reset, tokens.Secondary, "DAY STREAK", tokens.Reset)
		stat2 := fmt.Sprintf("%s%s%s%s %s%s%s", tokens.Accent, tokens.Bold, longestNumStr, tokens.Reset, tokens.Secondary, "LONGEST", tokens.Reset)
		addLine(stat1, utf8.RuneCountInString(currentStreakText))
		addLine(stat2, utf8.RuneCountInString(longestStreakText))
	}

	addEmptyLine()

	// 8. Bottom border
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
