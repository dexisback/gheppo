package config

import (
	"strings"
)

// Theme defines the visual design tokens for Gheppo presentation.
type Theme struct {
	Name        string    `json:"name"`
	Background  string    `json:"background"`
	Foreground  string    `json:"foreground"`
	Muted       string    `json:"muted"`
	Border      string    `json:"border,omitempty"`
	GraphEmpty  string    `json:"graphEmpty"`
	GraphLevels [5]string `json:"graphLevels"`
	Accent      string    `json:"accent"`
}

// Built-in theme registry containing the 5 initial themes.
var registry = map[string]Theme{
	"github": {
		Name:        "github",
		Background:  "#0D1117",
		Foreground:  "#F0F6FC",
		Muted:       "#8B949E",
		Border:      "#30363D",
		GraphEmpty:  "#161B22",
		GraphLevels: [5]string{"#161B22", "#0E4429", "#006D32", "#26A641", "#39D353"},
		Accent:      "#39D353",
	},
	"mono": {
		Name:        "mono",
		Background:  "#121212",
		Foreground:  "#F2F2F2",
		Muted:       "#969696",
		Border:      "#353535",
		GraphEmpty:  "#1C1C1C",
		GraphLevels: [5]string{"#1C1C1C", "#353535", "#606060", "#969696", "#F2F2F2"},
		Accent:      "#F2F2F2",
	},
	"catppuccin": {
		Name:        "catppuccin",
		Background:  "#1E1E2E",
		Foreground:  "#CDD6F4",
		Muted:       "#A6ADC8",
		Border:      "#45475A",
		GraphEmpty:  "#1E1E2E",
		GraphLevels: [5]string{"#1E1E2E", "#3D3150", "#624477", "#9168B5", "#CBA6F7"},
		Accent:      "#CBA6F7",
	},
	"nord": {
		Name:        "nord",
		Background:  "#2E3440",
		Foreground:  "#D8DEE9",
		Muted:       "#AAB4C3",
		Border:      "#4C566A",
		GraphEmpty:  "#2E3440",
		GraphLevels: [5]string{"#2E3440", "#3B4252", "#4C566A", "#5E81AC", "#88C0D0"},
		Accent:      "#88C0D0",
	},
	"gruvbox": {
		Name:        "gruvbox",
		Background:  "#282828",
		Foreground:  "#EBDBB2",
		Muted:       "#A89984",
		Border:      "#504945",
		GraphEmpty:  "#282828",
		GraphLevels: [5]string{"#282828", "#7A2810", "#AF3A03", "#D65D0E", "#FE8019"},
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
