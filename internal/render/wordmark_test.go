package render

import (
	"strings"
	"testing"
)

func TestGetWordmark(t *testing.T) {
	asciiWM := GetWordmark(ColorASCII)
	if !strings.Contains(asciiWM, "G H E P P O") || !strings.Contains(asciiWM, "........") {
		t.Errorf("unexpected ASCII wordmark: %q", asciiWM)
	}

	unicodeWM := GetWordmark(ColorTrueColor)
	if !strings.Contains(unicodeWM, "G H E P P O") || !strings.Contains(unicodeWM, "········") {
		t.Errorf("unexpected Unicode wordmark: %q", unicodeWM)
	}
}

func TestRenderWordmark(t *testing.T) {
	themeASCII := GetTheme(ColorASCII, ThemeDark)
	strASCII, lenASCII := RenderWordmark(themeASCII)
	if lenASCII <= 0 || !strings.Contains(strASCII, "G H E P P O") {
		t.Errorf("RenderWordmark(ASCII) = (%q, %d)", strASCII, lenASCII)
	}

	themeTC := GetTheme(ColorTrueColor, ThemeDark)
	strTC, lenTC := RenderWordmark(themeTC)
	if lenTC <= 0 || !strings.Contains(strTC, themeTC.Primary) {
		t.Errorf("RenderWordmark(TrueColor) = (%q, %d)", strTC, lenTC)
	}
}
