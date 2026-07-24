//go:build darwin

package hook

import "syscall"

// ctimeFromStat extracts the inode change time from a darwin syscall.Stat_t.
func ctimeFromStat(st *syscall.Stat_t) int64 {
	return st.Ctimespec.Sec*1e9 + st.Ctimespec.Nsec
}
