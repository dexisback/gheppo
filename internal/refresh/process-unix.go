//go:build !windows

package refresh

import "syscall"

// detachedProcessAttributes makes the child independent of the
// parent's process group/session so it survives after Gheppo exits.
func detachedProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}