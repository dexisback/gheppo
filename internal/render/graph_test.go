package render

import (
	"strings"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/stats"
)

func TestNewContributionGraph(t *testing.T) {
	grid := make([][]stats.Cell, 10)
	for w := 0; w < 10; w++ {
		grid[w] = make([]stats.Cell, 7)
	}

	graph := NewContributionGraph(grid, 5)
	if len(graph.Weeks) != 5 {
		t.Fatalf("expected 5 visible weeks, got %d", len(graph.Weeks))
	}
	if graph.VisibleWeeks != 5 {
		t.Errorf("VisibleWeeks = %d, want 5", graph.VisibleWeeks)
	}
	if graph.Width != 10 {
		t.Errorf("Width = %d, want 10", graph.Width)
	}
	if graph.Height != 7 {
		t.Errorf("Height = %d, want 7", graph.Height)
	}
}

func TestRenderContributionGraphASCII(t *testing.T) {
	grid := make([][]stats.Cell, 3)
	for w := 0; w < 3; w++ {
		grid[w] = make([]stats.Cell, 7)
		for d := 0; d < 7; d++ {
			grid[w][d] = stats.Cell{
				Count:  d,
				Bucket: d % 5,
				Empty:  d == 0, // Sunday is empty
			}
		}
	}

	graph := NewContributionGraph(grid, 3)
	theme := GetTheme(ColorASCII, ThemeDark)
	cfg := DefaultGraphConfig()

	rows := RenderContributionGraph(graph, theme, cfg)
	if len(rows) != 7 {
		t.Fatalf("expected 7 weekday rows, got %d", len(rows))
	}

	// First row (Sunday, d=0) should be empty cells "      " (3 weeks * 2 chars = 6 chars)
	if rows[0] != "      " {
		t.Errorf("row[0] = %q, want 6 spaces", rows[0])
	}

	// Second row (Monday, d=1, bucket=1) in ASCII should have glyphs
	if !strings.Contains(rows[1], "░") && !strings.Contains(rows[1], "-") {
		t.Errorf("row[1] = %q, expected level 1 glyph", rows[1])
	}
}

func TestRenderContributionGraphColors(t *testing.T) {
	grid := make([][]stats.Cell, 2)
	for w := 0; w < 2; w++ {
		grid[w] = make([]stats.Cell, 7)
		for d := 0; d < 7; d++ {
			grid[w][d] = stats.Cell{
				Bucket: 3,
			}
		}
	}

	graph := NewContributionGraph(grid, 2)
	theme := GetTheme(ColorTrueColor, ThemeDark)
	cfg := DefaultGraphConfig()

	rows := RenderContributionGraph(graph, theme, cfg)
	if len(rows) != 7 {
		t.Fatalf("expected 7 rows, got %d", len(rows))
	}

	// Should contain the truecolor ANSI code for Level 3 dark theme
	for i, row := range rows {
		if !strings.Contains(row, darkActivityLevel3) {
			t.Errorf("row %d = %q, missing level 3 color", i, row)
		}
	}
}

func TestRenderMonthHeader(t *testing.T) {
	weeks := make([][]stats.Cell, 12)
	baseDate := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)

	for w := 0; w < 12; w++ {
		weeks[w] = make([]stats.Cell, 7)
		for d := 0; d < 7; d++ {
			weeks[w][d] = stats.Cell{
				Date: baseDate.AddDate(0, 0, w*7+d),
			}
		}
	}

	header := RenderMonthHeader(weeks, 24)
	if !strings.Contains(header, "JAN") {
		t.Errorf("month header missing JAN: %q", header)
	}
	if !strings.Contains(header, "FEB") {
		t.Errorf("month header missing FEB: %q", header)
	}
}
