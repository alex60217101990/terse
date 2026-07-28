package hook

import (
	"fmt"
	"strings"
	"testing"
)

// TestShouldPoolGrepGroups_Boundary pins the size-guard decision deterministically
// (no sync.Pool involvement): a map at exactly the cap is still pooled, one past
// is dropped — clear() empties entries but does not shrink the bucket array, so
// an oversized map must not be retained.
func TestShouldPoolGrepGroups_Boundary(t *testing.T) {
	if !shouldPoolGrepGroups(0) {
		t.Error("an empty map must be poolable")
	}
	if !shouldPoolGrepGroups(maxPooledGrepGroups) {
		t.Errorf("a map at exactly the cap (%d) must still be pooled", maxPooledGrepGroups)
	}
	if shouldPoolGrepGroups(maxPooledGrepGroups + 1) {
		t.Errorf("a map one entry past the cap (%d) must be dropped, not pooled", maxPooledGrepGroups+1)
	}
}

// TestBuildGrepSummary_GroupsMapPooled checks the pool contract behaviorally,
// without asserting sync.Pool identity (neighboring tests' pool traffic steals
// the per-P slot, so identity is not observably stable — the earlier
// identity-based version was flaky in the full-package run). buildGrepSummary
// must clear its groups map before returning it, so every map the pool hands
// back is empty, whichever object arrives. Runs a tight Get→(build
// populates+Puts)→Get loop and asserts both len==0 and correct summary output
// each iteration. Deterministic under -race too (there Get always yields a
// fresh empty map).
func TestBuildGrepSummary_GroupsMapPooled(t *testing.T) {
	var b strings.Builder
	for i := range 20 {
		fmt.Fprintf(&b, "internal/pkg/f%d.go:%d:some match text\n", i%4, i)
	}
	content := b.String()

	for range 8 {
		// Any map currently in the pool must be empty: a prior buildGrepSummary
		// clears before Put, so no earlier call's entries can survive.
		m := grepGroupsPool.Get().(map[string][]grepMatch)
		if len(m) != 0 {
			t.Fatalf("pooled groups map must be empty (cleared before Put), got %d entries", len(m))
		}
		grepGroupsPool.Put(m)

		s, action := buildGrepSummary(content)
		if action != "grouped" || s == "" {
			t.Fatalf("expected a grouped summary, got action=%q summary=%q", action, s)
		}
		if !strings.Contains(s, "matches in") {
			t.Errorf("summary missing expected footer: %q", s)
		}
	}
}
