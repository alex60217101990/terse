package detect_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

func makeBigObject() string {
	var b strings.Builder
	b.WriteString(`{"service":"widget-api","version":"2.3.1","replicas":4,"debug":false,`)
	b.WriteString(`"description":"` + strings.Repeat("long descriptive text ", 40) + `",`)
	b.WriteString(`"items":[`)
	for i := range 30 {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"id":%d,"state":"ready","zone":"eu-%d"}`, i, i%3)
	}
	b.WriteString(`],"limits":{"cpu":"2","mem":"4Gi"}}`)
	return b.String()
}

func TestSummarizeJSONObject_Basic(t *testing.T) {
	in := makeBigObject()
	out := detect.SummarizeJSONObject(in)
	if out == "" || len(out) >= len(in) {
		t.Fatalf("expected shrink, got %d vs %d", len(out), len(in))
	}
	for _, must := range []string{`"widget-api"`, `"2.3.1"`, "replicas", "items", "30", "…"} {
		if !strings.Contains(out, must) {
			t.Errorf("missing %q:\n%s", must, out)
		}
	}
}

func TestSummarizeJSONObject_Negatives(t *testing.T) {
	for i, s := range []string{
		"", "plain text", `["array","not","object"]`,
		`{"small":"object"}`, // < 1KB
		`{"broken": tru`,     // invalid JSON (>1KB via padding below)
	} {
		if i == 4 {
			s += strings.Repeat(" ", 1100)
		}
		if out := detect.SummarizeJSONObject(s); out != "" {
			t.Errorf("case %d false positive:\n%s", i, out)
		}
	}
}

func TestSummarizeJSONObject_NonMatchZeroAlloc(t *testing.T) {
	in := strings.Repeat("prose line\n", 200)
	if n := testing.AllocsPerRun(20, func() { detect.SummarizeJSONObject(in) }); n != 0 {
		t.Fatalf("non-match allocated %.0f", n)
	}
}

func BenchmarkSummarizeJSONObject_Match(b *testing.B) {
	in := makeBigObject()
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeJSONObject(in)
	}
}

func BenchmarkSummarizeJSONObject_NonMatch(b *testing.B) {
	in := strings.Repeat("prose line\n", 200)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeJSONObject(in)
	}
}
