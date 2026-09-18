package render

import (
	"strings"
	"testing"

	"github.com/dexisback/gheppo/internal/config"
)

func TestThemeResolution(t *testing.T) {
	for _, themeName := range config.AvailableThemeNames() {
		cfgTheme, ok := config.GetTheme(themeName)
		if !ok {
			t.Fatalf("expected theme %q to exist", themeName)
		}

		// Test TrueColor resolution
		tcTheme := ResolveTheme(cfgTheme, ColorTrueColor)
		if tcTheme.Config.Name != themeName {
			t.Errorf("tcTheme.Config.Name = %q, want %q", tcTheme.Config.Name, themeName)
		}
		if !strings.HasPrefix(tcTheme.ContributionLevel4, "\033[38;2;") {
			t.Errorf("theme %q Level4 missing truecolor prefix: %q", themeName, tcTheme.ContributionLevel4)
		}
		if !strings.HasPrefix(tcTheme.EmptyCell, "\033[38;2;") {
			t.Errorf("theme %q EmptyCell missing truecolor prefix: %q", themeName, tcTheme.EmptyCell)
		}
		if tcTheme.EmptyCell == tcTheme.ContributionLevel1 {
			t.Errorf("theme %q EmptyCell (%q) must differ from Level1 (%q)", themeName, tcTheme.EmptyCell, tcTheme.ContributionLevel1)
		}

		// Verify Styles() has EmptyCell style defined
		styles := tcTheme.Styles()
		if styles.EmptyCell.GetForeground() == nil {
			t.Errorf("theme %q Styles().EmptyCell missing foreground", themeName)
		}

		// Test 256-color resolution
		c256Theme := ResolveTheme(cfgTheme, Color256)
		if !strings.HasPrefix(c256Theme.ContributionLevel4, "\033[38;5;") {
			t.Errorf("theme %q Level4 missing 256color prefix: %q", themeName, c256Theme.ContributionLevel4)
		}
		if !strings.HasPrefix(c256Theme.EmptyCell, "\033[38;5;") {
			t.Errorf("theme %q EmptyCell missing 256color prefix: %q", themeName, c256Theme.EmptyCell)
		}

		// Test ASCII resolution
		asciiTheme := ResolveTheme(cfgTheme, ColorASCII)
		if asciiTheme.ContributionLevel4 != "" {
			t.Errorf("theme %q Level4 in ASCII mode should be empty, got %q", themeName, asciiTheme.ContributionLevel4)
		}
		if asciiTheme.EmptyCell != "" {
			t.Errorf("theme %q EmptyCell in ASCII mode should be empty, got %q", themeName, asciiTheme.EmptyCell)
		}

		// Verify CellColor output
		for b := 0; b <= 4; b++ {
			c := tcTheme.CellColor(b)
			if !strings.Contains(c, "■") {
				t.Errorf("theme %q bucket %d missing block glyph: %q", themeName, b, c)
			}
		}

		// Verify CellColor(0) specifically uses tcTheme.EmptyCell
		c0 := tcTheme.CellColor(0)
		expectedC0 := tcTheme.EmptyCell + "■" + tcTheme.Reset
		if c0 != expectedC0 {
			t.Errorf("theme %q CellColor(0) = %q, want %q", themeName, c0, expectedC0)
		}
	}
}

func TestThemeDistinctColorSequences(t *testing.T) {
	themes := config.ListThemes()

	// Verify all themes have distinct Level 4 and distinct Level 0 (EmptyCell) colors
	for i := 0; i < len(themes); i++ {
		for j := i + 1; j < len(themes); j++ {
			t1 := ResolveTheme(themes[i], ColorTrueColor)
			t2 := ResolveTheme(themes[j], ColorTrueColor)

			if t1.ContributionLevel4 == t2.ContributionLevel4 {
				t.Errorf("expected themes %q and %q Level4 colors to differ, both are %q",
					themes[i].Name, themes[j].Name, t1.ContributionLevel4)
			}
			if t1.EmptyCell == t2.EmptyCell {
				t.Errorf("expected themes %q and %q EmptyCell colors to differ, both are %q",
					themes[i].Name, themes[j].Name, t1.EmptyCell)
			}
		}
	}
}

func TestThemeFormatHelpers(t *testing.T) {
	if got := formatNumber(1234567); got != "1,234,567" {
		t.Errorf("formatNumber(1234567) = %q, want '1,234,567'", got)
	}
	if got := formatNumber(0); got != "0" {
		t.Errorf("formatNumber(0) = %q, want '0'", got)
	}
	if got := formatFloat(3.14159); got != "3.1" {
		t.Errorf("formatFloat(3.14159) = %q, want '3.1'", got)
	}
}
