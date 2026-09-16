package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorMode specifies the terminal color capability level.
type ColorMode int

const (
	ColorTrueColor ColorMode = iota
	Color256
	ColorASCII
)

// ThemeMode specifies the terminal theme (dark or light background).
type ThemeMode int

const (
	ThemeDark ThemeMode = iota
	ThemeLight
)

// Styles encapsulates Lip Gloss style definitions for the presentation layer.
type Styles struct {
	Primary      lipgloss.Style
	Secondary    lipgloss.Style
	Muted        lipgloss.Style
	ActivityHigh lipgloss.Style
	Border       lipgloss.Style
	EmptyCell    lipgloss.Style
	Level1       lipgloss.Style
	Level2       lipgloss.Style
	Level3       lipgloss.Style
	Level4       lipgloss.Style
}

// DefaultStyles returns standard Lip Gloss styles for the given theme.
func DefaultStyles(theme Theme) Styles {
	if theme.Mode == ColorASCII {
		return Styles{
			Primary:      lipgloss.NewStyle(),
			Secondary:    lipgloss.NewStyle(),
			Muted:        lipgloss.NewStyle(),
			ActivityHigh: lipgloss.NewStyle(),
			Border:       lipgloss.NewStyle(),
			EmptyCell:    lipgloss.NewStyle(),
			Level1:       lipgloss.NewStyle(),
			Level2:       lipgloss.NewStyle(),
			Level3:       lipgloss.NewStyle(),
			Level4:       lipgloss.NewStyle(),
		}
	}

	if theme.Theme == ThemeLight {
		return Styles{
			Primary:      lipgloss.NewStyle().Foreground(lipgloss.Color("#1F2328")).Bold(true),
			Secondary:    lipgloss.NewStyle().Foreground(lipgloss.Color("#656D76")),
			Muted:        lipgloss.NewStyle().Foreground(lipgloss.Color("#57606A")),
			ActivityHigh: lipgloss.NewStyle().Foreground(lipgloss.Color("#30A14E")).Bold(true),
			Border:       lipgloss.NewStyle().Foreground(lipgloss.Color("#D0D7DE")),
			EmptyCell:    lipgloss.NewStyle().Foreground(lipgloss.Color("#EBEDF0")),
			Level1:       lipgloss.NewStyle().Foreground(lipgloss.Color("#9BE9A8")),
			Level2:       lipgloss.NewStyle().Foreground(lipgloss.Color("#40C463")),
			Level3:       lipgloss.NewStyle().Foreground(lipgloss.Color("#30A14E")),
			Level4:       lipgloss.NewStyle().Foreground(lipgloss.Color("#216E39")),
		}
	}

	// Dark Theme (default)
	return Styles{
		Primary:      lipgloss.NewStyle().Foreground(lipgloss.Color("#F0F6FC")).Bold(true),
		Secondary:    lipgloss.NewStyle().Foreground(lipgloss.Color("#8B949E")),
		Muted:        lipgloss.NewStyle().Foreground(lipgloss.Color("#6E7681")),
		ActivityHigh: lipgloss.NewStyle().Foreground(lipgloss.Color("#35D06F")).Bold(true),
		Border:       lipgloss.NewStyle().Foreground(lipgloss.Color("#30363D")),
		EmptyCell:    lipgloss.NewStyle().Foreground(lipgloss.Color("#21262D")),
		Level1:       lipgloss.NewStyle().Foreground(lipgloss.Color("#0E6B3B")),
		Level2:       lipgloss.NewStyle().Foreground(lipgloss.Color("#40C463")),
		Level3:       lipgloss.NewStyle().Foreground(lipgloss.Color("#35D06F")),
		Level4:       lipgloss.NewStyle().Foreground(lipgloss.Color("#7EE787")),
	}
}

// Theme encapsulates all UI color tokens and contribution cell styles.
type Theme struct {
	Mode         ColorMode
	Theme        ThemeMode
	Border       string
	Primary      string
	Secondary    string
	Muted        string
	ActivityHigh string // For year progress and green accent numbers
	Reset        string
	Bold         string

	// Contribution level colors
	EmptyCell          string
	ContributionLevel1 string
	ContributionLevel2 string
	ContributionLevel3 string
	ContributionLevel4 string

	// 3D Shadow styling
	ShadowBg string
}

// Styles returns Lip Gloss style definitions for this theme.
func (t Theme) Styles() Styles {
	return DefaultStyles(t)
}

