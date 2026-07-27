package detect_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

func TestFoldPathPrefix_Folds(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 12; i++ {
		b.WriteString("/Users/dev/work/src/github.com/acme/widget-service/internal/pkg/file")
		b.WriteByte(byte('a' + i))
		b.WriteString(".go\n")
	}
	in := b.String()
	out := detect.FoldPathPrefix(in)
	if len(out) >= len(in) {
		t.Fatalf("expected shrink, got %d vs %d", len(out), len(in))
	}
	if !strings.Contains(out, "§P=/Users/dev/work/src/github.com/acme/widget-service/internal/pkg§") {
		t.Errorf("prefix declaration missing:\n%s", out)
	}
	if !strings.Contains(out, "§P§/filea.go") {
		t.Errorf("substituted line missing:\n%s", out)
	}
}

func TestFoldPathPrefix_NoCommonPrefixUnchanged(t *testing.T) {
	in := "/a/x.go\n/b/y.go\n/c/z.go\nrandom prose line\n"
	if out := detect.FoldPathPrefix(in); out != in {
		t.Fatalf("no shared prefix — must be unchanged")
	}
}

func TestFoldPathPrefix_MixedLinesPreserved(t *testing.T) {
	var b strings.Builder
	b.WriteString("header: results below\n")
	for i := 0; i < 8; i++ {
		b.WriteString("  /very/long/shared/directory/prefix/of/paths/entry")
		b.WriteByte(byte('0' + i))
		b.WriteString(".txt: 12 matches\n")
	}
	in := b.String()
	out := detect.FoldPathPrefix(in)
	if !strings.Contains(out, "header: results below") {
		t.Errorf("non-path lines must survive verbatim:\n%s", out)
	}
	if len(out) >= len(in) {
		t.Fatalf("expected shrink")
	}
}

// If the input already contains our own token literally (e.g. it's quoting
// earlier compressed output, or someone's log message happens to include
// it), folding must bail out entirely — otherwise the folded "§P§" line and
// content that already read "§P§" verbatim become indistinguishable, and
// mechanical reconstruction (strip decl line, replace §P§→prefix) would
// corrupt the line that was never touched.
func TestFoldPathPrefix_AlreadyContainsToken_Unchanged(t *testing.T) {
	var b strings.Builder
	for i := range 8 {
		b.WriteString("/very/long/shared/directory/prefix/of/paths/entry")
		b.WriteByte(byte('0' + i))
		b.WriteString(".txt: 12 matches\n")
	}
	b.WriteString("note: token §P§ appears verbatim here\n")
	in := b.String()
	if out := detect.FoldPathPrefix(in); out != in {
		t.Fatalf("input already contains §P§ — must be byte-identical, unchanged:\ngot:\n%s\nwant:\n%s", out, in)
	}
}

// Same guard, but the literal occurrence is the declaration-line form
// ("§P=") rather than the substitution token ("§P§").
func TestFoldPathPrefix_AlreadyContainsDeclToken_Unchanged(t *testing.T) {
	var b strings.Builder
	for i := range 8 {
		b.WriteString("/very/long/shared/directory/prefix/of/paths/entry")
		b.WriteByte(byte('0' + i))
		b.WriteString(".txt: 12 matches\n")
	}
	b.WriteString("note: mentions §P=something here\n")
	in := b.String()
	if out := detect.FoldPathPrefix(in); out != in {
		t.Fatalf("input already contains §P= — must be byte-identical, unchanged:\ngot:\n%s\nwant:\n%s", out, in)
	}
}

func TestFoldPathPrefix_NonMatchZeroAlloc(t *testing.T) {
	in := strings.Repeat("prose without slashes here\n", 30)
	if n := testing.AllocsPerRun(20, func() { detect.FoldPathPrefix(in) }); n != 0 {
		t.Fatalf("non-match allocated %.0f", n)
	}
}

func BenchmarkFoldPathPrefix_Match(b *testing.B) {
	var sb strings.Builder
	for i := range 40 {
		sb.WriteString("/Users/dev/work/src/github.com/acme/widget-service/internal/pkg/file")
		sb.WriteByte(byte('a' + i%26))
		sb.WriteString(".go\n")
	}
	in := sb.String()
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.FoldPathPrefix(in)
	}
}

func BenchmarkFoldPathPrefix_NonMatch(b *testing.B) {
	in := strings.Repeat("prose without slashes here\n", 40)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.FoldPathPrefix(in)
	}
}
