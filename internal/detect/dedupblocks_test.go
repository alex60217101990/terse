package detect_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

// Two identical multi-line blocks separated by other content (non-adjacent):
// run-length collapse can't touch them, but block folding must.
func TestFoldRepeatedBlocks_FoldsNonAdjacentDuplicates(t *testing.T) {
	block := "# widget-client-interface-methods\n" +
		strings.Repeat("func (c *client) PostAPIWidgetV1Items(...) (*Resp, error)\n", 6)
	other := "## some other section\nunrelated line one\nunrelated line two\n"
	in := block + "\n\n" + other + "\n\n" + block + "\n"

	out := detect.FoldRepeatedBlocks(in)
	if len(out) >= len(in) {
		t.Fatalf("expected shrink, got %d >= %d", len(out), len(in))
	}
	// First occurrence stays verbatim; the whole block text must not appear twice.
	if strings.Count(out, "PostAPIWidgetV1Items(...)") == strings.Count(in, "PostAPIWidgetV1Items(...)") {
		t.Fatalf("second occurrence was not folded:\n%s", out)
	}
	// First occurrence preserved.
	if !strings.Contains(out, block) {
		t.Fatalf("first occurrence not preserved verbatim:\n%s", out)
	}
	// A back-reference marker must identify the folded block by its first line.
	if !strings.Contains(out, "widget-client-interface-methods") {
		t.Fatalf("marker should name the repeated block:\n%s", out)
	}
	// The unrelated content between the duplicates is untouched.
	if !strings.Contains(out, other) {
		t.Fatalf("interposed content altered:\n%s", out)
	}
}

// No duplicates -> byte-identical passthrough (never-worse, never-altered).
func TestFoldRepeatedBlocks_NoDuplicatesUnchanged(t *testing.T) {
	in := "block a line1\nblock a line2\n\nblock b line1\nblock b line2\n\nblock c\n"
	if out := detect.FoldRepeatedBlocks(in); out != in {
		t.Fatalf("unique content must be unchanged\n in: %q\nout: %q", in, out)
	}
}

// Tiny duplicate blocks must NOT fold — a marker would cost more than it saves.
func TestFoldRepeatedBlocks_SkipsTinyBlocks(t *testing.T) {
	in := "}\n\nx := 1\n\n}\n\nx := 1\n"
	if out := detect.FoldRepeatedBlocks(in); out != in {
		t.Fatalf("tiny blocks must be left alone\nout: %q", out)
	}
}

// Byte-exactness: separators of varying blank-line width around a kept block
// are reproduced exactly when nothing folds.
func TestFoldRepeatedBlocks_PreservesSeparatorsExactly(t *testing.T) {
	in := "alpha\nbeta\n\n\ngamma\ndelta\n\nlast\n"
	if out := detect.FoldRepeatedBlocks(in); out != in {
		t.Fatalf("separators not preserved byte-exactly\n in: %q\nout: %q", in, out)
	}
}

// The common case: multi-block output with NO duplicates. The function must
// do minimal work and, ideally, allocate nothing but the dedup map.
func BenchmarkFoldRepeatedBlocks_NoDup(b *testing.B) {
	var sb strings.Builder
	for i := range 12 {
		sb.WriteString("## section ")
		sb.WriteByte(byte('a' + i))
		sb.WriteString("\nunique line one for this section\nunique line two here\n\n")
	}
	in := sb.String()
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.FoldRepeatedBlocks(in)
	}
}

func BenchmarkFoldRepeatedBlocks(b *testing.B) {
	block := "# section-header\n" + strings.Repeat("some representative line of output here\n", 8)
	other := "## other\n" + strings.Repeat("noise line\n", 5)
	in := block + "\n\n" + other + "\n\n" + block + "\n\n" + other + "\n\n" + block + "\n"
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.FoldRepeatedBlocks(in)
	}
}

// Two near-duplicate blocks that differ only in volatile tokens: exercises the
// normalize + diff + fuzzy-marker path.
func BenchmarkFoldRepeatedBlocks_Fuzzy(b *testing.B) {
	mk := func(id int, ts string) string {
		return "### request-log\n" +
			"handler widget-service latency=42ms\n" +
			"request id=" + fmt.Sprint(id) + " completed at " + ts + "\n" +
			"payload bytes=2048 checksum=deadbeefcafe1234\n" +
			"status=ok worker=w1 retries=0\n"
	}
	in := mk(7, "2026-07-27T10:00:01Z") + "\n\n" + mk(9, "2026-07-27T10:00:02Z") + "\n"
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.FoldRepeatedBlocks(in)
	}
}

