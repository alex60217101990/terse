package detect_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

func makeDiff(context int) string {
	var b strings.Builder
	b.WriteString("diff --git a/pkg/widget.go b/pkg/widget.go\n")
	b.WriteString("index 1a2b3c4..5d6e7f8 100644\n--- a/pkg/widget.go\n+++ b/pkg/widget.go\n")
	b.WriteString("@@ -10,40 +10,41 @@ func Widget() {\n")
	for i := range context {
		b.WriteString(" \tunchanged line ")
		b.WriteString(strings.Repeat("x", 20))
		if i%7 == 0 {
			b.WriteString(" tail")
		}
		b.WriteByte('\n')
	}
	b.WriteString("-\told := value\n+\tnew := value\n")
	for range context {
		b.WriteString(" \tmore unchanged context padding line\n")
	}
	return b.String()
}

func TestSummarizeGitDiff_CollapsesContext(t *testing.T) {
	in := makeDiff(20)
	out := detect.SummarizeGitDiff(in)
	if out == "" || len(out) >= len(in) {
		t.Fatalf("expected shrink, got %d vs %d", len(out), len(in))
	}
	for _, must := range []string{"diff --git a/pkg/widget.go", "@@ -10,40 +10,41 @@", "-\told := value", "+\tnew := value", "⋯"} {
		if !strings.Contains(out, must) {
			t.Errorf("missing %q in:\n%s", must, out)
		}
	}
}

func TestSummarizeGitDiff_KeepsSmallContext(t *testing.T) {
	in := makeDiff(3) // ≤6 unchanged lines per run — nothing to collapse
	if out := detect.SummarizeGitDiff(in); out != "" {
		t.Fatalf("small diff must not summarize (no win), got:\n%s", out)
	}
}

func TestIsGitDiffOutput_Negative(t *testing.T) {
	for _, s := range []string{"", "hello\nworld", "--- notes\nplain text", "commit abc123\nAuthor: x"} {
		if detect.IsGitDiffOutput(s) {
			t.Errorf("false positive on %q", s)
		}
	}
}

func TestSummarizeGitDiff_NonMatchZeroAlloc(t *testing.T) {
	in := strings.Repeat("just a plain log line without diff markers\n", 50)
	if n := testing.AllocsPerRun(20, func() { detect.SummarizeGitDiff(in) }); n != 0 {
		t.Fatalf("non-match must not allocate, got %.0f allocs", n)
	}
}

func BenchmarkSummarizeGitDiff(b *testing.B) {
	in := makeDiff(20)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeGitDiff(in)
	}
}

func BenchmarkSummarizeGitDiff_NonMatch(b *testing.B) {
	in := strings.Repeat("plain log line\n", 200)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeGitDiff(in)
	}
}
