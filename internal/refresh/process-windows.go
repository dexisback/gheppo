//go:build windows

package refresh

import "syscall"

func detachedProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP |
			syscall.DETACHED_PROCESS |
			syscall.CREATE_NO_WINDOW,
	}
}