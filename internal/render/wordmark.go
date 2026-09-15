package render

// Gheppo trace wordmark following the design spec
// The wordmark should have a "trace/path" aesthetic with the G extending leftward

// GetWordmark returns the complete wordmark with trace for the active color mode.
func GetWordmark(mode ColorMode) string {
	if mode == ColorASCII {
		return wordmarkASCII
	}
	return wordmarkUnicode
}

const (
	// Unicode version with trace aesthetic
	// Compact, geometric, terminal-native
	wordmarkUnicode = "········ G H E P P O"

	// ASCII fallback
	wordmarkASCII = "........ G H E P P O"
)