// Blocks identical except volatile tokens (ids, hex, timestamps) fold into a
// marker that SHOWS the differing tokens — zero information loss.
func TestFoldRepeatedBlocks_FuzzyFoldsWithDiff(t *testing.T) {
	mk := func(id int, ts string) string {
		return "### request-log\n" +
			"handler widget-service latency=42ms\n" +
			"request id=" + fmt.Sprint(id) + " completed at " + ts + "\n" +
			"payload bytes=2048 checksum=deadbeefcafe1234\n" +
			"status=ok worker=w1 retries=0\n"
	}
	a := mk(7, "2026-07-27T10:00:01Z")
	b := mk(9, "2026-07-27T10:00:02Z")
	other := "### separator\ncompletely different content here\n"
	in := a + "\n\n" + other + "\n\n" + b + "\n"

	out := detect.FoldRepeatedBlocks(in)
	if len(out) >= len(in) {
		t.Fatalf("expected shrink, got %d vs %d", len(out), len(in))
	}
	// the differing tokens must BOTH be visible in the fold marker
	for _, must := range []string{"7->9", "10:00:01", "10:00:02"} {
		if !strings.Contains(out, must) {
			t.Errorf("marker must show difference %q:\n%s", must, out)
		}
	}
	if !strings.Contains(out, a) {
		t.Errorf("first occurrence must stay verbatim")
	}
}

// If the blocks differ in too many places, folding must NOT happen.
func TestFoldRepeatedBlocks_TooDifferentStaysVerbatim(t *testing.T) {
	a := "### sec\naaa bbb ccc ddd\neee fff ggg hhh\niii jjj kkk lll\nmmm nnn ooo ppp\n"
	b := "### sec\n111 222 333 444\n555 666 777 888\n999 000 111 222\n333 444 555 666\n"
	in := a + "\n\n" + b + "\n"
	if out := detect.FoldRepeatedBlocks(in); out != in {
		t.Fatalf("wholly different blocks must not fuzzy-fold:\n%s", out)
	}
}

// Blocks with the SAME normalized shape but more than the allowed number of
// differing volatile tokens must NOT fold — otherwise the marker would either
// omit a difference or be no smaller than the block.
func TestFoldRepeatedBlocks_ManyDiffsStayVerbatim(t *testing.T) {
	mk := func(a, b, c, d, e int) string {
		return fmt.Sprintf("### metrics-row\n"+
			"cpu=%d mem=%d disk=%d net=%d fd=%d handles fixed text tail here\n"+
			"second unchanging line of content to clear the size floor here\n", a, b, c, d, e)
	}
	x := mk(1, 2, 3, 4, 5)
	y := mk(6, 7, 8, 9, 10) // five differing tokens > maxFuzzyDiffs(4)
	in := x + "\n\n" + y + "\n"
	if out := detect.FoldRepeatedBlocks(in); out != in {
		t.Fatalf("blocks differing in >4 tokens must stay verbatim:\n%s", out)
	}
}

// An exact copy of a near-duplicate folds against the verbatim base (single
// hop), and every differing token is still shown — no information is lost and
// no reference dangles at a folded (non-verbatim) block.
func TestFoldRepeatedBlocks_ExactCopyOfNearDupFoldsToBase(t *testing.T) {
	mk := func(id int) string {
		return "### job-record\n" +
			"worker started job id=" + fmt.Sprint(id) + " on host node-alpha here\n" +
			"stage=complete elapsed unchanging descriptive tail to pass floor\n"
	}
	a := mk(1)
	b := mk(2)
	in := a + "\n\n" + b + "\n\n" + b + "\n" // base, near-dup, exact copy of near-dup
	out := detect.FoldRepeatedBlocks(in)
	if len(out) >= len(in) {
		t.Fatalf("expected shrink, got %d vs %d", len(out), len(in))
	}
	if !strings.Contains(out, a) {
		t.Fatalf("verbatim base must be preserved:\n%s", out)
	}
	// The base (id=1) must appear verbatim exactly once; both later occurrences
	// fold. Neither "id=2" copy may survive as raw block text.
	if got := strings.Count(out, "worker started job id=2 on host"); got != 0 {
		t.Fatalf("near-dup and its exact copy must both fold, found %d raw:\n%s", got, out)
	}
	// Both folded occurrences must show the 1->2 difference.
	if got := strings.Count(out, "1->2"); got != 2 {
		t.Fatalf("expected both folds to show 1->2, got %d:\n%s", got, out)
	}
}

