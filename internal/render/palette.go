package render

import (
	"fmt"
	"os"
	"strings"
)

type ColorMode int

const (
	ColorTrueColor ColorMode = iota
	Color256
	ColorASCII
)

type ThemeMode int

const (
	ThemeDark ThemeMode = iota
	ThemeLight
)

// DetectColorMode determines what level of terminal color support is available.
func DetectColorMode() ColorMode {
	if os.Getenv("NO_COLOR") != "" {
		return ColorASCII
	}

	colorterm := strings.ToLower(os.Getenv("COLORTERM"))
	if colorterm == "truecolor" || colorterm == "24bit" {
		return ColorTrueColor
	}

	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "256color") {
		return Color256
	}

	if term == "dumb" || term == "" {
		return ColorASCII
	}

	return ColorASCII
}

// DetectTheme determines whether to render for a dark or light terminal background.
func DetectTheme() ThemeMode {
	explicit := strings.ToLower(strings.TrimSpace(os.Getenv("GHEPPO_THEME")))
	if explicit == "light" {
		return ThemeLight
	}
	if explicit == "dark" {
		return ThemeDark
	}

	// Heuristic: check COLORFGBG
	colorfgbg := os.Getenv("COLORFGBG")
	if colorfgbg != "" {
		parts := strings.Split(colorfgbg, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			if bg == "15" || bg == "7" || bg == "white" {
				return ThemeLight
			}
		}
	}

	return ThemeDark
}

// Calibrated Palettes
var (
	// Dark Theme TrueColor (tuned for optimal contrast and vibrant progression)
	trueColorDarkPalette = map[int]string{
		0: "\033[38;2;33;38;45m■\033[0m",   // Level 0: subtle dark slate block (#21262d)
		1: "\033[38;2;14;68;41m■\033[0m",   // Level 1: deep emerald green (#0e4429)
		2: "\033[38;2;0;109;50m■\033[0m",   // Level 2: rich forest green (#006d32)
		3: "\033[38;2;38;166;65m■\033[0m",  // Level 3: bright leaf green (#26a641)
		4: "\033[38;2;87;255;122m■\033[0m", // Level 4: luminous mint / lime (#57ff7a)
	}

	// Dark Theme 256-color
	color256DarkPalette = map[int]string{
		0: "\033[38;5;238m■\033[0m", // Level 0: dark grey
		1: "\033[38;5;22m■\033[0m",  // Level 1: dark green
		2: "\033[38;5;28m■\033[0m",  // Level 2: medium green
		3: "\033[38;5;35m■\033[0m",  // Level 3: bright green
		4: "\033[38;5;46m■\033[0m",  // Level 4: neon green
	}

	// Light Theme TrueColor
	trueColorLightPalette = map[int]string{
		0: "\033[38;2;235;237;240m■\033[0m", // Level 0: light grey block (#ebedf0)
		1: "\033[38;2;155;233;168m■\033[0m", // Level 1: light mint (#9be9a8)
		2: "\033[38;2;64;196;99m■\033[0m",   // Level 2: spring green (#40c463)
		3: "\033[38;2;46;160;67m■\033[0m",   // Level 3: rich forest green (#2ea043)
		4: "\033[38;2;26;127;55m■\033[0m",   // Level 4: deep emerald (#1a7f37)
	}

	// Light Theme 256-color
	color256LightPalette = map[int]string{
		0: "\033[38;5;254m■\033[0m", // Level 0: light grey
		1: "\033[38;5;151m■\033[0m", // Level 1: light mint
		2: "\033[38;5;77m■\033[0m",  // Level 2: spring green
		3: "\033[38;5;71m■\033[0m",  // Level 3: deep green
		4: "\033[38;5;29m■\033[0m",  // Level 4: dark emerald
	}

	// ASCII Intensity Glyphs
	asciiPalette = map[int]string{
		0: "·", // Level 0: small centered dot
		1: "░", // Level 1: light shade
		2: "▒", // Level 2: medium shade
		3: "▓", // Level 3: dark shade
		4: "█", // Level 4: full solid block
	}

	// Pure ASCII fallback (when NO_COLOR is set)
	pureAsciiPalette = map[int]string{
		0: ".",
		1: "-",
		2: "=",
		3: "+",
		4: "#",
	}
)

