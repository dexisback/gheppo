package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/dexisback/gheppo/internal/config"
	"github.com/dexisback/gheppo/internal/stats"
)

// TestDetectColorMode verifies environment variable mapping to color modes.
func TestDetectColorMode(t *testing.T) {
	tests := []struct {
		name      string
		noColor   string
		colorTerm string
		term      string
		want      ColorMode
	}{
		{
			name:      "NO_COLOR",
			noColor:   "1",
			colorTerm: "truecolor",
			term:      "xterm-256color",
			want:      ColorASCII,
		},
		{
			name:      "truecolor",
			colorTerm: "truecolor",
			term:      "xterm-256color",
			want:      ColorTrueColor,
		},
		{
			name:      "24bit",
			colorTerm: "24bit",
			term:      "xterm-256color",
			want:      ColorTrueColor,
		},
		{
			name: "256 color",
			term: "xterm-256color",
			want: Color256,
		},
		{
			name: "unknown/dumb terminal",
			term: "dumb",
			want: ColorASCII,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", tt.noColor)
			t.Setenv("COLORTERM", tt.colorTerm)
			t.Setenv("TERM", tt.term)

			got := DetectColorMode()
			if got != tt.want {
				t.Errorf("DetectColorMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDetectTheme verifies theme detection logic.
func TestDetectTheme(t *testing.T) {
	tests := []struct {
		name        string
		gheppoTheme string
		colorfgbg   string
		want        ThemeMode
	}{
		{
			name:        "explicit dark",
			gheppoTheme: "dark",
			want:        ThemeDark,
		},
		{
			name:        "explicit light",
			gheppoTheme: "light",
			want:        ThemeLight,
		},
		{
			name:      "colorfgbg dark",
			colorfgbg: "15;0",
			want:      ThemeDark,
		},
		{
			name:      "colorfgbg light",
			colorfgbg: "0;15",
			want:      ThemeLight,
		},
		{
			name: "default is dark",
			want: ThemeDark,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GHEPPO_THEME", tt.gheppoTheme)
			t.Setenv("COLORFGBG", tt.colorfgbg)

			got := DetectTheme()
			if got != tt.want {
				t.Errorf("DetectTheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGridNilSummary verifies nil safety.
func TestGridNilSummary(t *testing.T) {
	output := Grid(nil)
	if output != "" {
		t.Errorf("Grid(nil) = %q, want empty string", output)
	}
}

// TestGridEmptySummary verifies rendering of empty summary.
func TestGridEmptySummary(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("COLUMNS", "80")

	summary := &stats.Summary{
		Login: "testuser",
	}

	output := Grid(summary)
	if output == "" {
		t.Fatal("Grid(empty summary) returned empty output")
	}

	if !strings.Contains(output, "0 CONTRIBUTIONS") {
		t.Errorf("output missing 0 CONTRIBUTIONS: %q", output)
	}
	if !strings.Contains(output, "@testuser") {
		t.Errorf("output missing username: %q", output)
	}
}

// TestGridFullCardElements verifies all required sections are present.
func TestGridFullCardElements(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("COLUMNS", "80")

	// 5 weeks of data
	grid := make([][]stats.Cell, 5)
	startDate := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)
	for w := 0; w < 5; w++ {
		grid[w] = make([]stats.Cell, 7)
		for d := 0; d < 7; d++ {
			date := startDate.AddDate(0, 0, w*7+d)
			grid[w][d] = stats.Cell{
				Date:   date,
				Count:  d + 1,
				Bucket: (d % 4) + 1,
			}
		}
	}

	summary := &stats.Summary{
		Login:         "dexisback",
		Total:         1386,
		CurrentStreak: 7,
		LongestStreak: 39,
		FetchedAt:     time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
		Grid:          grid,
	}

	output := Grid(summary)

	// 1. Check vertical divider between columns
	if !strings.Contains(output, "|") {
		t.Errorf("output missing vertical divider: %q", output)
	}

	// 2. Check year and total contributions
	if !strings.Contains(output, "2026") {
		t.Errorf("output missing year: %q", output)
	}
	if !strings.Contains(output, "1,386 CONTRIBUTIONS") {
		t.Errorf("output missing total contributions with comma: %q", output)
	}

	// 3. Check stats (current streak is no longer shown by default per design spec)
	// Longest streak should be shown
	if !strings.Contains(output, "39 DAYS") {
		t.Errorf("output missing longest streak days: %q", output)
	}
	if !strings.Contains(output, "LONGEST STREAK") {
		t.Errorf("output missing longest streak label: %q", output)
	}

	// 4. Check username
	if !strings.Contains(output, "@dexisback") {
		t.Errorf("output missing @dexisback: %q", output)
	}

	// 5. Verify NO weekday labels (Sun, Mon, Tue, etc.)
	for _, day := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
		if strings.Contains(output, day) {
			t.Errorf("output contains weekday label %q (must NOT have weekday labels): %q", day, output)
		}
	}

	// 6. Check month label
	if !strings.Contains(output, "JAN") {
		t.Errorf("output missing JAN month label: %q", output)
	}
}

// TestResponsiveBreakpoints verifies rendering across terminal widths.
func TestResponsiveBreakpoints(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	// Generate 52 weeks
	grid := make([][]stats.Cell, 52)
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for w := 0; w < 52; w++ {
		grid[w] = make([]stats.Cell, 7)
		for d := 0; d < 7; d++ {
			grid[w][d] = stats.Cell{
				Date:   startDate.AddDate(0, 0, w*7+d),
				Count:  1,
				Bucket: 2,
			}
		}
	}

	summary := &stats.Summary{
		Login:         "dexisback",
		Total:         500,
		CurrentStreak: 12,
		LongestStreak: 45,
		FetchedAt:     time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
		Grid:          grid,
	}

	widths := []int{46, 50, 60, 70, 80, 100, 120, 160}
	for _, w := range widths {
		t.Run(string(rune(w)), func(t *testing.T) {
			layout := ComputeLayout(w, len(grid))
			card := RenderCard(summary, layout, ColorASCII, ThemeDark)

			if card == "" {
				t.Fatalf("rendered empty card for width %d", w)
			}

			lines := strings.Split(card, "\n")
			for i, line := range lines {
				// Strip ANSI if any and check line length does not exceed terminal width
				if len(line) > w+2 {
					t.Errorf("width %d: line %d length %d exceeds terminal width: %q", w, i, len(line), line)
				}
			}
		})
	}
}

// TestTrueColorAnd256Palettes verifies ANSI sequence generation for dark/light themes.
func TestTrueColorAnd256Palettes(t *testing.T) {
	for bucket := 0; bucket <= 4; bucket++ {
		tcDark := CellColor(ColorTrueColor, ThemeDark, bucket)
		if !strings.Contains(tcDark, "\033[38;2;") {
			t.Errorf("tcDark bucket %d missing truecolor: %q", bucket, tcDark)
		}

		tcLight := CellColor(ColorTrueColor, ThemeLight, bucket)
		if !strings.Contains(tcLight, "\033[38;2;") {
			t.Errorf("tcLight bucket %d missing truecolor: %q", bucket, tcLight)
		}

		c256Dark := CellColor(Color256, ThemeDark, bucket)
		if !strings.Contains(c256Dark, "\033[38;5;") {
			t.Errorf("c256Dark bucket %d missing 256color: %q", bucket, c256Dark)
		}

		c256Light := CellColor(Color256, ThemeLight, bucket)
		if !strings.Contains(c256Light, "\033[38;5;") {
			t.Errorf("c256Light bucket %d missing 256color: %q", bucket, c256Light)
		}
	}
}

// TestAnimateNonInteractiveFallback verifies Animate writes without error in non-interactive environment.
func TestAnimateNonInteractiveFallback(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("GHEPPO_NO_ANIMATION", "1")

	var buf bytes.Buffer
	summary := &stats.Summary{
		Login: "dexisback",
		Total: 100,
	}

	Animate(&buf, summary)
	if !strings.Contains(buf.String(), "@dexisback") {
		t.Errorf("Animate output missing username: %s", buf.String())
	}
}

// TestRenderDirect verifies the Render entry point with explicit width.
func TestRenderDirect(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	summary := &stats.Summary{
		Login: "directuser",
		Total: 42,
	}

	output := Render(summary, 90)
	if !strings.Contains(output, "@directuser") {
		t.Errorf("Render() missing username: %q", output)
	}
	if !strings.Contains(output, "42 CONTRIBUTIONS") {
		t.Errorf("Render() missing contributions: %q", output)
	}
}

// TestRenderWithTheme verifies rendering with explicit Theme.
func TestRenderWithTheme(t *testing.T) {
	summary := &stats.Summary{
		Login: "themeuser",
		Total: 10,
	}
	layout := CalculateLayout(80, 0)
	theme := GetTheme(ColorASCII, ThemeDark)

	output := RenderWithTheme(summary, layout, theme)
	if !strings.Contains(output, "@themeuser") {
		t.Errorf("RenderWithTheme() missing username: %q", output)
	}
}

// TestRenderMultiThemes verifies rendering output with all registered themes.
func TestRenderMultiThemes(t *testing.T) {
	summary := &stats.Summary{
		Login:         "dexisback",
		Total:         542,
		LongestStreak: 18,
		DailyAverage:  3.2,
		Followers:     50,
		Following:     40,
		Repos:         15,
		TotalStars:    99,
		BestDayCount:  20,
		BestDay:       time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
	}
	layout := CalculateLayout(80, 0)

	for _, themeName := range config.AvailableThemeNames() {
		cfgTheme, ok := config.GetTheme(themeName)
		if !ok {
			t.Fatalf("theme %q not found in registry", themeName)
		}

		// TrueColor render pass
		tcTheme := ResolveTheme(cfgTheme, ColorTrueColor)
		tcOutput := RenderWithTheme(summary, layout, tcTheme)

		if !strings.Contains(tcOutput, "@dexisback") {
			t.Errorf("theme %q output missing username: %q", themeName, tcOutput)
		}
		if !strings.Contains(tcOutput, "542") || !strings.Contains(tcOutput, "CONTRIBUTIONS") {
			t.Errorf("theme %q output missing contributions: %q", themeName, tcOutput)
		}
		if !strings.Contains(tcOutput, tcTheme.ActivityHigh) {
			t.Errorf("theme %q output missing theme accent color %q", themeName, tcTheme.ActivityHigh)
		}

		// 256-color render pass
		c256Theme := ResolveTheme(cfgTheme, Color256)
		c256Output := RenderWithTheme(summary, layout, c256Theme)
		if !strings.Contains(c256Output, "@dexisback") {
			t.Errorf("theme %q 256-color output missing username", themeName)
		}
	}
}

func TestRenderLeetCodeSummary(t *testing.T) {
	summary := &stats.Summary{
		Source:         "leetcode",
		Login:          "tourist_lc",
		Total:          487,
		ProblemsSolved: 487,
		EasySolved:     210,
		MediumSolved:   231,
		HardSolved:     46,
		ContestRating:  1842,
		GlobalRanking:  38421,
		LongestStreak:  42,
		DailyAverage:   2.7,
		BestDayCount:   8,
		BestDay:        time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
	}
	layout := CalculateLayout(80, 0)
	theme := GetTheme(ColorASCII, ThemeDark)

	output := RenderWithTheme(summary, layout, theme)

	if !strings.Contains(output, "@tourist_lc") {
		t.Errorf("missing username: %q", output)
	}
	if !strings.Contains(output, "487 PROBLEMS SOLVED") {
		t.Errorf("missing PROBLEMS SOLVED: %q", output)
	}
	if !strings.Contains(output, "CONTEST RATING") || !strings.Contains(output, "GLOBAL RANK") {
		t.Errorf("missing contest rating/rank headers: %q", output)
	}
	if !strings.Contains(output, "1,842") || !strings.Contains(output, "#38,421") {
		t.Errorf("missing contest rating/rank values: %q", output)
	}
	if !strings.Contains(output, "EASY") || !strings.Contains(output, "MEDIUM") || !strings.Contains(output, "HARD") {
		t.Errorf("missing difficulty breakdown: %q", output)
	}
	if !strings.Contains(output, "210") || !strings.Contains(output, "231") || !strings.Contains(output, "46") {
		t.Errorf("missing difficulty solved numbers: %q", output)
	}
}

func TestRenderLeetCodeStatsPanel(t *testing.T) {
	summary := &stats.Summary{
		Source:         "leetcode",
		Login:          "unranked_user",
		Total:          100,
		ProblemsSolved: 100,
		EasySolved:     50,
		MediumSolved:   40,
		HardSolved:     10,
		GlobalRanking:  105221,
		Reputation:     230780,
	}
	theme := GetTheme(ColorASCII, ThemeDark)
	rows := RenderStatsPanel(summary, theme)

	var joined strings.Builder
	for _, r := range rows {
		joined.WriteString(r.Content + "\n")
	}
	content := joined.String()

	if !strings.Contains(content, "GLOBAL RANK") || !strings.Contains(content, "REPUTATION") {
		t.Errorf("missing unranked headers: %q", content)
	}
	if !strings.Contains(content, "#105,221") || !strings.Contains(content, "230,780") {
		t.Errorf("missing unranked values: %q", content)
	}
	if !strings.Contains(content, "EASY") || !strings.Contains(content, "50") {
		t.Errorf("missing difficulty breakdown: %q", content)
	}
}

