//go:build windows

package herdr

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func detachAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
	}
}
