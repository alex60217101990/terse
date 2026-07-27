package detect_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

func makePyTrace(frames int) string {
	var b strings.Builder
	b.WriteString("Traceback (most recent call last):\n")
	for i := range frames {
		fmt.Fprintf(&b, "  File \"/app/pkg/module_%d.py\", line %d, in handler_%d\n    result = compute(x)\n", i, 10+i, i)
	}
	b.WriteString("ValueError: widget id out of range\n")
	return b.String()
}

func makeGoPanic(frames int) string {
	var b strings.Builder
	b.WriteString("panic: runtime error: index out of range [7] with length 3\n\ngoroutine 1 [running]:\n")
	for i := range frames {
		fmt.Fprintf(&b, "example.com/pkg.func%d(0x%x)\n\t/src/pkg/file%d.go:%d +0x%x\n", i, i, i, 10+i, i)
	}
	return b.String()
}

func TestSummarizeStackTrace_Python(t *testing.T) {
	in := makePyTrace(30)
	out := detect.SummarizeStackTrace(in)
	if out == "" || len(out) >= len(in)/2 {
		t.Fatalf("expected >=2x shrink, got %d vs %d", len(out), len(in))
	}
	for _, must := range []string{"Traceback", "ValueError: widget id out of range", "module_0.py", "module_29.py", "⋯"} {
		if !strings.Contains(out, must) {
			t.Errorf("missing %q:\n%s", must, out)
		}
	}
}

func TestSummarizeStackTrace_GoPanic(t *testing.T) {
	in := makeGoPanic(40)
	out := detect.SummarizeStackTrace(in)
	if out == "" {
		t.Fatal("go panic should summarize")
	}
	if !strings.Contains(out, "panic: runtime error: index out of range") {
		t.Errorf("panic message must survive verbatim:\n%s", out)
	}
}

func TestSummarizeStackTrace_SmallStaysVerbatim(t *testing.T) {
	if out := detect.SummarizeStackTrace(makePyTrace(4)); out != "" {
		t.Fatalf("small trace must pass through, got:\n%s", out)
	}
}

func TestSummarizeStackTrace_NonMatchZeroAlloc(t *testing.T) {
	in := strings.Repeat("ordinary build output line\n", 60)
	if n := testing.AllocsPerRun(20, func() { detect.SummarizeStackTrace(in) }); n != 0 {
		t.Fatalf("non-match allocated %.0f", n)
	}
}

func BenchmarkSummarizeStackTrace_Python(b *testing.B) {
	in := makePyTrace(30)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeStackTrace(in)
	}
}

func BenchmarkSummarizeStackTrace_GoPanic(b *testing.B) {
	in := makeGoPanic(40)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeStackTrace(in)
	}
}

func BenchmarkSummarizeStackTrace_NonMatch(b *testing.B) {
	in := strings.Repeat("ordinary build output line\n", 200)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeStackTrace(in)
	}
}
