package cache

import (
	"os"
	"path/filepath"
	"strconv"
)

// gcInterval is the minimum spacing between automatic gc runs.
const gcInterval = int64(24 * 3600) // seconds

// gcStampPath returns the on-disk marker recording when automatic gc last ran.
func gcStampPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".qdf-hook", ".gc-stamp")
}

// ShouldRunGC reports whether an automatic gc is due (>= 24h since the last),
// and if so records nowSec as the new stamp. Best-effort: an unreadable or
// missing stamp is treated as "due".
func ShouldRunGC(nowSec int64) bool {
	p := gcStampPath()
	if data, err := os.ReadFile(p); err == nil {
		if last, perr := strconv.ParseInt(string(data), 10, 64); perr == nil && nowSec-last < gcInterval {
			return false
		}
	}
	_ = writeFileLazy(p, []byte(strconv.FormatInt(nowSec, 10)))
	return true
}
