package cmd

import "strings"

// removePathEntry removes dir from a semicolon-delimited PATH value (the
// Windows registry format). Entries are trimmed and compared
// case-insensitively, since Windows paths are case-insensitive. It reports
// whether the value changed.
func removePathEntry(pathValue, dir string) (string, bool) {
	entries := strings.Split(pathValue, ";")
	kept := make([]string, 0, len(entries))
	changed := false
	for _, entry := range entries {
		if strings.EqualFold(strings.TrimSpace(entry), dir) {
			changed = true
			continue
		}
		kept = append(kept, entry)
	}
	if !changed {
		return pathValue, false
	}
	return strings.Join(kept, ";"), true
}
