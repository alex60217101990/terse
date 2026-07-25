package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/cache"
)

// writeBlob makes a fake blob file of n bytes with a given mtime.
func writeBlob(t *testing.T, dir, name string, n int, mtime time.Time) string {
	t.Helper()
	p := filepath.Join(dir, name+".blob")
	if err := os.WriteFile(p, make([]byte, n), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPruneDir_TTLFloorDropsOld(t *testing.T) {
	dir := t.TempDir()
	usage := filepath.Join(t.TempDir(), "u.qdf")
	now := time.Now()
	old := writeBlob(t, dir, "old", 100, now.Add(-1000*time.Hour))
	fresh := writeBlob(t, dir, "fresh", 100, now)

	removed, _ := cache.PruneDir(dir, usage, 1<<30 /*huge cap*/, 720*time.Hour, now.Unix(), false)
	if removed != 1 {
		t.Fatalf("expected 1 removed by TTL, got %d", removed)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("old blob should be gone")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("fresh blob should survive")
	}
}

func TestPruneDir_SizeCapEvictsLowestScore(t *testing.T) {
	dir := t.TempDir()
	usage := filepath.Join(t.TempDir(), "u.qdf")
	now := time.Now()

	// Three 1000-byte blobs, all fresh (no TTL drop). Cap 2500 -> 80% target is
	// 2000, so dropping the single lowest-score entry (1000 bytes, total
	// 3000 -> 2000) is enough to reach the target.
	writeBlob(t, dir, "hot", 1000, now)
	writeBlob(t, dir, "warm", 1000, now)
	cold := writeBlob(t, dir, "cold", 1000, now)

	// Usage: hot most-hit, warm some, cold never -> cold has score 0.
	idx := cache.UsageIndex{}
	idx["hot"] = cache.UsageStat{Hits: 100, LastUsed: now.Unix()}
	idx["warm"] = cache.UsageStat{Hits: 10, LastUsed: now.Unix()}
	if err := cache.SaveUsage(usage, idx); err != nil {
		t.Fatal(err)
	}

	removed, _ := cache.PruneDir(dir, usage, 2500, 720*time.Hour, now.Unix(), false)
	if removed != 1 {
		t.Fatalf("expected 1 evicted to reach 80%% of cap, got %d", removed)
	}
	if _, err := os.Stat(cold); !os.IsNotExist(err) {
		t.Error("cold (score 0) blob should be the one evicted")
	}
}

func TestCacheScore_HotBeatsCold(t *testing.T) {
	now := int64(1_000_000)
	hot := cache.CacheScore(cache.UsageStat{Hits: 50, LastUsed: now}, now)
	coldOld := cache.CacheScore(cache.UsageStat{Hits: 50, LastUsed: now - 3600*100}, now)
	never := cache.CacheScore(cache.UsageStat{Hits: 0, LastUsed: now}, now)
	if !(hot > coldOld && coldOld > never) {
		t.Errorf("expected hot(%.3f) > coldOld(%.3f) > never(%.3f)", hot, coldOld, never)
	}
}

// TestPruneDir_DryRunReportsToTargetNotAll guards the dry-run size-cap count:
// a dry run must report evicting only down to 80% of the cap, not every blob.
// Regression for a bug where drop() didn't decrement the running total in
// dry-run, so the size-cap loop never hit its break and counted all blobs.
func TestPruneDir_DryRunReportsToTargetNotAll(t *testing.T) {
	dir := t.TempDir()
	usage := filepath.Join(t.TempDir(), "u.qdf")
	now := time.Now()
	// Three 1000-byte fresh blobs; cap 2500 -> 80% target 2000 -> exactly one
	// eviction reaches the target.
	writeBlob(t, dir, "a", 1000, now)
	writeBlob(t, dir, "b", 1000, now)
	writeBlob(t, dir, "c", 1000, now)

	removed, _ := cache.PruneDir(dir, usage, 2500, 720*time.Hour, now.Unix(), true /*dryRun*/)
	if removed != 1 {
		t.Fatalf("dry-run should report 1 eviction to reach 80%% target, got %d", removed)
	}
	// Dry run must not delete anything.
	if ents, _ := os.ReadDir(dir); len(ents) != 3 {
		t.Fatalf("dry-run must not delete blobs, %d remain", len(ents))
	}
}
