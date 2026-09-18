package config

import (
	"strings"
)

// Theme defines the visual design tokens for Gheppo presentation.
type Theme struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName,omitempty"`
	Description string    `json:"description,omitempty"`
	Background  string    `json:"background"`
	Foreground  string    `json:"foreground"`
	Muted       string    `json:"muted"`
	Border      string    `json:"border,omitempty"`
	GraphEmpty  string    `json:"graphEmpty"`
	GraphLevels [5]string `json:"graphLevels"`
	Accent      string    `json:"accent"`
}

// Title returns DisplayName if set, otherwise the capitalized Name.
func (t Theme) Title() string {
	if t.DisplayName != "" {
		return t.DisplayName
	}
	if len(t.Name) > 0 {
		return strings.ToUpper(t.Name[:1]) + t.Name[1:]
	}
	return t.Name
}

// Built-in theme registry containing the 5 initial themes.
var registry = map[string]Theme{
	"github": {
		Name:        "github",
		DisplayName: "GitHub",
		Description: "The classic GitHub contribution graph palette.",
		Background:  "#0D1117",
		Foreground:  "#F0F6FC",
		Muted:       "#8B949E",
		Border:      "#30363D",
		GraphEmpty:  "#21262D",
		GraphLevels: [5]string{"#21262D", "#0E4429", "#006D32", "#26A641", "#39D353"},
		Accent:      "#39D353",
	},
	"mono": {
		Name:        "mono",
		DisplayName: "Mono",
		Description: "High-contrast monochrome grayscale palette.",
		Background:  "#121212",
		Foreground:  "#F2F2F2",
		Muted:       "#969696",
		Border:      "#353535",
		GraphEmpty:  "#282828",
		GraphLevels: [5]string{"#282828", "#454545", "#707070", "#A2A2A2", "#F2F2F2"},
		Accent:      "#F2F2F2",
	},
	"catppuccin": {
		Name:        "catppuccin",
		DisplayName: "Catppuccin",
		Description: "Soothing pastel palette for the terminal.",
		Background:  "#1E1E2E",
		Foreground:  "#CDD6F4",
		Muted:       "#A6ADC8",
		Border:      "#45475A",
		GraphEmpty:  "#313244",
		GraphLevels: [5]string{"#313244", "#45355E", "#624477", "#9168B5", "#CBA6F7"},
		Accent:      "#CBA6F7",
	},
	"nord": {
		Name:        "nord",
		DisplayName: "Nord",
		Description: "A clean and arctic-inspired palette for the terminal.",
		Background:  "#2E3440",
		Foreground:  "#D8DEE9",
		Muted:       "#AAB4C3",
		Border:      "#4C566A",
		GraphEmpty:  "#3B4252",
		GraphLevels: [5]string{"#3B4252", "#4C6584", "#5E81AC", "#81A1C1", "#88C0D0"},
		Accent:      "#88C0D0",
	},
	"gruvbox": {
		Name:        "gruvbox",
		DisplayName: "Gruvbox",
		Description: "Retro groove warm autumn palette with rich contrast.",
		Background:  "#282828",
		Foreground:  "#EBDBB2",
		Muted:       "#A89984",
		Border:      "#504945",
		GraphEmpty:  "#3C3836",
		GraphLevels: [5]string{"#3C3836", "#7A2810", "#AF3A03", "#D65D0E", "#FE8019"},
		Accent:      "#FE8019",
	},
}

// Ordered list of built-in theme identifiers.
var themeOrder = []string{"github", "mono", "catppuccin", "nord", "gruvbox"}

// DefaultThemeName is the fallback theme when none is configured.
const DefaultThemeName = "github"

// DefaultTheme returns the default GitHub theme definition.
func DefaultTheme() Theme {
	return registry[DefaultThemeName]
}

// GetTheme retrieves a theme by name (case-insensitive).
func GetTheme(name string) (Theme, bool) {
	key := strings.ToLower(strings.TrimSpace(name))
	t, ok := registry[key]
	return t, ok
}

// ResolveTheme retrieves a theme by name or returns the default theme if not found or empty.
func ResolveTheme(name string) Theme {
	if t, ok := GetTheme(name); ok {
		return t
	}
	return DefaultTheme()
}

// ListThemes returns all registered themes in canonical order.
func ListThemes() []Theme {
	themes := make([]Theme, 0, len(themeOrder))
	for _, name := range themeOrder {
		if t, ok := registry[name]; ok {
			themes = append(themes, t)
		}
	}
	return themes
}

// AvailableThemeNames returns the names of all registered themes in canonical order.
func AvailableThemeNames() []string {
	names := make([]string, len(themeOrder))
	copy(names, themeOrder)
	return names
}

// IsValidTheme reports whether a theme name exists in the registry.
func IsValidTheme(name string) bool {
	_, ok := GetTheme(name)
	return ok
}
