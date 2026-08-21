package cache_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
)

func TestUnifiedDiff_NoChange(t *testing.T) {
	content := []byte("line1\nline2\nline3\n")
	got := cache.UnifiedDiff(content, content, 3)
	if got != "" {
		t.Errorf("identical content should produce empty diff, got:\n%s", got)
	}
}

func TestUnifiedDiff_SingleLineChange(t *testing.T) {
	old := []byte("line1\nline2\nline3\n")
	newer := []byte("line1\nLINE2\nline3\n")
	got := cache.UnifiedDiff(old, newer, 1)
	if !strings.Contains(got, "-line2") {
		t.Errorf("diff should contain -line2:\n%s", got)
	}
	if !strings.Contains(got, "+LINE2") {
		t.Errorf("diff should contain +LINE2:\n%s", got)
	}
	if !strings.Contains(got, "@@") {
		t.Errorf("diff should contain @@ hunk header:\n%s", got)
	}
}

func TestUnifiedDiff_AddedLines(t *testing.T) {
	old := []byte("a\nb\n")
	newer := []byte("a\nb\nc\nd\n")
	got := cache.UnifiedDiff(old, newer, 0)
	if !strings.Contains(got, "+c") || !strings.Contains(got, "+d") {
		t.Errorf("diff should show added lines:\n%s", got)
	}
	if !strings.Contains(got, "@@") {
		t.Errorf("diff should contain @@ hunk header:\n%s", got)
	}
}

func TestUnifiedDiff_DeletedLines(t *testing.T) {
	old := []byte("a\nb\nc\n")
	newer := []byte("a\nc\n")
	got := cache.UnifiedDiff(old, newer, 0)
	if !strings.Contains(got, "-b") {
		t.Errorf("diff should show -b:\n%s", got)
	}
	if !strings.Contains(got, "@@") {
		t.Errorf("diff should contain @@ hunk header:\n%s", got)
	}
}

func TestUnifiedDiff_PureInsertion(t *testing.T) {
	old := []byte("a\nb\n")
	newer := []byte("a\nb\nc\n")
	got := cache.UnifiedDiff(old, newer, 0)
	if !strings.Contains(got, "+c") {
		t.Errorf("diff should show +c:\n%s", got)
	}
}

func TestUnifiedDiff_PureInsertion_HunkHeader(t *testing.T) {
	old := []byte("a\nb\n")
	newer := []byte("a\nb\nc\nd\n")
	got := cache.UnifiedDiff(old, newer, 0)
	// Should contain @@ -2,0 +3,2 @@ (inserting after line 2, zero old lines, 2 new lines)
	if !strings.Contains(got, "@@ -2,0 +3,2 @@") {
		t.Errorf("pure insertion hunk header wrong:\n%s", got)
	}
}

func TestUnifiedDiff_PureDeletion(t *testing.T) {
	old := []byte("a\nb\nc\n")
	newer := []byte("a\nc\n")
	got := cache.UnifiedDiff(old, newer, 0)
	if !strings.Contains(got, "-b") {
		t.Errorf("diff should show -b:\n%s", got)
	}
}

func TestUnifiedDiff_EmptyOld(t *testing.T) {
	old := []byte("")
	newer := []byte("a\nb\n")
	got := cache.UnifiedDiff(old, newer, 0)
	if !strings.Contains(got, "+a") || !strings.Contains(got, "+b") {
		t.Errorf("diff from empty should show all lines added:\n%s", got)
	}
}

func TestUnifiedDiff_EmptyNew(t *testing.T) {
	old := []byte("a\nb\n")
	newer := []byte("")
	got := cache.UnifiedDiff(old, newer, 0)
	if !strings.Contains(got, "-a") || !strings.Contains(got, "-b") {
		t.Errorf("diff to empty should show all lines deleted:\n%s", got)
	}
}

func TestUnifiedDiff_LargeFile(t *testing.T) {
	v1, err := os.ReadFile("testdata/encoder_go_v1.txt")
	if err != nil {
		t.Fatalf("read v1: %v", err)
	}
	v2, err := os.ReadFile("testdata/encoder_go_v2.txt")
	if err != nil {
		t.Fatalf("read v2: %v", err)
	}
	got := cache.UnifiedDiff(v1, v2, 3)
	if got == "" {
		t.Fatal("expected non-empty diff between v1 and v2")
	}
	// v2 has changed lines; diff must contain deletions and insertions.
	if !strings.Contains(got, "-") {
		t.Errorf("diff should contain at least one deletion:\n%s", got)
	}
	if !strings.Contains(got, "+") {
		t.Errorf("diff should contain at least one insertion:\n%s", got)
	}
}

func TestIsBinaryContent(t *testing.T) {
	if cache.IsBinaryContent([]byte("hello world\n")) {
		t.Error("plain text should not be binary")
	}
	if !cache.IsBinaryContent([]byte("hello\x00world")) {
		t.Error("null byte should be binary")
	}
	if !cache.IsBinaryContent(append([]byte{0xFF, 0xFE}, []byte("bad utf8")...)) {
		t.Error("invalid UTF-8 should be binary")
	}
}

func BenchmarkUnifiedDiff(b *testing.B) {
	// 500-line file, 10 lines changed
	var old, newer strings.Builder
	for i := range 500 {
		old.WriteString("func foo() { // line\n")
		if i >= 100 && i < 110 {
			newer.WriteString("func bar() { // changed\n")
		} else {
			newer.WriteString("func foo() { // line\n")
		}
	}
	oldB := []byte(old.String())
	newB := []byte(newer.String())

	for b.Loop() {
		_ = cache.UnifiedDiff(oldB, newB, 3)
	}
}

