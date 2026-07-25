//go:build windows

package hook

import "os"

// statCtimeNS returns 0 on platforms without a POSIX ctime; the deny gate then
// degrades to mtime+size (unchanged from prior behavior).
func statCtimeNS(fi os.FileInfo) int64 { return 0 }
