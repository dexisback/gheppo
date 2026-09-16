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

		// Test 256-color resolution
		c256Theme := ResolveTheme(cfgTheme, Color256)
		if !strings.HasPrefix(c256Theme.ContributionLevel4, "\033[38;5;") {
			t.Errorf("theme %q Level4 missing 256color prefix: %q", themeName, c256Theme.ContributionLevel4)
		}

		// Test ASCII resolution
		asciiTheme := ResolveTheme(cfgTheme, ColorASCII)
		if asciiTheme.ContributionLevel4 != "" {
			t.Errorf("theme %q Level4 in ASCII mode should be empty, got %q", themeName, asciiTheme.ContributionLevel4)
		}

		// Verify CellColor output
		for b := 0; b <= 4; b++ {
			c := tcTheme.CellColor(b)
			if !strings.Contains(c, "■") {
				t.Errorf("theme %q bucket %d missing block glyph: %q", themeName, b, c)
			}
		}
	}
}

func TestThemeDistinctColorSequences(t *testing.T) {
	ghTheme, _ := config.GetTheme("github")
	nordTheme, _ := config.GetTheme("nord")
	catppuccinTheme, _ := config.GetTheme("catppuccin")

	ghRender := ResolveTheme(ghTheme, ColorTrueColor)
	nordRender := ResolveTheme(nordTheme, ColorTrueColor)
	catRender := ResolveTheme(catppuccinTheme, ColorTrueColor)

	if ghRender.ContributionLevel4 == nordRender.ContributionLevel4 {
		t.Errorf("expected GitHub and Nord Level4 colors to differ, both are %q", ghRender.ContributionLevel4)
	}
	if ghRender.ContributionLevel4 == catRender.ContributionLevel4 {
		t.Errorf("expected GitHub and Catppuccin Level4 colors to differ, both are %q", ghRender.ContributionLevel4)
	}
	if nordRender.ContributionLevel4 == catRender.ContributionLevel4 {
		t.Errorf("expected Nord and Catppuccin Level4 colors to differ, both are %q", nordRender.ContributionLevel4)
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
