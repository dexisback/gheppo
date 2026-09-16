package render

import (
	"strings"
	"testing"
	"time"
)

func TestCalculateYearProgressAt(t *testing.T) {
	// Mid-year: July 2 ~ 50%
	midYear := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	progress := CalculateYearProgressAt(midYear)
	if progress < 48 || progress > 52 {
		t.Errorf("mid-year progress = %d, expected ~50", progress)
	}

	// Start of year: Jan 1 00:00 -> 0%
	startYear := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	progressStart := CalculateYearProgressAt(startYear)
	if progressStart != 0 {
		t.Errorf("start-year progress = %d, expected 0", progressStart)
	}
}

func TestRenderYearProgressBar(t *testing.T) {
	themeASCII := GetTheme(ColorASCII, ThemeDark)
	barASCII := RenderYearProgressBar(50, 10, themeASCII)
	if barASCII != "#####-----" {
		t.Errorf("ASCII 50%% 10-char bar = %q, want '#####-----'", barASCII)
	}

	themeTC := GetTheme(ColorTrueColor, ThemeDark)
	barTC := RenderYearProgressBar(50, 10, themeTC)
	if !strings.Contains(barTC, "█████") || !strings.Contains(barTC, "░░░░░") {
		t.Errorf("TrueColor 50%% 10-char bar = %q, want 5 filled blocks and 5 empty blocks", barTC)
	}
}
