//go:build !windows

package cmd

import "os"

// removeExecutable deletes the running binary directly. Non-Windows systems
// allow unlinking an executable while it is mapped, so no special handling
// is required.
func removeExecutable(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
