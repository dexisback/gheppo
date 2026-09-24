//go:build windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// cleanupUserPath removes the Gheppo install directory from the Windows user
// PATH (HKCU\Environment\Path), reversing the entry that scripts/install.sh
// adds when the user confirms the PATH prompt. It preserves the original
// registry value kind (REG_SZ vs REG_EXPAND_SZ) so existing %VAR% entries
// keep expanding.
func cleanupUserPath() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating Gheppo executable: %w", err)
	}
	dir := filepath.Dir(executable)

	key, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("opening user environment key: %w", err)
	}
	defer key.Close()

	value, valueType, err := key.GetStringValue("Path")
	if err != nil {
		// No user PATH value (or unreadable): nothing to clean up.
		return nil
	}

	updated, changed := removePathEntry(value, dir)
	if !changed {
		return nil
	}

	if valueType == registry.EXPAND_SZ {
		return key.SetExpandStringValue("Path", updated)
	}
	return key.SetStringValue("Path", updated)
}
