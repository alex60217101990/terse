package detect_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

func makeFixedTable(rows int) string {
	var b strings.Builder
	b.WriteString("CONTAINER ID   IMAGE          COMMAND        STATUS         PORTS\n")
	for i := range rows {
		fmt.Fprintf(&b, "%012x   img-%02d:latest  \"/bin/run\"     Up %d hours     8080/tcp\n", i, i%20, i%24)
	}
	return b.String()
}

func makeCSV(rows int) string {
	var b strings.Builder
	b.WriteString("id,name,status,region,latency_ms\n")
	for i := range rows {
		fmt.Fprintf(&b, "%d,widget-%d,ok,eu-%d,%d\n", i, i, i%3, 10+i%90)
	}
	return b.String()
}

func TestSummarizeTable_FixedWidth(t *testing.T) {
	in := makeFixedTable(40)
	out := detect.SummarizeTable(in)
	if out == "" || len(out) >= len(in)/2 {
		t.Fatalf("expected >=2x shrink, got %d vs %d", len(out), len(in))
	}
	if !strings.Contains(out, "CONTAINER ID") {
		t.Errorf("header must survive:\n%s", out)
	}
	if !strings.Contains(out, "40 rows") {
		t.Errorf("row count must be stated:\n%s", out)
	}
	// first and last data rows survive verbatim
	if !strings.Contains(out, "img-00:latest") || !strings.Contains(out, fmt.Sprintf("img-%02d:latest", 39%20)) {
		t.Errorf("first/last rows must survive:\n%s", out)
	}
}

func TestSummarizeTable_CSV(t *testing.T) {
	in := makeCSV(50)
	out := detect.SummarizeTable(in)
	if out == "" || len(out) >= len(in)/2 {
		t.Fatalf("expected >=2x shrink, got %d vs %d", len(out), len(in))
	}
	if !strings.Contains(out, "id,name,status,region,latency_ms") {
		t.Errorf("header must survive:\n%s", out)
	}
}

func TestSummarizeTable_Negatives(t *testing.T) {
	cases := []string{
		"",
		"short\ntable\n", // <6 lines
		strings.Repeat("prose sentence without columns\n", 10), // no delimiter/alignment
		makeFixedTable(4),                      // too few rows
		"a,b\n1,2\n3\n4,5,6\n7,8\n9,10\nx,y\n", // inconsistent column count
	}
	for i, s := range cases {
		if out := detect.SummarizeTable(s); out != "" {
			t.Errorf("case %d: false positive:\n%s", i, out)
		}
	}
}

func TestSummarizeTable_NonMatchZeroAlloc(t *testing.T) {
	in := strings.Repeat("plain prose line with no structure at all\n", 40)
	if n := testing.AllocsPerRun(20, func() { detect.SummarizeTable(in) }); n != 0 {
		t.Fatalf("non-match allocated %.0f", n)
	}
}

func BenchmarkSummarizeTable(b *testing.B) {
	in := makeFixedTable(40)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeTable(in)
	}
}

func BenchmarkSummarizeTable_NonMatch(b *testing.B) {
	in := strings.Repeat("plain prose line with no structure at all\n", 40)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeTable(in)
	}
}
