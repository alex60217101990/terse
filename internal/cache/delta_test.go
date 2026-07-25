package cache_test

import (
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
	d := string(cache.UnifiedDiff([]byte(old), []byte(nw), 3))
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
