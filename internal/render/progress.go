package render

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// CalculateYearProgress returns the percentage of the current calendar year elapsed (0-100).
func CalculateYearProgress() int {
	return CalculateYearProgressAt(time.Now())
}

// CalculateYearProgressAt returns the percentage of the year elapsed at a given point in time.
func CalculateYearProgressAt(now time.Time) int {
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

// RenderSegmentedProgressBar builds a segmented, terminal-native progress bar matching the reference design.
func RenderSegmentedProgressBar(percentage, width int, theme Theme) string {
	if width < 1 {
		width = 14
	}

	filled := (percentage * width) / 100
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	unfilled := width - filled

	var bar strings.Builder
	fillChar := "▌"
	emptyChar := "▌"

	if theme.Mode == ColorASCII {
		fillChar = "#"
		emptyChar = "-"
		for i := 0; i < filled; i++ {
			bar.WriteString(fillChar)
		}
		for i := 0; i < unfilled; i++ {
			bar.WriteString(emptyChar)
		}
		return bar.String()
	}

	// Crisp white for active segments, dark slate/charcoal for inactive segments
	activeColor := theme.Bold + "\033[38;2;255;255;255m"
	if theme.Mode == Color256 {
		activeColor = "\033[1;38;5;15m"
	}
	if theme.Theme == ThemeLight {
		activeColor = theme.Bold + theme.Primary
	}

	bar.WriteString(activeColor)
	for i := 0; i < filled; i++ {
		bar.WriteString(fillChar)
	}
	bar.WriteString(theme.Reset)

	bar.WriteString(theme.Border)
	for i := 0; i < unfilled; i++ {
		bar.WriteString(emptyChar)
	}
	bar.WriteString(theme.Reset)

	return bar.String()
}

// RenderYearProgress renders the complete "YEAR [BAR] PERCENTAGE" indicator matching Image 3.
func RenderYearProgress(year, percentage, barWidth int, theme Theme) (string, int) {
	yearStr := fmt.Sprintf("%d", year)
	pctStr := fmt.Sprintf("%d%%", percentage)

	if barWidth < 1 {
		barWidth = 14
	}

	bar := RenderSegmentedProgressBar(percentage, barWidth, theme)

	// Visible length: year (4) + space (1) + barWidth + space (1) + len(pctStr)
	visLen := utf8.RuneCountInString(yearStr) + 1 + barWidth + 1 + utf8.RuneCountInString(pctStr)

	formatted := fmt.Sprintf("%s%s%s%s %s %s%s%s",
		theme.Bold, theme.Primary, yearStr, theme.Reset,
		bar,
		theme.Secondary, pctStr, theme.Reset,
	)

	return formatted, visLen
}

// Backward compatibility helpers
func calculateYearProgress() int {
	return CalculateYearProgress()
}

func RenderYearProgressBar(percentage, width int, theme Theme) string {
	return RenderSegmentedProgressBar(percentage, width, theme)
}

func renderYearProgressBar(percentage, width int, mode ColorMode, theme ThemeMode) string {
	return RenderSegmentedProgressBar(percentage, width, GetTheme(mode, theme))
}
