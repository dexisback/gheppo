package render

// Typographic Wordmark for Gheppo
const (
	WordmarkUnicode = "ɢ ʜ ᴇ ᴘ ᴘ ᴏ"
	WordmarkASCII   = "G H E P P O"
)

// GetWordmark returns the styled wordmark for the active color mode.
func GetWordmark(mode ColorMode) string {
	if mode == ColorASCII {
		return WordmarkASCII
	}
	return WordmarkUnicode
}

// GetWordmarkLines returns the wordmark as lines for compatibility.
func GetWordmarkLines(mode ColorMode) []string {
	return []string{GetWordmark(mode)}
}

