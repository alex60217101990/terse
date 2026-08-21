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

// TestParseGrepLine_RejectsImplausiblePaths is the data-corruption guard: a
// colon-delimited line only classifies as a grep "file:line:text" content
// line when its file segment is a believable path. Timestamped logs and
// colon-delimited config dumps must be rejected outright, while real
// ripgrep/grep output must still parse.
func TestParseGrepLine_RejectsImplausiblePaths(t *testing.T) {
	reject := []string{
		"2024-01-01T00:00:00 some log message", // ISO timestamp, no path chars
		"12:30:45 message",                     // clock time, no path chars
		"service:8000:description text",        // config dump, no path chars
		"2024-01-15.log:5:something",           // date-shaped file segment (4 digits + '-')
		"some file.go:1:x",                     // space in file segment
	}
	// db.host:5432:desc (no spaces) is acceptable residual risk and is NOT in
	// the reject set — it legitimately looks like a path (has a '.').
	for _, s := range reject {
		if _, _, _, ok := parseGrepLine(s); ok {
			t.Errorf("parseGrepLine(%q) classified as grep; want rejected", s)
		}
	}

	accept := []string{
		"internal/pkg/file.go:123:\tcode line",
		"path/with-dash/x_test.go:9:text",
		"README.md:5:# Title",
		"a/b.c:1:x",
	}
	for _, s := range accept {
		if _, _, _, ok := parseGrepLine(s); !ok {
			t.Errorf("parseGrepLine(%q) rejected; want classified as grep", s)
		}
	}
}

// TestLooksLikeGrep_TimestampCorpus is the end-to-end shape probe check: whole
// blocks of timestamped logs / config dumps must not be mistaken for grep
// output, while a block of real grep hits must be.
func TestLooksLikeGrep_TimestampCorpus(t *testing.T) {
	logs := strings.Repeat("2024-01-01T00:00:00 some log message\n", 10)
	if looksLikeGrep(logs) {
		t.Error("timestamped logs must not classify as grep")
	}
	clock := strings.Repeat("12:30:45 message\n", 10)
	if looksLikeGrep(clock) {
		t.Error("clock-time logs must not classify as grep")
	}
	cfg := "service:8000:description text\nworker:9000:another description\n" +
		strings.Repeat("gateway:7000:more text here\n", 8)
	if looksLikeGrep(cfg) {
		t.Error("colon-delimited config dump must not classify as grep")
	}

	realGrepHits := "internal/pkg/file.go:123:\tcode line\n" +
		"path/with-dash/x_test.go:9:text\n" +
		"README.md:5:# Title\n" +
		"a/b.c:1:x\n" +
		strings.Repeat("internal/hook/bash.go:10:package hook\n", 6)
	if !looksLikeGrep(realGrepHits) {
		t.Error("real grep output must classify as grep")
	}
}
