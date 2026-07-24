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

// TestAnalyzeJSONArray_TopVals_DeterministicOnTies guards a §ref-breaking bug:
// TopVals was sorted by count only, so string values tied at the same count
// fell out in Go map-iteration order — the same input produced different bytes
// run to run, and content-addressed dedup never fired. The sort now tiebreaks
// by key ascending. Build a column with several values all tied at count 2 and
// assert the result is identical across many analyses and in key order.
func TestAnalyzeJSONArray_TopVals_DeterministicOnTies(t *testing.T) {
	// 5 distinct "tag" values, each appearing exactly twice -> all tied.
	tags := []string{"delta", "alpha", "echo", "charlie", "bravo"}
	var b []byte
	b = append(b, '[')
	first := true
	for _, tag := range tags {
		for range 2 {
			if !first {
				b = append(b, ',')
			}
			first = false
			b = append(b, []byte(`{"tag":"`+tag+`"}`)...)
		}
	}
	b = append(b, ']')

	var want []string
	for i := range 20 {
		st, err := detect.AnalyzeJSONArray(b, 1000)
		if err != nil {
			t.Fatalf("AnalyzeJSONArray: %v", err)
		}
		var tagCol *detect.ColStats
		for j := range st.Columns {
			if st.Columns[j].Name == "tag" {
				tagCol = &st.Columns[j]
			}
		}
		if tagCol == nil {
			t.Fatal("no 'tag' column")
		}
		if i == 0 {
			want = append(want, tagCol.TopVals...)
			// All tied at count 2 -> must be key-ascending.
			expect := []string{`"alpha"×2`, `"bravo"×2`, `"charlie"×2`, `"delta"×2`, `"echo"×2`}
			if len(tagCol.TopVals) != len(expect) {
				t.Fatalf("TopVals = %v, want %v", tagCol.TopVals, expect)
			}
			for k := range expect {
				if tagCol.TopVals[k] != expect[k] {
					t.Fatalf("TopVals not key-ascending on ties: got %v, want %v", tagCol.TopVals, expect)
				}
			}
		} else if len(tagCol.TopVals) != len(want) {
			t.Fatalf("run %d: TopVals length changed: %v vs %v", i, tagCol.TopVals, want)
		} else {
			for k := range want {
				if tagCol.TopVals[k] != want[k] {
					t.Fatalf("run %d: TopVals nondeterministic: %v vs %v", i, tagCol.TopVals, want)
				}
			}
		}
	}
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
