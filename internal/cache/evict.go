package cache

import (
	"math"
	"os"
	"sort"
	"strconv"
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
	ageHours := float64(nowSec-entry.LastReadAt) / 3600.0
	if ageHours < 0 {
		ageHours = 0
	}
	return float64(entry.ReadCount) * math.Exp(-DecayLambda()*ageHours)
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
	entries := make([]scored, 0, len(state.Files))
	for path, e := range state.Files {
		entries = append(entries, scored{path, UtilityScore(e, now), e.ReadCount})
	}
	// Sort ascending by score; break ties by ReadCount ascending so
	// zero-read entries are always evicted before frequently-read ones.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].score != entries[j].score {
			return entries[i].score < entries[j].score
		}
		return entries[i].readCount < entries[j].readCount
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
