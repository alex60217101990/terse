//go:build !windows

package daemon

import "syscall"

// detachSysProcAttr detaches the spawned daemon from the hook's session so it
// outlives the short-lived hook process (Setsid = new session, no controlling
// terminal).
func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