// TestUnifiedDiff_NoOverlappingHunks is a correctness regression: two changes
// separated by an equal-run G with ctx < G <= 2*ctx must merge into ONE hunk.
// The old merge threshold (G <= ctx) left them as two hunks whose ranges
// overlapped — duplicating context lines and emitting @@ headers that overran
// the file (e.g. "@@ -5,4" on a 6-line file). Trigger is extremely common:
// any two edits 4-5 unchanged lines apart at the default ctx=3.
func TestUnifiedDiff_NoOverlappingHunks(t *testing.T) {
	old := "a\nx\nx\nx\nx\nb\n"
	nw := "A\nx\nx\nx\nx\nB\n"
	d := cache.UnifiedDiff([]byte(old), []byte(nw), 3)
	if h := strings.Count(d, "@@ -"); h != 1 {
		t.Fatalf("want 1 merged hunk, got %d\n%s", h, d)
	}
	if x := strings.Count(d, "\n x"); x != 4 {
		t.Fatalf("want 4 context 'x' lines (no duplication), got %d\n%s", x, d)
	}
	if !strings.Contains(d, "@@ -1,6 +1,6 @@") {
		t.Fatalf("want header spanning the whole 6-line file, got:\n%s", d)
	}
}

// TestUnifiedDiff_BailsOnPathologicalDissimilar: two large, fully-dissimilar
// inputs would drive full-trace Myers to O((N+M)^2) transient memory. The
// edit-distance cap must bail (empty diff → caller serves full content) rather
// than allocate gigabytes, while a large but SIMILAR input still diffs.
func TestUnifiedDiff_BailsOnPathologicalDissimilar(t *testing.T) {
	var a, b strings.Builder
	for i := range 4000 {
		fmt.Fprintf(&a, "old-line-%d\n", i)
		fmt.Fprintf(&b, "new-different-%d\n", i) // shares no line with a
	}
	if d := cache.UnifiedDiff([]byte(a.String()), []byte(b.String()), 3); d != "" {
		t.Fatalf("fully-dissimilar large input must bail to empty diff, got %d bytes", len(d))
	}

	// Large but similar (one changed line) still produces a diff.
	var c, e strings.Builder
	for i := range 4000 {
		fmt.Fprintf(&c, "line-%d\n", i)
		if i == 2000 {
			e.WriteString("line-CHANGED\n")
		} else {
			fmt.Fprintf(&e, "line-%d\n", i)
		}
	}
	if d := cache.UnifiedDiff([]byte(c.String()), []byte(e.String()), 3); d == "" {
		t.Fatal("large similar input (1 change) must still diff, got empty")
	}
}

// TestUnifiedDiff_DeterministicRepeated pins byte-identical output across
// repeated calls on the same large input. This guards against pooled-buffer
// reuse bugs (e.g. a stale trace snapshot or scratch slice leaking data from
// a previous call into the next one).
func TestUnifiedDiff_DeterministicRepeated(t *testing.T) {
	v1, err := os.ReadFile("testdata/encoder_go_v1.txt")
	if err != nil {
		t.Fatalf("read v1: %v", err)
	}
	v2, err := os.ReadFile("testdata/encoder_go_v2.txt")
	if err != nil {
		t.Fatalf("read v2: %v", err)
	}

	want := cache.UnifiedDiff(v1, v2, 3)
	if want == "" {
		t.Fatal("expected non-empty reference diff")
	}
	for i := range 50 {
		got := cache.UnifiedDiff(v1, v2, 3)
		if got != want {
			t.Fatalf("iteration %d: diff output diverged from reference\nwant:\n%s\ngot:\n%s", i, want, got)
		}
	}
}

// TestUnifiedDiff_ParallelDistinctInputs guards against cross-goroutine
// scratch sharing: 8 goroutines diff distinct inputs concurrently, each
// asserting its result equals its own serial (single-goroutine) reference.
// If pooled scratch buffers ever leaked between goroutines, this would
// surface as one goroutine's diff contaminated by another's data.
func TestUnifiedDiff_ParallelDistinctInputs(t *testing.T) {
	t.Parallel()

	const goroutines = 8

	type input struct{ old, newer []byte }
	inputs := make([]input, goroutines)
	wants := make([]string, goroutines)

	for g := range goroutines {
		var old, newer strings.Builder
		for i := range 300 {
			fmt.Fprintf(&old, "g%d-line-%d\n", g, i)
			if i%17 == g%17 {
				fmt.Fprintf(&newer, "g%d-CHANGED-%d\n", g, i)
			} else {
				fmt.Fprintf(&newer, "g%d-line-%d\n", g, i)
			}
		}
		inputs[g] = input{old: []byte(old.String()), newer: []byte(newer.String())}
		wants[g] = cache.UnifiedDiff(inputs[g].old, inputs[g].newer, 3)
		if wants[g] == "" {
			t.Fatalf("goroutine %d: expected non-empty reference diff", g)
		}
	}

	for g := range goroutines {
		t.Run(fmt.Sprintf("g%d", g), func(t *testing.T) {
			t.Parallel()
			for i := range 20 {
				got := cache.UnifiedDiff(inputs[g].old, inputs[g].newer, 3)
				if got != wants[g] {
					t.Fatalf("goroutine %d iteration %d: diff diverged from serial reference\nwant:\n%s\ngot:\n%s", g, i, wants[g], got)
				}
			}
		})
	}
}
