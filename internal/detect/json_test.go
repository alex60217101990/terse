package detect_test

import (
	"os"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/detect"
	"github.com/alex60217101990/qdf-hook/internal/summary"
)

func TestIsJSONArray(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{`[{"a":1}]`, true},
		{`[ { "a": 1 }, { "b": 2 } ]`, true},
		{`{"a":1}`, false}, // object, not array
		{`just text`, false},
		{`[1,2,3]`, false}, // array of scalars, not objects
		{``, false},
	}
	for _, c := range cases {
		got := detect.IsJSONArray(c.input)
		if got != c.want {
			t.Errorf("IsJSONArray(%q) = %v, want %v", c.input[:minInt(len(c.input), 30)], got, c.want)
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestAnalyzeJSONArray_1k(t *testing.T) {
	data, err := os.ReadFile("../../testdata/json_array_1k.json")
	if err != nil {
		t.Fatal(err)
	}
	stats, err := detect.AnalyzeJSONArray(data, 1000)
	if err != nil {
		t.Fatalf("AnalyzeJSONArray: %v", err)
	}
	if stats.RowCount != 1000 {
		t.Errorf("RowCount = %d, want 1000", stats.RowCount)
	}
	if len(stats.Columns) == 0 {
		t.Fatal("no columns detected")
	}
	// Find status column.
	var statusCol *detect.ColStats
	for i := range stats.Columns {
		if stats.Columns[i].Name == "status" {
			statusCol = &stats.Columns[i]
		}
	}
	if statusCol == nil {
		t.Fatal("status column not found")
	}
	if statusCol.ConstVal == "" && statusCol.Cardinality > 2 {
		t.Errorf("status should have cardinality ≤ 2, got %d", statusCol.Cardinality)
	}
}

func TestColumnarSummary_1k(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/json_array_1k.json")
	stats, _ := detect.AnalyzeJSONArray(data, 1000)
	s := summary.ColumnarSummary("testdata/json_array_1k.json", stats)
	if s == "" {
		t.Fatal("summary is empty")
	}
	// Must be dramatically shorter than source.
	if len(s) > len(data)/10 {
		t.Errorf("summary too long: %d chars vs source %d chars", len(s), len(data))
	}
	t.Log("summary:\n", s)
}

func BenchmarkAnalyzeJSONArray(b *testing.B) {
	data, _ := os.ReadFile("../../testdata/json_array_1k.json")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detect.AnalyzeJSONArray(data, 1000)
	}
}
