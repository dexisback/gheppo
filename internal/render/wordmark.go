package render

import (
	"unicode/utf8"
)

// Gheppo trace wordmark following the design spec.
// The wordmark has a "trace/path" aesthetic with dots extending leftward.

const (
	// Unicode version with trace aesthetic
	// Compact, geometric, terminal-native
	wordmarkUnicode = "········ G H E P P O"

	// ASCII fallback
	wordmarkASCII = "........ G H E P P O"
)

// WordmarkState captures animation or visual phase state for future wordmark effects.
type WordmarkState struct {
	Progress float64 // 0.0 to 1.0
}

// GetWordmark returns the complete wordmark with trace for the active color mode.
func GetWordmark(mode ColorMode) string {
	if mode == ColorASCII {
		return wordmarkASCII
	}
	return wordmarkUnicode
}

// RenderWordmark returns the styled wordmark string and its visible rune length.
func RenderWordmark(theme Theme) (string, int) {
	return RenderWordmarkWithState(theme, WordmarkState{Progress: 1.0})
}

// RenderWordmarkWithState allows rendering the wordmark with an external animation progress state.
func RenderWordmarkWithState(theme Theme, state WordmarkState) (string, int) {
	raw := GetWordmark(theme.Mode)
	visLen := utf8.RuneCountInString(raw)

	if theme.Mode == ColorASCII {
		return raw, visLen
	}

	colored := theme.Primary + raw + theme.Reset
	return colored, visLen
}
