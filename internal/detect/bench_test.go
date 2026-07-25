package detect_test

import (
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/alex60217101990/terse/internal/detect"
)

func TestIsGoBenchOutput(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/bench_output.txt")
	if !detect.IsGoBenchOutput(string(data)) {
		t.Error("bench_output.txt should be detected as benchmark output")
	}
}

func TestSummarizeGoBench(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/bench_output.txt")
	s := detect.SummarizeGoBench(string(data))
	if !strings.Contains(s, "ns/op") {
		t.Errorf("bench summary should contain ns/op: %s", s)
	}
	if !strings.Contains(s, "Benchmark") {
		t.Errorf("bench summary should contain benchmark names: %s", s)
	}
}

// TestSummarizeGoBench_TruncatesOnRuneBoundary is the B4 regression: a long
// benchmark name with multi-byte runes must be truncated on a rune boundary,
// never byte-sliced into invalid UTF-8.
func TestSummarizeGoBench_TruncatesOnRuneBoundary(t *testing.T) {
	name := "Benchmark" + strings.Repeat("é", 40) // 49 runes, multi-byte
	in := name + "-8\t1000\t123 ns/op\t45 B/op\t2 allocs/op\n"
	s := detect.SummarizeGoBench(in)
	if !utf8.ValidString(s) {
		t.Fatalf("bench summary must stay valid UTF-8 after truncation: %q", s)
	}
	if !strings.Contains(s, "...") {
		t.Errorf("expected the long name to be truncated with '...': %q", s)
	}
}
