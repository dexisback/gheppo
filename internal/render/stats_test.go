package render

import (
	"strings"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/stats"
)

func TestRenderStatsPanel(t *testing.T) {
	s := &stats.Summary{
		Login:         "dexisback",
		Total:         1000,
		LongestStreak: 42,
		BestDay:       time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		BestDayCount:  30,
		DailyAverage:  4.2,
		Followers:     150,
		Following:     45,
		Repos:         12,
		TotalStars:    88,
	}

	theme := GetTheme(ColorASCII, ThemeDark)
	rows := RenderStatsPanel(s, theme)

	if len(rows) == 0 {
		t.Fatal("expected stats rows, got none")
	}

	joined := ""
	for _, r := range rows {
		joined += r.Content + "\n"
	}

	if !strings.Contains(joined, "42 DAYS") {
		t.Errorf("stats missing streak: %s", joined)
	}
	if !strings.Contains(joined, "LONGEST STREAK") {
		t.Errorf("stats missing streak label: %s", joined)
	}
	if !strings.Contains(joined, "MAY 20 · 30") {
		t.Errorf("stats missing best day: %s", joined)
	}
	if !strings.Contains(joined, "4.2 / DAY") {
		t.Errorf("stats missing daily average: %s", joined)
	}
	if !strings.Contains(joined, "@dexisback") {
		t.Errorf("stats missing username: %s", joined)
	}
	if !strings.Contains(joined, "150") || !strings.Contains(joined, "FOLLOWERS") {
		t.Errorf("stats missing followers: %s", joined)
	}
	if !strings.Contains(joined, "88") || !strings.Contains(joined, "TOTAL STARS") {
		t.Errorf("stats missing stars: %s", joined)
	}
}

func TestRenderStatsPanelNil(t *testing.T) {
	theme := GetTheme(ColorASCII, ThemeDark)
	rows := RenderStatsPanel(nil, theme)
	if rows != nil {
		t.Errorf("expected nil rows for nil summary, got %v", rows)
	}
}
