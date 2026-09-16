package render

import (
	"strings"
	"time"
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

// RenderYearProgressBar builds an ANSI-styled horizontal progress bar for year completion.
func RenderYearProgressBar(percentage, width int, theme Theme) string {
	if width < 1 {
		width = 10
	}

	filled := (percentage * width) / 100
	if filled > width {
		filled = width
	}

	unfilled := width - filled

	var bar strings.Builder
	fillChar := "█"
	emptyChar := "░"

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

	// Colored version
	bar.WriteString(theme.ActivityHigh)
	for i := 0; i < filled; i++ {
		bar.WriteString(fillChar)
	}
	bar.WriteString(theme.Reset)
	bar.WriteString(theme.Muted)
	for i := 0; i < unfilled; i++ {
		bar.WriteString(emptyChar)
	}
	bar.WriteString(theme.Reset)

	return bar.String()
}

// Backward compatibility helpers
func calculateYearProgress() int {
	return CalculateYearProgress()
}

func renderYearProgressBar(percentage, width int, mode ColorMode, theme ThemeMode) string {
	return RenderYearProgressBar(percentage, width, GetTheme(mode, theme))
}
