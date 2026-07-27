package detect_test

import (
	"strings"
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