// PaletteTokens is an alias for Theme to maintain backwards compatibility.
type PaletteTokens = Theme

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

	darkShadowBg   = "\033[48;2;14;65;35m"    // #0E4123
	lightShadowBg  = "\033[48;2;208;215;222m" // #D0D7DE

	// 256-color approximations
	// Dark Theme
	dark256Border         = "\033[38;5;237m"
	dark256PrimaryText    = "\033[38;5;231m"
	dark256SecondaryText  = "\033[38;5;246m"
	dark256MutedText      = "\033[38;5;243m"
	dark256EmptyCell      = "\033[38;5;236m"
	dark256ActivityLevel1 = "\033[38;5;28m"
	dark256ActivityLevel2 = "\033[38;5;35m"
	dark256ActivityLevel3 = "\033[38;5;48m"
	dark256ActivityLevel4 = "\033[38;5;120m"
	dark256ShadowBg       = "\033[48;5;22m"

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
	light256ShadowBg       = "\033[48;5;252m"

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

// GetTheme returns the complete theme tokens for the given mode and theme.
func GetTheme(mode ColorMode, theme ThemeMode) Theme {
	if mode == ColorASCII {
		return Theme{
			Mode:     ColorASCII,
			Theme:    theme,
			Reset:    "",
			Bold:     "",
			ShadowBg: "",
		}
	}

	if theme == ThemeDark {
		if mode == Color256 {
			return Theme{
				Mode:               Color256,
				Theme:              ThemeDark,
				Border:             dark256Border,
				Primary:            dark256PrimaryText,
				Secondary:          dark256SecondaryText,
				Muted:              dark256MutedText,
				ActivityHigh:       dark256ActivityLevel3,
				Reset:              reset,
				Bold:               bold,
				EmptyCell:          dark256EmptyCell,
				ContributionLevel1: dark256ActivityLevel1,
				ContributionLevel2: dark256ActivityLevel2,
				ContributionLevel3: dark256ActivityLevel3,
				ContributionLevel4: dark256ActivityLevel4,
				ShadowBg:           dark256ShadowBg,
			}
		}
		// TrueColor Dark
		return Theme{
			Mode:               ColorTrueColor,
			Theme:              ThemeDark,
			Border:             darkBorder,
			Primary:            darkPrimaryText,
			Secondary:          darkSecondaryText,
			Muted:              darkMutedText,
			ActivityHigh:       darkActivityLevel3,
			Reset:              reset,
			Bold:               bold,
			EmptyCell:          darkEmptyCell,
			ContributionLevel1: darkActivityLevel1,
			ContributionLevel2: darkActivityLevel2,
			ContributionLevel3: darkActivityLevel3,
			ContributionLevel4: darkActivityLevel4,
			ShadowBg:           darkShadowBg,
		}
	}

	// Light Theme
	if mode == Color256 {
		return Theme{
			Mode:               Color256,
			Theme:              ThemeLight,
			Border:             light256Border,
			Primary:            light256PrimaryText,
			Secondary:          light256SecondaryText,
			Muted:              light256MutedText,
			ActivityHigh:       light256ActivityLevel3,
			Reset:              reset,
			Bold:               bold,
			EmptyCell:          light256EmptyCell,
			ContributionLevel1: light256ActivityLevel1,
			ContributionLevel2: light256ActivityLevel2,
			ContributionLevel3: light256ActivityLevel3,
			ContributionLevel4: light256ActivityLevel4,
			ShadowBg:           light256ShadowBg,
		}
	}

	// TrueColor Light
	return Theme{
		Mode:               ColorTrueColor,
		Theme:              ThemeLight,
		Border:             lightBorder,
		Primary:            lightPrimaryText,
		Secondary:          lightSecondaryText,
		Muted:              lightMutedText,
		ActivityHigh:       lightActivityLevel3,
		Reset:              reset,
		Bold:               bold,
		EmptyCell:          lightEmptyCell,
		ContributionLevel1: lightActivityLevel1,
		ContributionLevel2: lightActivityLevel2,
		ContributionLevel3: lightActivityLevel3,
		ContributionLevel4: lightActivityLevel4,
		ShadowBg:           lightShadowBg,
	}
}

// GetTokens returns ANSI styling tokens for the active mode and theme.
// Retained for backward compatibility.
func GetTokens(mode ColorMode, theme ThemeMode) Theme {
	return GetTheme(mode, theme)
}

// CellColor returns the formatted cell glyph for the given mode, theme, and bucket.
func CellColor(mode ColorMode, theme ThemeMode, bucket int) string {
	t := GetTheme(mode, theme)
	return t.CellColor(bucket)
}

// CellColor returns the formatted cell glyph for this theme.
func (t Theme) CellColor(bucket int) string {
	if bucket < 0 {
		bucket = 0
	}
	if bucket > 4 {
		bucket = 4
	}

	if t.Mode == ColorASCII {
		glyph := asciiGlyphs[bucket]
		if os.Getenv("NO_COLOR") != "" {
			glyph = pureAsciiGlyphs[bucket]
		}
		return glyph
	}

	block := "■"
	var color string
	switch bucket {
	case 0:
		color = t.EmptyCell
	case 1:
		color = t.ContributionLevel1
	case 2:
		color = t.ContributionLevel2
	case 3:
		color = t.ContributionLevel3
	case 4:
		color = t.ContributionLevel4
	}

	return color + block + t.Reset
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
