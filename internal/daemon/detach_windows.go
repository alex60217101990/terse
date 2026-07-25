//go:build windows

package daemon

import "syscall"

// detachSysProcAttr detaches the spawned daemon on Windows: a new process
// group + DETACHED_PROCESS so it has no console and is not killed when the
// hook process exits.
func detachSysProcAttr() *syscall.SysProcAttr {
	const (
		createNewProcessGroup = 0x00000200
		detachedProcess       = 0x00000008
	)
	return &syscall.SysProcAttr{CreationFlags: createNewProcessGroup | detachedProcess}
}
