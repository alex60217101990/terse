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
	Removed        int
	Kept           int
	FreedBytes     int64
	BlobsRemoved   int   `json:"blobs_removed"`
	BlobBytesFreed int64 `json:"blob_bytes_freed"`
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
			return nil //nolint:nilerr // GC is best-effort: skip unreadable entries, don't abort the walk
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

	// Prune refs/ and last/ blob stores by combined size cap + TTL, splitting
	// the cap evenly across the two dirs (each bounded independently keeps
	// one from starving the other; the combined ceiling is still maxSize).
	nowSec2 := time.Now().Unix()
	maxSize := CacheMaxSize()
	ttl := CacheTTL()
	half := maxSize / 2
	for _, d := range []struct{ dir, usage string }{
		{RefsDir(), UsageRefsPath()},
		{LastOutDir(), UsageLastPath()},
	} {
		rm, freed := PruneDir(d.dir, d.usage, half, ttl, nowSec2, dryRun)
		result.BlobsRemoved += rm
		result.BlobBytesFreed += freed
	}

	return result, err
}

// SweepBlobs prunes the refs/ and last/ blob stores to the configured size cap
// (split evenly across the two dirs) and TTL. It is blob-only: it never touches
// session .qdf state, so — unlike RunGC — it is safe to call from a hook on the
// current session without any risk of evicting the state that request just
// wrote or is about to read. Best-effort; errors are swallowed inside PruneDir.
func SweepBlobs(nowSec int64) {
	half := CacheMaxSize() / 2
	ttl := CacheTTL()
	PruneDir(RefsDir(), UsageRefsPath(), half, ttl, nowSec, false)
	PruneDir(LastOutDir(), UsageLastPath(), half, ttl, nowSec, false)
	GCCaptures(24 * time.Hour)
}

// AutoSweep runs a throttled blob prune: it calls SweepBlobs at most once per
// gc interval (gated by ShouldRunGC's timestamp stamp). Wired into the
// installed SessionStart hook (qdf-hook daemon --ensure) so the disk cache is
// bounded on every session start even when no daemon ends up running — for a
// CLI-only user the daemon sweep ticker never fires, so this is their only
// automatic prune. Best-effort; never fails the caller.
func AutoSweep(nowSec int64) {
	if ShouldRunGC(nowSec) {
		SweepBlobs(nowSec)
	}
}

// GCCaptures deletes capture files older than maxAge. A capture only exists so a
// capped view stays recoverable; once the session that produced it is gone,
// nothing can ask for it again.
func GCCaptures(maxAge time.Duration) {
	entries, err := os.ReadDir(CaptureDir())
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(CaptureDir(), e.Name()))
	}
}
