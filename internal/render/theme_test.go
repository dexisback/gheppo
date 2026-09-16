package render

import (
	"strings"
	"testing"
)

func TestThemeCellColors(t *testing.T) {
	themeTC := GetTheme(ColorTrueColor, ThemeDark)
	if themeTC.EmptyCell != darkEmptyCell {
		t.Errorf("EmptyCell = %q, want %q", themeTC.EmptyCell, darkEmptyCell)
	}
	if themeTC.ContributionLevel1 != darkActivityLevel1 {
		t.Errorf("Level1 = %q, want %q", themeTC.ContributionLevel1, darkActivityLevel1)
	}
	if themeTC.ContributionLevel4 != darkActivityLevel4 {
		t.Errorf("Level4 = %q, want %q", themeTC.ContributionLevel4, darkActivityLevel4)
	}

	cell4 := themeTC.CellColor(4)
	if !strings.Contains(cell4, darkActivityLevel4) || !strings.Contains(cell4, "■") {
		t.Errorf("CellColor(4) = %q, want level 4 block", cell4)
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