// CellColor returns the formatted cell glyph for the given mode, theme, and bucket.
func CellColor(mode ColorMode, theme ThemeMode, bucket int) string {
	if bucket < 0 {
		bucket = 0
	}
	if bucket > 4 {
		bucket = 4
	}

	if mode == ColorASCII {
		if os.Getenv("NO_COLOR") != "" {
			return pureAsciiPalette[bucket]
		}
		return asciiPalette[bucket]
	}

	if theme == ThemeLight {
		if mode == Color256 {
			return color256LightPalette[bucket]
		}
		return trueColorLightPalette[bucket]
	}

	if mode == Color256 {
		return color256DarkPalette[bucket]
	}
	return trueColorDarkPalette[bucket]
}

// UI Color tokens for text, borders, headers, and metadata
type PaletteTokens struct {
	Border    string
	Wordmark  string
	Primary   string
	Secondary string
	Muted     string
	Accent    string
	Reset     string
	Bold      string
}

// GetTokens returns ANSI styling tokens for the active mode and theme.
func GetTokens(mode ColorMode, theme ThemeMode) PaletteTokens {
	if mode == ColorASCII {
		return PaletteTokens{}
	}

	reset := "\033[0m"
	bold := "\033[1m"

	if theme == ThemeLight {
		if mode == Color256 {
			return PaletteTokens{
				Border:    "\033[38;5;250m", // #bcbcbc
				Wordmark:  "\033[38;5;235m", // #262626
				Primary:   "\033[38;5;234m", // #1c1c1c
				Secondary: "\033[38;5;242m", // #6c6c6c
				Muted:     "\033[38;5;246m", // #949494
				Accent:    "\033[38;5;29m",  // #00875f
				Reset:     reset,
				Bold:      bold,
			}
		}
		// TrueColor Light
		return PaletteTokens{
			Border:    "\033[38;2;208;215;222m", // #d0d7de
			Wordmark:  "\033[38;2;36;41;47m",    // #24292f
			Primary:   "\033[38;2;36;41;47m",    // #24292f
			Secondary: "\033[38;2;87;96;106m",   // #57606a
			Muted:     "\033[38;2;140;149;159m", // #8c959f
			Accent:    "\033[38;2;26;127;55m",   // #1a7f37
			Reset:     reset,
			Bold:      bold,
		}
	}

	// Dark Theme
	if mode == Color256 {
		return PaletteTokens{
			Border:    "\033[38;5;238m", // #444444
			Wordmark:  "\033[38;5;255m", // #eeeeee
			Primary:   "\033[38;5;253m", // #dadada
			Secondary: "\033[38;5;245m", // #8a8a8a
			Muted:     "\033[38;5;240m", // #585858
			Accent:    "\033[38;5;46m",  // #00ff00
			Reset:     reset,
			Bold:      bold,
		}
	}

	// TrueColor Dark
	return PaletteTokens{
		Border:    "\033[38;2;48;54;61m",    // #30363d (GitHub dark border)
		Wordmark:  "\033[38;2;240;246;252m", // #f0f6fc (Crisp high-contrast white)
		Primary:   "\033[38;2;230;237;243m", // #e6edf3 (Primary text)
		Secondary: "\033[38;2;139;148;158m", // #8b949e (Secondary text)
		Muted:     "\033[38;2;110;118;129m", // #6e7681 (Muted labels)
		Accent:    "\033[38;2;57;211;83m",   // #39d353 (Vibrant green accent)
		Reset:     reset,
		Bold:      bold,
	}
}

func formatNumber(n int) string {
	in := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(in)+(len(in)-1)/3)
	for i, c := range in {
		if i > 0 && (len(in)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}
