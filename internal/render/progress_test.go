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
	if !strings.Contains(barTC, "▌▌▌▌▌") {
		t.Errorf("TrueColor 50%% 10-char bar = %q, want 5 filled slats and 5 empty slats", barTC)
	}
}

func TestRenderYearProgress(t *testing.T) {
	theme := GetTheme(ColorASCII, ThemeDark)
	formatted, visLen := RenderYearProgress(2026, 70, 15, theme)

	if !strings.Contains(formatted, "2026") {
		t.Errorf("missing year in progress: %q", formatted)
	}
	if !strings.Contains(formatted, "70%") {
		t.Errorf("missing percentage in progress: %q", formatted)
	}
	expectedLen := 4 + 1 + 15 + 1 + 3 // 2026 (4) + space (1) + bar (15) + space (1) + 70% (3) = 24
	if visLen != expectedLen {
		t.Errorf("visLen = %d, want %d", visLen, expectedLen)
	}
}

