//go:build !windows

package cmd

// cleanupUserPath is a no-op on non-Windows platforms: the persistent PATH
// entry added by scripts/install.sh lives in the shell rc file and is removed
// along with the shell integration block.
func cleanupUserPath() error { return nil }
