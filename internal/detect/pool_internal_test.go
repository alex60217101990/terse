package detect

import "testing"

// TestShouldPoolDedupMap_Boundary pins the exact cap decision: a map at
// maxPooledDedupMapLen entries is still small enough to pool, one entry past
// it is not. This is the arithmetic the pool-cap fix hinges on; testing it
// directly (rather than via sync.Pool Get/Put timing, which is unspecified
// and flaky under -race/GOMAXPROCS>1) pins the behavior deterministically.
func TestShouldPoolDedupMap_Boundary(t *testing.T) {
	if !shouldPoolDedupMap(maxPooledDedupMapLen) {
		t.Errorf("a map at exactly the cap (%d) should still be pooled", maxPooledDedupMapLen)
	}
	if shouldPoolDedupMap(maxPooledDedupMapLen + 1) {
		t.Errorf("a map one entry past the cap (%d) should be dropped, not pooled", maxPooledDedupMapLen+1)
	}
	if !shouldPoolDedupMap(0) {
		t.Error("an empty map should be pooled")
	}
}

// TestPutSeen_ClearsRegardlessOfSize verifies putSeen always empties the map
// (dropped or pooled), so a dropped oversized map never leaks its old
// entries into whatever eventually reclaims it, and a pooled map never
// carries stale keys (whose backing content may be aliased into a caller's
// string) into the next FoldRepeatedBlocks call.
func TestPutSeen_ClearsRegardlessOfSize(t *testing.T) {
	small := map[string]struct{}{"a": {}}
	putSeen(small, len(small))
	if len(small) != 0 {
		t.Errorf("pooled (small) map should be cleared, len=%d", len(small))
	}

	big := make(map[string]struct{}, maxPooledDedupMapLen+8)
	for i := range maxPooledDedupMapLen + 1 {
		big[string(rune(i))] = struct{}{}
	}
	putSeen(big, len(big))
	if len(big) != 0 {
		t.Errorf("dropped (oversized) map should still be cleared, len=%d", len(big))
	}
}

// TestPutSeenNorm_ClearsRegardlessOfSize mirrors TestPutSeen_ClearsRegardlessOfSize
// for the seenNorm/seenNormPool pair.
func TestPutSeenNorm_ClearsRegardlessOfSize(t *testing.T) {
	small := map[string]string{"a": "b"}
	putSeenNorm(small, len(small))
	if len(small) != 0 {
		t.Errorf("pooled (small) map should be cleared, len=%d", len(small))
	}

	big := make(map[string]string, maxPooledDedupMapLen+8)
	for i := range maxPooledDedupMapLen + 1 {
		big[string(rune(i))] = "v"
	}
	putSeenNorm(big, len(big))
	if len(big) != 0 {
		t.Errorf("dropped (oversized) map should still be cleared, len=%d", len(big))
	}
}
