package hook

import (
	"strings"
	"testing"
)

// TestTryGrep_ZeroAllocsOnProse guards the perf fix: tryGrep must bail out via
// looksLikeGrep before ever calling buildGrepSummary for non-grep-shaped
// content. Previously every generic Bash payload paid buildGrepSummary's tree
// fallback (buildGlobTree -> topGlobDir -> strings.SplitN per line), which
// regressed BenchmarkDaemonRoundtrip from 64 to 121 allocs/op.
func TestTryGrep_ZeroAllocsOnProse(t *testing.T) {
	content := strings.Repeat("this is an ordinary line of prose output, not grep matches\n", 40)

	allocs := testing.AllocsPerRun(100, func() {
		_ = tryGrep(content)
	})
	if allocs != 0 {
		t.Errorf("tryGrep on plain prose: got %v allocs/op, want 0", allocs)
	}
}

// TestLooksLikeGrep verifies the shape probe accepts grep-shaped content and
// rejects ordinary text, matching buildGrepSummary's own parsed/lineCount
// classification (parsed == 0 || parsed < lineCount/2 => not grep).
func TestLooksLikeGrep(t *testing.T) {
	grepShaped := "internal/hook/bash.go:10:package hook\ninternal/hook/grep.go:5:import \"fmt\"\n"
	if !looksLikeGrep(grepShaped) {
		t.Error("expected grep-shaped content to pass the probe")
	}

	prose := "this is not grep output at all, just some prose\nwith a few lines\n"
	if looksLikeGrep(prose) {
		t.Error("expected plain prose to fail the probe")
	}

	if looksLikeGrep("") {
		t.Error("expected empty content to fail the probe")
	}
}

// worth gates a summary by ratio: the loose gate keeps a 25–49% win that the
// strict gate would discard (the old single 0.5 gate silently dropped those).
func TestWorth_Gates(t *testing.T) {
	content := strings.Repeat("x", 1000)
	sixtyPct := strings.Repeat("x", 600) // 60% of original -> 40% saved

	if worth(sixtyPct, content, minSummaryRatio) {
		t.Error("strict 0.5 gate must reject a 60% summary")
	}
	if !worth(sixtyPct, content, minSummaryRatioLoose) {
		t.Error("loose 0.75 gate must accept a 60% summary")
	}
	if worth("", content, minSummaryRatioLoose) {
		t.Error("empty summary must never be worth it")
	}
}
