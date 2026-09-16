package render

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dexisback/gheppo/internal/config"
)

// ColorMode specifies the terminal color capability level.
type ColorMode int

const (
	ColorTrueColor ColorMode = iota
	Color256
	ColorASCII
)

// ThemeMode specifies the terminal theme mode (retained for backward compatibility).
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

	fg := theme.Config.Foreground
	if fg == "" {
		fg = "#F0F6FC"
	}
	muted := theme.Config.Muted
	if muted == "" {
		muted = "#8B949E"
	}
	accent := theme.Config.Accent
	if accent == "" {
		accent = theme.Config.GraphLevels[4]
	}
	border := theme.Config.Border
	if border == "" {
		border = "#30363D"
	}

	return Styles{
		Primary:      lipgloss.NewStyle().Foreground(lipgloss.Color(fg)).Bold(true),
		Secondary:    lipgloss.NewStyle().Foreground(lipgloss.Color(muted)),
		Muted:        lipgloss.NewStyle().Foreground(lipgloss.Color(muted)),
		ActivityHigh: lipgloss.NewStyle().Foreground(lipgloss.Color(accent)).Bold(true),
		Border:       lipgloss.NewStyle().Foreground(lipgloss.Color(border)),
		EmptyCell:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Config.GraphLevels[0])),
		Level1:       lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Config.GraphLevels[1])),
		Level2:       lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Config.GraphLevels[2])),
		Level3:       lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Config.GraphLevels[3])),
		Level4:       lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Config.GraphLevels[4])),
	}
}

// Theme encapsulates resolved ANSI color tokens and contribution cell styles derived from config.Theme.
type Theme struct {
	Config       config.Theme
	Mode         ColorMode
	Theme        ThemeMode
	Border       string
	Primary      string
	Secondary    string
	Muted        string
	ActivityHigh string // Accent color for high-visibility highlights
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

const (
	resetAnsi = "\033[0m"
	boldAnsi  = "\033[1m"
)

// ASCII intensity glyphs (NO_COLOR)
var asciiGlyphs = map[int]string{
	0: "·",
	1: "░",
	2: "▒",
	3: "▓",
	4: "█",
}

// Pure ASCII fallback
var pureAsciiGlyphs = map[int]string{
	0: ".",
	1: "-",
	2: "=",
	3: "+",
	4: "#",
}

// ResolveTheme maps a config.Theme and ColorMode into a resolved presentation Theme.
func ResolveTheme(cfgTheme config.Theme, mode ColorMode) Theme {
	// If empty theme config was provided, fallback to default
	if cfgTheme.Name == "" {
		cfgTheme = config.DefaultTheme()
	}

	if mode == ColorASCII {
		return Theme{
			Config: cfgTheme,
			Mode:   ColorASCII,
			Reset:  "",
			Bold:   "",
		}
	}

	if mode == Color256 {
		return Theme{
			Config:             cfgTheme,
			Mode:               Color256,
			Border:             hexTo256ColorFg(cfgTheme.Border),
			Primary:            hexTo256ColorFg(cfgTheme.Foreground),
			Secondary:          hexTo256ColorFg(cfgTheme.Muted),
			Muted:              hexTo256ColorFg(cfgTheme.Muted),
			ActivityHigh:       hexTo256ColorFg(cfgTheme.Accent),
			Reset:              resetAnsi,
			Bold:               boldAnsi,
			EmptyCell:          hexTo256ColorFg(cfgTheme.GraphLevels[0]),
			ContributionLevel1: hexTo256ColorFg(cfgTheme.GraphLevels[1]),
			ContributionLevel2: hexTo256ColorFg(cfgTheme.GraphLevels[2]),
			ContributionLevel3: hexTo256ColorFg(cfgTheme.GraphLevels[3]),
			ContributionLevel4: hexTo256ColorFg(cfgTheme.GraphLevels[4]),
			ShadowBg:           hexTo256ColorBg(cfgTheme.GraphLevels[1]),
		}
	}

	// ColorTrueColor (default)
	return Theme{
		Config:             cfgTheme,
		Mode:               ColorTrueColor,
		Border:             hexToTrueColorFg(cfgTheme.Border),
		Primary:            hexToTrueColorFg(cfgTheme.Foreground),
		Secondary:          hexToTrueColorFg(cfgTheme.Muted),
		Muted:              hexToTrueColorFg(cfgTheme.Muted),
		ActivityHigh:       hexToTrueColorFg(cfgTheme.Accent),
		Reset:              resetAnsi,
		Bold:               boldAnsi,
		EmptyCell:          hexToTrueColorFg(cfgTheme.GraphLevels[0]),
		ContributionLevel1: hexToTrueColorFg(cfgTheme.GraphLevels[1]),
		ContributionLevel2: hexToTrueColorFg(cfgTheme.GraphLevels[2]),
		ContributionLevel3: hexToTrueColorFg(cfgTheme.GraphLevels[3]),
		ContributionLevel4: hexToTrueColorFg(cfgTheme.GraphLevels[4]),
		ShadowBg:           hexToTrueColorBg(cfgTheme.GraphLevels[1]),
	}
}

// GetTheme returns the resolved Theme for the active configuration and given color mode.
func GetTheme(mode ColorMode, _ ...ThemeMode) Theme {
	cfgTheme := config.GetSelectedTheme()
	return ResolveTheme(cfgTheme, mode)
}

// GetTokens returns ANSI styling tokens for the active mode and theme.
func GetTokens(mode ColorMode, theme ThemeMode) Theme {
	return GetTheme(mode, theme)
}

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

func parseHexColor(hex string) (r, g, b uint8, ok bool) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0, 0, 0, false
	}
	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return uint8(val >> 16), uint8((val >> 8) & 0xFF), uint8(val & 0xFF), true
}

func hexToTrueColorFg(hex string) string {
	r, g, b, ok := parseHexColor(hex)
	if !ok {
		return ""
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

func hexToTrueColorBg(hex string) string {
	r, g, b, ok := parseHexColor(hex)
	if !ok {
		return ""
	}
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

func rgbTo256Index(r, g, b uint8) int {
	if r == g && g == b {
		if r < 8 {
			return 16
		}
		if r > 248 {
			return 231
		}
		return 232 + int((float64(r)-8)/247*24)
	}
	r6 := int(r) * 5 / 255
	g6 := int(g) * 5 / 255
	b6 := int(b) * 5 / 255
	return 16 + 36*r6 + 6*g6 + b6
}

func hexTo256ColorFg(hex string) string {
	r, g, b, ok := parseHexColor(hex)
	if !ok {
		return ""
	}
	idx := rgbTo256Index(r, g, b)
	return fmt.Sprintf("\033[38;5;%dm", idx)
}

func hexTo256ColorBg(hex string) string {
	r, g, b, ok := parseHexColor(hex)
	if !ok {
		return ""
	}
	idx := rgbTo256Index(r, g, b)
	return fmt.Sprintf("\033[48;5;%dm", idx)
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
