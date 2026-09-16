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

func TestRenderContributionGraphNilAndEmpty(t *testing.T) {
	theme := GetTheme(ColorASCII, ThemeDark)
	cfg := DefaultGraphConfig()

	// Empty graph
	emptyGraph := NewContributionGraph(nil, 0)
	rows := RenderContributionGraph(emptyGraph, theme, cfg)
	if len(rows) != 7 {
		t.Fatalf("expected 7 rows for empty graph, got %d", len(rows))
	}
	for i, r := range rows {
		if r != "" {
			t.Errorf("row %d = %q, want empty string", i, r)
		}
	}
}

func TestRenderContributionGraphAllBuckets(t *testing.T) {
	// 1 week with buckets 0 to 4 across the first 5 weekdays
	grid := make([][]stats.Cell, 1)
	grid[0] = make([]stats.Cell, 7)
	for d := 0; d <= 4; d++ {
		grid[0][d] = stats.Cell{
			Bucket: d,
		}
	}

	graph := NewContributionGraph(grid, 1)
	theme := GetTheme(ColorTrueColor, ThemeDark)
	cfg := DefaultGraphConfig()

	rows := RenderContributionGraph(graph, theme, cfg)
	if len(rows) != 7 {
		t.Fatalf("expected 7 rows, got %d", len(rows))
	}

	// Verify bucket 0 uses darkEmptyCell
	if !strings.Contains(rows[0], darkEmptyCell) {
		t.Errorf("bucket 0 missing darkEmptyCell: %q", rows[0])
	}
	// Verify bucket 1 uses darkActivityLevel1
	if !strings.Contains(rows[1], darkActivityLevel1) {
		t.Errorf("bucket 1 missing darkActivityLevel1: %q", rows[1])
	}
	// Verify bucket 2 uses darkActivityLevel2
	if !strings.Contains(rows[2], darkActivityLevel2) {
		t.Errorf("bucket 2 missing darkActivityLevel2: %q", rows[2])
	}
	// Verify bucket 3 uses darkActivityLevel3
	if !strings.Contains(rows[3], darkActivityLevel3) {
		t.Errorf("bucket 3 missing darkActivityLevel3: %q", rows[3])
	}
	// Verify bucket 4 uses darkActivityLevel4
	if !strings.Contains(rows[4], darkActivityLevel4) {
		t.Errorf("bucket 4 missing darkActivityLevel4: %q", rows[4])
	}
}

func TestRenderContributionGraph256Color(t *testing.T) {
	grid := make([][]stats.Cell, 1)
	grid[0] = make([]stats.Cell, 7)
	for d := 0; d < 7; d++ {
		grid[0][d] = stats.Cell{
			Bucket: 4,
		}
	}

	graph := NewContributionGraph(grid, 1)
	theme := GetTheme(Color256, ThemeDark)
	cfg := DefaultGraphConfig()

	rows := RenderContributionGraph(graph, theme, cfg)
	if len(rows) != 7 {
		t.Fatalf("expected 7 rows, got %d", len(rows))
	}

	for i, r := range rows {
		if !strings.Contains(r, "\033[38;5;120m") {
			t.Errorf("row %d = %q, missing 256-color level 4", i, r)
		}
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
				Empty:  d == 0, // Sunday is empty padding
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

	// First row (Sunday, d=0) should be empty padding cells "      " (3 weeks * 2 chars = 6 chars)
	if rows[0] != "      " {
		t.Errorf("row[0] = %q, want 6 spaces", rows[0])
	}

	// Second row (Monday, d=1, bucket=1) in ASCII should have glyphs
	if !strings.Contains(rows[1], "░") && !strings.Contains(rows[1], "-") {
		t.Errorf("row[1] = %q, expected level 1 glyph", rows[1])
	}
}

func TestRenderContributionGraphNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	grid := make([][]stats.Cell, 1)
	grid[0] = make([]stats.Cell, 7)
	for d := 0; d < 5; d++ {
		grid[0][d] = stats.Cell{
			Bucket: d,
		}
	}

	mode := DetectColorMode()
	if mode != ColorASCII {
		t.Fatalf("expected ColorASCII with NO_COLOR, got %v", mode)
	}

	theme := GetTheme(mode, ThemeDark)
	cfg := DefaultGraphConfig()
	graph := NewContributionGraph(grid, 1)

	rows := RenderContributionGraph(graph, theme, cfg)
	// Pure ascii glyphs should be used (. - = + #)
	if !strings.Contains(rows[0], ".") {
		t.Errorf("row 0 missing '.' for bucket 0: %q", rows[0])
	}
	if !strings.Contains(rows[1], "-") {
		t.Errorf("row 1 missing '-' for bucket 1: %q", rows[1])
	}
	if !strings.Contains(rows[4], "#") {
		t.Errorf("row 4 missing '#' for bucket 4: %q", rows[4])
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
	if !strings.Contains(header, "MAR") {
		t.Errorf("month header missing MAR: %q", header)
	}
}

func TestRenderMonthHeaderWidthBound(t *testing.T) {
	weeks := make([][]stats.Cell, 20)
	baseDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for w := 0; w < 20; w++ {
		weeks[w] = make([]stats.Cell, 7)
		for d := 0; d < 7; d++ {
			weeks[w][d] = stats.Cell{
				Date: baseDate.AddDate(0, 0, w*7+d),
			}
		}
	}

	maxGraphWidth := 20
	header := RenderMonthHeader(weeks, maxGraphWidth)
	if len(header) > maxGraphWidth {
		t.Errorf("month header length %d exceeds max graph width %d", len(header), maxGraphWidth)
	}
}
