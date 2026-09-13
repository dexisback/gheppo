//go:build windows

package refresh

import "syscall"

const (
	detachedProcess = 0x00000008
	createNoWindow   = 0x08000000
)

func detachedProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP |
			detachedProcess |
			createNoWindow,
	}
}