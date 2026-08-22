package cache

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
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

	// Captures are their own store with their own bound; a manual gc that left
	// them alone would report a tidy cache while the biggest directory grew.
	// Nothing to dry-run against: GCCaptures reports only what it deleted, so
	// a dry run skips it rather than pretending.
	if !dryRun {
		rm, freed := GCCaptures(CaptureTTL())
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
	_, _ = GCCaptures(CaptureTTL())
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

// CaptureMaxSize caps the total bytes parked in captures/. A capture is only
// worth keeping until the model asks for the elided output back, which happens
// within the turn or not at all, so this is deliberately smaller than the
// refs/last cap: a heavy day can produce thousands of them.
const CaptureMaxSize = 32 << 20 // 32 MiB

// DefaultCaptureTTL is how long a capture stays recoverable. A capped view is
// asked about within the turn that produced it or never, so this is a day, not
// the 30 the blob stores keep.
const DefaultCaptureTTL = 24 * time.Hour

// CaptureTTL is the capture age limit: env QDF_CAPTURE_TTL (a Go duration), or
// DefaultCaptureTTL when unset or unparseable.
func CaptureTTL() time.Duration {
	if v := os.Getenv("QDF_CAPTURE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return DefaultCaptureTTL
}

// GCCaptures prunes captures/ by age and then by size, and reports what it
// removed. Age alone is not a bound: a heavy session can write thousands of
// captures inside one TTL window, so whatever is left after the age pass is
// trimmed oldest-first until it fits CaptureMaxSize.
func GCCaptures(maxAge time.Duration) (removed int, freed int64) {
	dir := CaptureDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0
	}
	type capture struct {
		path string
		size int64
		mod  int64
	}
	kept := make([]capture, 0, len(entries))
	var total int64
	cutoff := time.Now().Add(-maxAge)
	for _, e := range entries {
		info, ierr := e.Info()
		if ierr != nil || e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if info.ModTime().Before(cutoff) {
			if os.Remove(path) == nil {
				removed++
				freed += info.Size()
			}
			continue
		}
		kept = append(kept, capture{path, info.Size(), info.ModTime().UnixNano()})
		total += info.Size()
	}
	if total <= CaptureMaxSize {
		return removed, freed
	}
	// Oldest first: the newest captures are the ones a running turn might still
	// come back for.
	sort.Slice(kept, func(i, j int) bool { return kept[i].mod < kept[j].mod })
	for _, c := range kept {
		if total <= CaptureMaxSize {
			break
		}
		if os.Remove(c.path) == nil {
			removed++
			freed += c.size
			total -= c.size
		}
	}
	return removed, freed
}
