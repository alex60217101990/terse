//go:build linux

package hook

import "syscall"

// ctimeFromStat extracts the inode change time from a linux syscall.Stat_t.
func ctimeFromStat(st *syscall.Stat_t) int64 {
	return st.Ctim.Sec*1e9 + st.Ctim.Nsec
}
