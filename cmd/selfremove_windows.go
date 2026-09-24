//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// removeExecutable deletes the binary from within the binary itself. While
// gheppo.exe is running, Windows keeps its image mapped and refuses the
// delete with a sharing violation, but a rename (even of a locked file) is
// allowed. So: rename the running file to a scratch name in the same
// directory, then spawn a detached cmd.exe that waits ~2 seconds (letting
// this process exit and unmap the image) before deleting the renamed file.
func removeExecutable(path string) error {
	// Fast path: the file is not locked (e.g. invoked from a copy), so a
	// plain delete just works.
	err := os.Remove(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}

	renamed := path + fmt.Sprintf(".%d.old", os.Getpid())
	if renameErr := os.Rename(path, renamed); renameErr != nil {
		// Report the original delete error, which is the more useful one.
		return err
	}

	c := exec.Command("cmd.exe", "/C", "ping", "-n", "3", "127.0.0.1", "&&", "del", "/F", "/Q", renamed)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := c.Start(); err != nil {
		return fmt.Errorf("scheduling removal of %s: %w", renamed, err)
	}
	return nil
}
