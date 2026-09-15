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

// Design-specified color palettes
// Dark Theme: TrueColor exact values from design spec
var (
	// Dark Theme - TrueColor
	darkBorder        = "\033[38;2;48;54;61m"    // #30363D
	darkPrimaryText   = "\033[38;2;240;246;252m" // #F0F6FC
	darkSecondaryText = "\033[38;2;139;148;158m" // #8B949E
	darkMutedText     = "\033[38;2;110;118;129m" // #6E7681

	darkEmptyCell      = "\033[38;2;33;38;45m"    // #21262D
	darkActivityLevel1 = "\033[38;2;14;107;59m"   // #0E6B3B
	darkActivityLevel2 = "\033[38;2;64;196;99m"   // #40C463
	darkActivityLevel3 = "\033[38;2;53;208;111m"  // #35D06F
	darkActivityLevel4 = "\033[38;2;126;231;135m" // #7EE787

	// Light Theme - TrueColor
	lightBorder        = "\033[38;2;208;215;222m" // #D0D7DE
	lightPrimaryText   = "\033[38;2;31;35;40m"    // #1F2328
	lightSecondaryText = "\033[38;2;101;109;118m" // #656D76
	lightMutedText     = "\033[38;2;87;96;106m"   // #57606A

	lightEmptyCell      = "\033[38;2;235;237;240m" // #EBEDF0
	lightActivityLevel1 = "\033[38;2;155;233;168m" // #9BE9A8
	lightActivityLevel2 = "\033[38;2;64;196;99m"   // #40C463
	lightActivityLevel3 = "\033[38;2;48;161;78m"   // #30A14E
	lightActivityLevel4 = "\033[38;2;33;110;57m"   // #216E39

	// 256-color approximations
	// Dark Theme
	dark256Border         = "\033[38;5;237m"
	dark256PrimaryText    = "\033[38;5;231m"
	dark256SecondaryText  = "\033[38;5;246m"
	dark256MutedText      = "\033[38;5;243m"
	dark256EmptyCell      = "\033[38;5;235m"
	dark256ActivityLevel1 = "\033[38;5;28m"
	dark256ActivityLevel2 = "\033[38;5;41m"
	dark256ActivityLevel3 = "\033[38;5;48m"
	dark256ActivityLevel4 = "\033[38;5;83m"

	// Light Theme
	light256Border         = "\033[38;5;252m"
	light256PrimaryText    = "\033[38;5;235m"
	light256SecondaryText  = "\033[38;5;241m"
	light256MutedText      = "\033[38;5;243m"
	light256EmptyCell      = "\033[38;5;254m"
	light256ActivityLevel1 = "\033[38;5;157m"
	light256ActivityLevel2 = "\033[38;5;77m"
	light256ActivityLevel3 = "\033[38;5;71m"
	light256ActivityLevel4 = "\033[38;5;29m"

	reset = "\033[0m"
	bold  = "\033[1m"
)

// ASCII intensity glyphs (NO_COLOR)
var asciiGlyphs = map[int]string{
	0: "·",
	1: "░",
	2: "▒",
	3: "▓",
	4: "█",
}

// Pure ASCII fallback (when unicode might not work)
var pureAsciiGlyphs = map[int]string{
	0: ".",
	1: "-",
	2: "=",
	3: "+",
	4: "#",
}

// CellColor returns the formatted cell glyph for the given mode, theme, and bucket.
func CellColor(mode ColorMode, theme ThemeMode, bucket int) string {
	if bucket < 0 {
		bucket = 0
	}
	if bucket > 4 {
		bucket = 4
	}

	if mode == ColorASCII {
		glyph := asciiGlyphs[bucket]
		if os.Getenv("NO_COLOR") != "" {
			glyph = pureAsciiGlyphs[bucket]
		}
		return glyph
	}

	// Block character for colored cells
	block := "■"

	var color string
	if theme == ThemeDark {
		if mode == Color256 {
			switch bucket {
			case 0:
				color = dark256EmptyCell
			case 1:
				color = dark256ActivityLevel1
			case 2:
				color = dark256ActivityLevel2
			case 3:
				color = dark256ActivityLevel3
			case 4:
				color = dark256ActivityLevel4
			}
		} else { // TrueColor
			switch bucket {
			case 0:
				color = darkEmptyCell
			case 1:
				color = darkActivityLevel1
			case 2:
				color = darkActivityLevel2
			case 3:
				color = darkActivityLevel3
			case 4:
				color = darkActivityLevel4
			}
		}
	} else { // Light theme
		if mode == Color256 {
			switch bucket {
			case 0:
				color = light256EmptyCell
			case 1:
				color = light256ActivityLevel1
			case 2:
				color = light256ActivityLevel2
			case 3:
				color = light256ActivityLevel3
			case 4:
				color = light256ActivityLevel4
			}
		} else { // TrueColor
			switch bucket {
			case 0:
				color = lightEmptyCell
			case 1:
				color = lightActivityLevel1
			case 2:
				color = lightActivityLevel2
			case 3:
				color = lightActivityLevel3
			case 4:
				color = lightActivityLevel4
			}
		}
	}

	return color + block + reset
}

// UI Color tokens for text, borders, headers, and metadata
type PaletteTokens struct {
	Border       string
	Primary      string
	Secondary    string
	Muted        string
	ActivityHigh string // For year progress and green accent numbers
	Reset        string
	Bold         string
}

// GetTokens returns ANSI styling tokens for the active mode and theme.
func GetTokens(mode ColorMode, theme ThemeMode) PaletteTokens {
	if mode == ColorASCII {
		return PaletteTokens{
			Reset: "",
			Bold:  "",
		}
	}

	if theme == ThemeDark {
		if mode == Color256 {
			return PaletteTokens{
				Border:       dark256Border,
				Primary:      dark256PrimaryText,
				Secondary:    dark256SecondaryText,
				Muted:        dark256MutedText,
				ActivityHigh: dark256ActivityLevel3,
				Reset:        reset,
				Bold:         bold,
			}
		}
		// TrueColor Dark
		return PaletteTokens{
			Border:       darkBorder,
			Primary:      darkPrimaryText,
			Secondary:    darkSecondaryText,
			Muted:        darkMutedText,
			ActivityHigh: darkActivityLevel3,
			Reset:        reset,
			Bold:         bold,
		}
	}

	// Light Theme
	if mode == Color256 {
		return PaletteTokens{
			Border:       light256Border,
			Primary:      light256PrimaryText,
			Secondary:    light256SecondaryText,
			Muted:        light256MutedText,
			ActivityHigh: light256ActivityLevel3,
			Reset:        reset,
			Bold:         bold,
		}
	}
	// TrueColor Light
	return PaletteTokens{
		Border:       lightBorder,
		Primary:      lightPrimaryText,
		Secondary:    lightSecondaryText,
		Muted:        lightMutedText,
		ActivityHigh: lightActivityLevel3,
		Reset:        reset,
		Bold:         bold,
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

func formatFloat(f float64) string {
	return fmt.Sprintf("%.1f", f)
}
