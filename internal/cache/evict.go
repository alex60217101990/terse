package cache

import (
	"cmp"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// DecayLambda returns the exponential-decay rate λ used by UtilityScore.
// Reads QDF_DECAY_LAMBDA env var; defaults to 0.1 (half-life ≈ 7 h).
func DecayLambda() float64 {
	if v := os.Getenv("QDF_DECAY_LAMBDA"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.1
}

// UtilityScore computes a combined recency+frequency score for a FileEntry.
// Higher = more valuable to keep. Uses exponential decay: ReadCount * e^(-λ * ageHours).
func UtilityScore(entry FileEntry, nowSec int64) float64 {
	return utilityScore(entry, nowSec, DecayLambda())
}

// utilityScore is UtilityScore with the decay rate passed in, so a caller
// scoring many entries in a loop reads QDF_DECAY_LAMBDA once (hoisted) instead
// of re-parsing the env per entry.
func utilityScore(entry FileEntry, nowSec int64, lambda float64) float64 {
	ageHours := float64(nowSec-entry.LastReadAt) / 3600.0
	if ageHours < 0 {
		ageHours = 0
	}
	return float64(entry.ReadCount) * math.Exp(-lambda*ageHours)
}

// Evict removes lowest-utility entries from state when len(Files) > maxFiles.
// Drops bottom entries until the map is at 80% of maxFiles.
// Called automatically by Save.
func Evict(state *SessionState, maxFiles int) {
	if len(state.Files) <= maxFiles {
		return
	}
	type scored struct {
		path      string
		score     float64
		readCount int
	}
	now := state.nowSec()
	lambda := DecayLambda() // hoisted: parse the env once, not per entry
	entries := make([]scored, 0, len(state.Files))
	for path, e := range state.Files {
		entries = append(entries, scored{path, utilityScore(e, now, lambda), e.ReadCount})
	}
	// Sort ascending by score; break ties by ReadCount ascending so
	// zero-read entries are always evicted before frequently-read ones.
	slices.SortFunc(entries, func(a, b scored) int {
		if c := cmp.Compare(a.score, b.score); c != 0 {
			return c
		}
		return cmp.Compare(a.readCount, b.readCount)
	})
	// Drop lowest-scoring entries until we're at 80% of maxFiles.
	target := maxFiles * 4 / 5
	for i := range entries {
		if len(state.Files) <= target {
			break
		}
		delete(state.Files, entries[i].path)
	}
}

// CacheScore is the recency×frequency utility of a cache entry, analogous to
// UtilityScore for sessions: Hits * e^(-lambda * ageHours). A never-hit entry
// (Hits 0) scores 0 and is evicted first.
func CacheScore(s UsageStat, nowSec int64) float64 {
	return cacheScore(s, nowSec, DecayLambda())
}

// cacheScore is CacheScore with the decay rate passed in (hoisted out of the
// PruneDir loop so the env is parsed once, not per blob).
func cacheScore(s UsageStat, nowSec int64, lambda float64) float64 {
	ageHours := float64(nowSec-s.LastUsed) / 3600.0
	if ageHours < 0 {
		ageHours = 0 // clock skew must not inflate the score
	}
	return float64(s.Hits) * math.Exp(-lambda*ageHours)
}

// PruneDir bounds one blob directory. It drops entries older than ttl
// (LastUsed if known, else file mtime) unconditionally, then — if the total
// remaining bytes exceed maxSize — evicts the lowest CacheScore entries until
// the total is at most 80% of maxSize. Blob files are deleted and their usage
// entries removed; the pruned usage index is written back. Best-effort:
// per-file errors are skipped, never fatal.
func PruneDir(blobDir, usagePath string, maxSize int64, ttl time.Duration, nowSec int64, dryRun bool) (removed int, freed int64) {
	ents, err := os.ReadDir(blobDir)
	if err != nil {
		return 0, 0
	}
	idx := LoadUsage(usagePath)

	type blob struct {
		key   string // hash / last key (filename without ".blob")
		path  string
		size  int64
		score float64
		used  int64 // LastUsed or mtime, unix seconds
	}
	var blobs []blob
	var total int64
	ttlSec := int64(ttl / time.Second)
	lambda := DecayLambda() // hoisted: parse the env once, not per blob

	for _, de := range ents {
		name := de.Name()
		if de.IsDir() || !strings.HasSuffix(name, ".blob") {
			continue
		}
		info, ierr := de.Info()
		if ierr != nil {
			continue
		}
		key := strings.TrimSuffix(name, ".blob")
		used := info.ModTime().Unix()
		if u, ok := idx[key]; ok && u.LastUsed > 0 {
			used = u.LastUsed
		}
		blobs = append(blobs, blob{
			key:   key,
			path:  filepath.Join(blobDir, name),
			size:  info.Size(),
			score: cacheScore(idx[key], nowSec, lambda),
			used:  used,
		})
		total += info.Size()
	}

	drop := func(b blob) {
		if !dryRun {
			if err := os.Remove(b.path); err != nil {
				return // couldn't remove — don't count it or advance the total
			}
			delete(idx, b.key)
		}
		// total is decremented in dry-run too, so the size-cap loop's
		// `total <= target` break reflects what WOULD be freed — otherwise a
		// dry run reports evicting every blob instead of just down to target.
		removed++
		freed += b.size
		total -= b.size
	}

	// 1) TTL floor: drop anything older than ttl.
	kept := blobs[:0]
	for _, b := range blobs {
		if nowSec-b.used > ttlSec {
			drop(b)
		} else {
			kept = append(kept, b)
		}
	}
	blobs = kept

	// 2) Size cap: if still over, evict lowest score until <= 80% of cap.
	if total > maxSize {
		slices.SortFunc(blobs, func(a, b blob) int {
			if c := cmp.Compare(a.score, b.score); c != 0 {
				return c // ascending: lowest first
			}
			return cmp.Compare(a.used, b.used) // tie-break: oldest first
		})
		target := maxSize * 4 / 5
		for _, b := range blobs {
			if total <= target {
				break
			}
			drop(b)
		}
	}

	if !dryRun && removed > 0 {
		_ = SaveUsage(usagePath, idx)
	}
	return removed, freed
}
