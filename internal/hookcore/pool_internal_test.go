package hookcore

import (
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/cache"
)

// TestFlushBufPool_DropsOversizedBuffer drives FlushDirty with a session
// whose encode blows past cache.MaxPooledBufSize, then verifies the pool
// does NOT hand that oversized buffer back out on the next Get — it must
// return a fresh, small buffer instead (New's 4096 default), proving the
// cap-before-Put guard actually fires rather than merely compiling.
func TestFlushBufPool_DropsOversizedBuffer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := NewMemStore()
	s := m.StateStore()

	// One giant file well over cache.MaxPooledBufSize (4MB) so the qdf encode
	// grows flushBufPool's buffer past the cap.
	st := cache.NewSessionState()
	st.Files["/big"] = cache.FileEntry{Content: make([]byte, 6<<20)}
	s.SaveSession("big-sess", st)

	m.FlushDirty() // must encode >4MB and, per the fix, drop the buffer

	bufPtr := flushBufPool.Get().(*[]byte)
	defer flushBufPool.Put(bufPtr)
	if cap(*bufPtr) > cache.MaxPooledBufSize {
		t.Fatalf("pool returned an oversized buffer (cap=%d) after a >4MB flush; "+
			"the cap-before-Put guard should have dropped it", cap(*bufPtr))
	}
}

// TestPruneUsage_DropsOldEntries seeds usageRefs/usageLast with a mix of
// stale and fresh entries directly (bypassing Bump so LastUsed is exact),
// runs PruneUsage, and asserts: (1) only the stale entries are dropped from
// RAM, and (2) after a subsequent FlushDirty the on-disk sidecar reflects the
// same pruning — i.e. the RAM prune actually prevents resurrection rather
// than being undone by the next flush.
func TestPruneUsage_DropsOldEntries(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := NewMemStore()

	now := time.Now().Unix()
	maxAge := 720 * time.Hour
	ttlSec := int64(maxAge / time.Second)

	stale := now - ttlSec - 3600 // 1h past the TTL floor
	fresh := now - 60            // well within TTL

	m.usageMu.Lock()
	m.usageRefs["stale-ref"] = cache.UsageStat{Hits: 1, LastUsed: stale}
	m.usageRefs["fresh-ref"] = cache.UsageStat{Hits: 1, LastUsed: fresh}
	m.usageLast["stale-last"] = cache.UsageStat{Hits: 1, LastUsed: stale}
	m.usageLast["fresh-last"] = cache.UsageStat{Hits: 1, LastUsed: fresh}
	m.usageMu.Unlock()

	m.PruneUsage(maxAge)

	m.usageMu.Lock()
	_, staleRefStillThere := m.usageRefs["stale-ref"]
	_, freshRefStillThere := m.usageRefs["fresh-ref"]
	_, staleLastStillThere := m.usageLast["stale-last"]
	_, freshLastStillThere := m.usageLast["fresh-last"]
	m.usageMu.Unlock()

	if staleRefStillThere {
		t.Error("PruneUsage should have dropped the stale usageRefs entry")
	}
	if !freshRefStillThere {
		t.Error("PruneUsage should have kept the fresh usageRefs entry")
	}
	if staleLastStillThere {
		t.Error("PruneUsage should have dropped the stale usageLast entry")
	}
	if !freshLastStillThere {
		t.Error("PruneUsage should have kept the fresh usageLast entry")
	}

	// After a flush, the sidecar on disk must not resurrect the pruned entry.
	m.FlushDirty()
	refsIdx := cache.LoadUsage(cache.UsageRefsPath())
	if _, ok := refsIdx["stale-ref"]; ok {
		t.Error("flushed usage-refs sidecar resurrected a pruned entry")
	}
	if _, ok := refsIdx["fresh-ref"]; !ok {
		t.Error("flushed usage-refs sidecar lost the fresh entry it should have kept")
	}
	lastIdx := cache.LoadUsage(cache.UsageLastPath())
	if _, ok := lastIdx["stale-last"]; ok {
		t.Error("flushed usage-last sidecar resurrected a pruned entry")
	}
	if _, ok := lastIdx["fresh-last"]; !ok {
		t.Error("flushed usage-last sidecar lost the fresh entry it should have kept")
	}
}

// TestPruneUsage_EmptyIsSafe guards against a nil/empty-index panic.
func TestPruneUsage_EmptyIsSafe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := NewMemStore()
	m.PruneUsage(720 * time.Hour)
}
