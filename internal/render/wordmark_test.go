package render

import (
	"strings"
	"testing"
)

func TestWordmarkCanvasGeometry(t *testing.T) {
	canvas := BuildWordmarkCanvas()
	if canvas.Width != 46 {
		t.Errorf("expected canvas width 46, got %d", canvas.Width)
	}
	if canvas.Height != 3 {
		t.Errorf("expected canvas height 3, got %d", canvas.Height)
	}
	if len(canvas.Points) == 0 {
		t.Fatal("expected canvas to have points, got 0")
	}

	for i, pt := range canvas.Points {
		if pt.Row < 0 || pt.Row >= canvas.Height {
			t.Errorf("point %d: row %d out of bounds [0, %d)", i, pt.Row, canvas.Height)
		}
		if pt.Col < 0 || pt.Col >= canvas.Width {
			t.Errorf("point %d: col %d out of bounds [0, %d)", i, pt.Col, canvas.Width)
		}
		if pt.Threshold < 0.0 || pt.Threshold > 1.0 {
			t.Errorf("point %d: threshold %.2f out of bounds [0.0, 1.0]", i, pt.Threshold)
		}
		if pt.Glyph == "" || pt.AsciiAlt == "" {
			t.Errorf("point %d: missing glyph or ascii alt", i)
		}
	}
}

func TestWordmarkProgressStates(t *testing.T) {
	theme := GetTheme(ColorASCII, ThemeDark)

	steps := []float64{0.0, 0.25, 0.50, 0.75, 1.0}
	prevNonSpace := 0

	for _, p := range steps {
		lines, visLen := RenderWordmarkWithState(theme, WordmarkState{Progress: p})
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d for progress %.2f", len(lines), p)
		}
		if visLen != 46 {
			t.Errorf("expected visLen 46, got %d for progress %.2f", visLen, p)
		}

		// Count non-space characters
		nonSpace := 0
		for _, l := range lines {
			for _, ch := range l {
				if ch != ' ' {
					nonSpace++
				}
			}
		}

		if nonSpace < prevNonSpace {
			t.Errorf("expected non-space count to increase monotonically: prev %d, got %d at progress %.2f", prevNonSpace, nonSpace, p)
		}
		prevNonSpace = nonSpace
	}
}

func TestWordmarkStaticMatch(t *testing.T) {
	for _, mode := range []ColorMode{ColorTrueColor, Color256, ColorASCII} {
		theme := GetTheme(mode, ThemeDark)
		staticLines, staticLen := RenderWordmark(theme)
		dynamicLines, dynamicLen := RenderWordmarkWithState(theme, WordmarkState{Progress: 1.0})

		if staticLen != dynamicLen {
			t.Errorf("mode %v: static len %d != dynamic len %d", mode, staticLen, dynamicLen)
		}
		if len(staticLines) != len(dynamicLines) {
			t.Fatalf("mode %v: line count mismatch", mode)
		}
		for i := range staticLines {
			if staticLines[i] != dynamicLines[i] {
				t.Errorf("mode %v: line %d mismatch:\nstatic:  %q\ndynamic: %q", mode, i, staticLines[i], dynamicLines[i])
			}
		}
	}
}

func TestWordmarkColorModes(t *testing.T) {
	// ASCII mode should contain no escape sequences
	themeASCII := GetTheme(ColorASCII, ThemeDark)
	linesASCII, _ := RenderWordmark(themeASCII)
	for i, line := range linesASCII {
		if strings.Contains(line, "\033") {
			t.Errorf("ASCII line %d contains ANSI escape: %q", i, line)
		}
	}

	// TrueColor mode should contain 24-bit ANSI escapes
	themeTC := GetTheme(ColorTrueColor, ThemeDark)
	linesTC, _ := RenderWordmark(themeTC)
	hasEscape := false
	for _, line := range linesTC {
		if strings.Contains(line, "\033[38;2;") {
			hasEscape = true
			break
		}
	}
	if !hasEscape {
		t.Errorf("TrueColor wordmark missing 24-bit ANSI codes")
	}

	// 256 color mode should contain 256-color ANSI escapes
	theme256 := GetTheme(Color256, ThemeDark)
	lines256, _ := RenderWordmark(theme256)
	has256Escape := false
	for _, line := range lines256 {
		if strings.Contains(line, "\033[38;5;") {
			has256Escape = true
			break
		}
	}
	if !has256Escape {
		t.Errorf("256-color wordmark missing 256-color ANSI codes")
	}
}

func TestWordmarkResponsiveWidth(t *testing.T) {
	theme := GetTheme(ColorASCII, ThemeDark)

	// Width constrained to 40 (trace trimmed by 6)
	lines40, visLen40 := RenderWordmarkWithStateAndWidth(theme, WordmarkState{Progress: 1.0}, 40)
	if visLen40 != 40 {
		t.Errorf("expected visLen 40, got %d", visLen40)
	}
	for _, l := range lines40 {
		if len(l) != 40 {
			t.Errorf("expected line length 40, got %d", len(l))
		}
	}

	// Width constrained to 32 (trace fully omitted, lettering intact)
	lines32, visLen32 := RenderWordmarkWithStateAndWidth(theme, WordmarkState{Progress: 1.0}, 32)
	if visLen32 != 32 {
		t.Errorf("expected visLen 32, got %d", visLen32)
	}
	for _, l := range lines32 {
		if len(l) != 32 {
			t.Errorf("expected line length 32, got %d", len(l))
		}
	}
}

func TestGetWordmarkCompatibility(t *testing.T) {
	asciiWM := GetWordmark(ColorASCII)
	if !strings.Contains(asciiWM, "#") && !strings.Contains(asciiWM, "-") {
		t.Errorf("unexpected ASCII wordmark format: %q", asciiWM)
	}

	tcWM := GetWordmark(ColorTrueColor)
	if !strings.Contains(tcWM, "█") && !strings.Contains(tcWM, "▀") {
		t.Errorf("unexpected TrueColor wordmark format: %q", tcWM)
	}
}
