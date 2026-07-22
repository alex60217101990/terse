package cache

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GCResult holds the outcome of a GC run.
type GCResult struct {
	Removed    int
	Kept       int
	FreedBytes int64
}

// RunGC scans all session files in StateDir and removes those whose utility
// score falls below minScore. If dryRun is true, files are reported but not
// deleted. Corrupt or unreadable files are also evicted (counted as Removed).
func RunGC(dryRun bool, minScore float64) (GCResult, error) {
	dir := StateDir()
	var result GCResult
	nowSec := time.Now().Unix()
	lambda := DecayLambda()

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".qdf") {
			return nil
		}

		info, _ := d.Info()
		sessionID := strings.TrimSuffix(filepath.Base(path), ".qdf")
		state, loadErr := Load(sessionID)

		var sessionScore float64
		if loadErr != nil || state == nil {
			sessionScore = 0 // corrupt → evict
		} else {
			var totalReads int
			var lastRead int64
			for _, e := range state.Files {
				totalReads += e.ReadCount
				if e.LastReadAt > lastRead {
					lastRead = e.LastReadAt
				}
			}
			if lastRead == 0 {
				lastRead = nowSec - 86400 // default: 1 day old
			}
			ageHours := float64(nowSec-lastRead) / 3600.0
			if ageHours < 0 {
				ageHours = 0 // clock skew (NTP step back) must not inflate the score
			}
			sessionScore = float64(totalReads) * math.Exp(-lambda*ageHours)
		}

		if sessionScore < minScore {
			if dryRun {
				fmt.Printf("[dry-run] would remove %s (score=%.4f)\n", filepath.Base(path), sessionScore)
			} else {
				_ = os.Remove(path)
				if info != nil {
					result.FreedBytes += info.Size()
				}
			}
			result.Removed++
		} else {
			result.Kept++
		}
		return nil
	})
	return result, err
}
