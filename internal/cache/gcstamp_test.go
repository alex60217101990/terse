package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/cache"
)

func TestShouldRunGC_Throttle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	now := time.Now().Unix()
	if !cache.ShouldRunGC(now) {
		t.Fatal("first call (no stamp) should run")
	}
	if cache.ShouldRunGC(now + 3600) { // 1h later
		t.Error("within 24h should NOT run again")
	}
	if !cache.ShouldRunGC(now + 25*3600) { // 25h later
		t.Error("after 24h should run again")
	}
}

// TestAutoSweep_ThrottledPrune verifies AutoSweep prunes over-cap blobs on
// first call and is throttled (no-op) within the gc interval — this is the
// disk-bound trigger wired into the SessionStart `daemon --ensure` hook.
func TestAutoSweep_ThrottledPrune(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("QDF_CACHE_MAX_SIZE", "2000")
	refs := cache.RefsDir()
	if err := os.MkdirAll(refs, 0o755); err != nil {
		t.Fatal(err)
	}
	mk := func() {
		for _, n := range []string{"a", "b", "c"} {
			_ = os.WriteFile(filepath.Join(refs, n+".blob"), make([]byte, 1000), 0o600)
		}
	}
	mk()
	now := int64(2_000_000_000)
	cache.AutoSweep(now) // first call: stamp absent -> runs, prunes to <=80% of 1000 (half cap)
	if ents, _ := os.ReadDir(refs); len(ents) >= 3 {
		t.Fatalf("AutoSweep first call should prune over-cap blobs, %d remain", len(ents))
	}
	// Refill and call again within the interval — throttle must skip it.
	mk()
	cache.AutoSweep(now + 3600) // +1h < 24h
	if ents, _ := os.ReadDir(refs); len(ents) != 3 {
		t.Errorf("AutoSweep within 24h must be a no-op, got %d blobs", len(ents))
	}
}
