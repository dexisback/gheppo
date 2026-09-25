package render

import (
	"strings"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/config"
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

	// Verify bucket 0 uses theme.EmptyCell
	if !strings.Contains(rows[0], theme.EmptyCell) {
		t.Errorf("bucket 0 missing theme.EmptyCell: %q", rows[0])
	}
	// Verify bucket 1 uses theme.ContributionLevel1
	if !strings.Contains(rows[1], theme.ContributionLevel1) {
		t.Errorf("bucket 1 missing theme.ContributionLevel1: %q", rows[1])
	}
	// Verify bucket 2 uses theme.ContributionLevel2
	if !strings.Contains(rows[2], theme.ContributionLevel2) {
		t.Errorf("bucket 2 missing theme.ContributionLevel2: %q", rows[2])
	}
	// Verify bucket 3 uses theme.ContributionLevel3
	if !strings.Contains(rows[3], theme.ContributionLevel3) {
		t.Errorf("bucket 3 missing theme.ContributionLevel3: %q", rows[3])
	}
	// Verify bucket 4 uses theme.ContributionLevel4
	if !strings.Contains(rows[4], theme.ContributionLevel4) {
		t.Errorf("bucket 4 missing theme.ContributionLevel4: %q", rows[4])
	}
}

func TestRenderContributionGraphZeroActivityAcrossAllThemes(t *testing.T) {
	themes := config.ListThemes()
	cfg := DefaultGraphConfig()

	for _, cfgTheme := range themes {
		t.Run(cfgTheme.Name, func(t *testing.T) {
			// Create a 2-week grid of zero-activity days (bucket 0, empty: false)
			grid := make([][]stats.Cell, 2)
			for w := 0; w < 2; w++ {
				grid[w] = make([]stats.Cell, 7)
				for d := 0; d < 7; d++ {
					grid[w][d] = stats.Cell{
						Bucket: 0,
						Empty:  false,
					}
				}
			}

			graph := NewContributionGraph(grid, 2)

			// 1. TrueColor Mode: zero-activity cells must contain theme's EmptyCell escape
			tcTheme := ResolveTheme(cfgTheme, ColorTrueColor)
			tcRows := RenderContributionGraph(graph, tcTheme, cfg)
			for rIdx, row := range tcRows {
				if !strings.Contains(row, tcTheme.EmptyCell) {
					t.Errorf("theme %q row %d missing TrueColor EmptyCell escape (%q): %q",
						cfgTheme.Name, rIdx, tcTheme.EmptyCell, row)
				}
				if !strings.Contains(row, "■") {
					t.Errorf("theme %q row %d missing block glyph: %q", cfgTheme.Name, rIdx, row)
				}
				// Empty cells should NOT be rendered with Level1 color
				if strings.Contains(row, tcTheme.ContributionLevel1) {
					t.Errorf("theme %q row %d should not contain Level1 color for zero-activity cells", cfgTheme.Name, rIdx)
				}
			}

			// 2. 256-Color Mode: zero-activity cells must contain 256-color EmptyCell escape
			c256Theme := ResolveTheme(cfgTheme, Color256)
			c256Rows := RenderContributionGraph(graph, c256Theme, cfg)
			for rIdx, row := range c256Rows {
				if !strings.Contains(row, c256Theme.EmptyCell) {
					t.Errorf("theme %q row %d missing 256-color EmptyCell escape (%q): %q",
						cfgTheme.Name, rIdx, c256Theme.EmptyCell, row)
				}
			}
		})
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
		if !strings.Contains(r, theme.ContributionLevel4) {
			t.Errorf("row %d = %q, missing 256-color level 4 %q", i, r, theme.ContributionLevel4)
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

	// First row (Sunday, d=0) is padding: matching GitHub's graph it
	// renders as visible no-contribution cells ("·" glyph per week).
	// 3 weeks * 2 chars = 6 visible characters.
	if rows[0] != "· · · " {
		t.Errorf("row[0] = %q, want no-contribution cells (\"· · · \")", rows[0])
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
