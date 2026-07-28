//go:build !race

// The race detector disables sync.Pool (Get always allocates fresh, Put is a
// no-op) so it can surface use-after-Put races. That makes pool-reuse identity
// unobservable, so this spot-check is excluded from -race builds.

package hook

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// TestBuildGrepSummary_GroupsMapPooled is the addendum spot-check: the
// file→matches map must be taken from grepGroupsPool, cleared, and returned —
// so a repeated summary reuses the same map instead of allocating a fresh one.
// A bare AllocsPerRun can't isolate that single map alloc among the ~45 the
// call makes, so this asserts pool reuse + clearing by pointer identity.
func TestBuildGrepSummary_GroupsMapPooled(t *testing.T) {
	var b strings.Builder
	for i := range 20 {
		fmt.Fprintf(&b, "internal/pkg/f%d.go:%d:some match text\n", i%4, i)
	}
	content := b.String()

	// Seed the pool with a sentinel so we can detect its reuse.
	sentinel := make(map[string][]grepMatch)
	sentinelPtr := reflect.ValueOf(sentinel).Pointer()
	grepGroupsPool.Put(sentinel)

	if s, action := buildGrepSummary(content); action != "grouped" || s == "" {
		t.Fatalf("expected a grouped summary, got action=%q summary=%q", action, s)
	}

	// The call must have taken the sentinel, used it, cleared it, and returned
	// it — so the next Get yields the same, now-empty map.
	got := grepGroupsPool.Get().(map[string][]grepMatch)
	if reflect.ValueOf(got).Pointer() != sentinelPtr {
		t.Error("buildGrepSummary did not reuse the pooled groups map (map alloc not eliminated)")
	}
	if len(got) != 0 {
		t.Errorf("pooled groups map must be cleared before reuse, got %d entries", len(got))
	}
}