// Zero-loss guard: if two differing tokens in the SAME block both have old
// value "111" but diverge to different new values ("222" vs "333"), a marker
// listing "111->222, 111->333" could not tell a reader which occurrence of
// "111" becomes which — ambiguous reconstruction. The fold must be refused,
// leaving BOTH blocks verbatim (output byte-identical to input).
func TestFoldRepeatedBlocks_AmbiguousDuplicateOldStaysVerbatim(t *testing.T) {
	a := "### dup-old-values\n" +
		"left=111 right=111 padding text to clear the block size floor nicely here\n" +
		"second line of unchanging content that keeps this block comfortably sized\n"
	b := "### dup-old-values\n" +
		"left=222 right=333 padding text to clear the block size floor nicely here\n" +
		"second line of unchanging content that keeps this block comfortably sized\n"
	in := a + "\n\n" + b + "\n"

	out := detect.FoldRepeatedBlocks(in)
	if out != in {
		t.Fatalf("ambiguous duplicate-old fold must be refused (verbatim):\n in: %q\nout: %q", in, out)
	}
}

// Duplicate old values that do NOT diverge (both instances stay "111" in both
// blocks) are not diff pairs at all — only the genuinely differing token
// ("extra") becomes a pair, so no ambiguity exists and the fold must proceed
// normally.
func TestFoldRepeatedBlocks_DuplicateUnchangedTokensStillFold(t *testing.T) {
	a := "### dup-nochange\n" +
		"left=111 right=111 extra=1 padding text so this line clears the floor nicely\n" +
		"second unchanging line of content that keeps the block comfortably sized well\n"
	b := "### dup-nochange\n" +
		"left=111 right=111 extra=2 padding text so this line clears the floor nicely\n" +
		"second unchanging line of content that keeps the block comfortably sized well\n"
	in := a + "\n\n" + b + "\n"

	out := detect.FoldRepeatedBlocks(in)
	if len(out) >= len(in) {
		t.Fatalf("expected shrink (non-ambiguous fold), got %d vs %d:\n%s", len(out), len(in), out)
	}
	if !strings.Contains(out, "1->2") {
		t.Fatalf("marker must show the 1->2 difference:\n%s", out)
	}
	if !strings.Contains(out, a) {
		t.Fatalf("first occurrence must stay verbatim:\n%s", out)
	}
	if strings.Count(out, "extra=2") != 0 {
		t.Fatalf("second occurrence must fold, not stay raw:\n%s", out)
	}
}

// Exact-dup behavior is unchanged (regression).
func TestFoldRepeatedBlocks_ExactStillWins(t *testing.T) {
	block := "# hdr\n" + strings.Repeat("same line of content here\n", 6)
	in := block + "\n\n" + block + "\n"
	out := detect.FoldRepeatedBlocks(in)
	if !strings.Contains(out, "[repeat: ") {
		t.Fatalf("exact dup must use the exact marker:\n%s", out)
	}
}

// The dedup maps are recycled via sync.Pool. Run distinct payloads through
// FoldRepeatedBlocks concurrently and require each to equal its own serial
// result — a pooled map that leaked entries across calls, or was shared
// between goroutines, would corrupt a fold and diverge (and trip -race).
func TestFoldRepeatedBlocks_ConcurrentPoolIsolation(t *testing.T) {
	inputs := make([]string, 8)
	want := make([]string, 8)
	for i := range inputs {
		// Each payload folds differently (distinct headers, exact + fuzzy
		// blocks with per-i volatile tokens) so a cross-call key leak shows up.
		block := fmt.Sprintf("# section-%d-header-line-that-is-long-enough\n", i) +
			strings.Repeat(fmt.Sprintf("payload line for section %d content here\n", i), 6)
		fuzzy := fmt.Sprintf("id=%d aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa req=%d\n", i, i)
		fuzzy2 := fmt.Sprintf("id=%d aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa req=%d\n", i+100, i+100)
		in := block + "\n\n" + fuzzy + "\n\n" + block + "\n\n" + fuzzy2 + "\n"
		inputs[i] = in
		want[i] = detect.FoldRepeatedBlocks(in)
	}

	var wg sync.WaitGroup
	for range 200 {
		for i := range inputs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if got := detect.FoldRepeatedBlocks(inputs[i]); got != want[i] {
					t.Errorf("payload %d diverged under concurrency:\nwant %q\ngot  %q", i, want[i], got)
				}
			}(i)
		}
	}
	wg.Wait()
}
