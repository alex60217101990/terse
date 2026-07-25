//go:build !windows

package hook

import (
	"os"
	"syscall"
)

// statCtimeNS returns the file's inode change time (ctime) in nanoseconds.
// ctime advances on every content and metadata change and cannot be moved
// backward from userspace, so it detects content changes that forge mtime
// (cp -p, touch -r, rsync --times).
func statCtimeNS(fi os.FileInfo) int64 {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0
	}
	return ctimeFromStat(st) // Ctim (linux) / Ctimespec (darwin) — see below
}
