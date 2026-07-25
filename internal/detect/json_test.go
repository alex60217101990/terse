package detect_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
	"github.com/alex60217101990/terse/internal/summary"
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
			t.Errorf("IsJSONArray(%q) = %v, want %v", c.input[:min(len(c.input), 30)], got, c.want)
		}
	}
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

// TestAnalyzeJSONArray_UnescapesStringValues guards the escape-aware fix: two
// encodings of the same string must count as one distinct value, not inflate
// cardinality or render escapes literally.
func TestAnalyzeJSONArray_UnescapesStringValues(t *testing.T) {
	// Interpreted string: é becomes a UTF-8 é byte sequence in the first
	// value, while \\u00e9 stays a literal JSON é escape in the second —
	// the same logical string "café" in two encodings.
	data := []byte("[{\"v\":\"café\"},{\"v\":\"caf\\u00e9\"}]")
	st, err := detect.AnalyzeJSONArray(data, 100)
	if err != nil {
		t.Fatalf("AnalyzeJSONArray: %v", err)
	}
	var col *detect.ColStats
	for i := range st.Columns {
		if st.Columns[i].Name == "v" {
			col = &st.Columns[i]
		}
	}
	if col == nil {
		t.Fatal("no 'v' column")
	}
	if col.Cardinality != 1 {
		t.Errorf(`"café" and "café" must fold to 1 distinct value, got Cardinality=%d TopVals=%v`,
			col.Cardinality, col.TopVals)
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
	for b.Loop() {
		_, _ = detect.AnalyzeJSONArray(data, 1000)
	}
}

// TestAnalyzeJSONArray_ColumnsBounded is a robustness regression: a single row
// with a huge number of distinct keys must not allocate an unbounded number of
// column accumulators. Rows are capped by maxRows, but columns were previously
// uncapped — a one-row object with millions of keys bypassed the row cap and
// drove ~100x memory amplification (model/attacker-controlled input).
func TestAnalyzeJSONArray_ColumnsBounded(t *testing.T) {
	var b []byte
	b = append(b, '[', '{')
	const n = 5000
	for i := range n {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, []byte(`"k`+strconv.Itoa(i)+`":1`)...)
	}
	b = append(b, '}', ']')

	stats, err := detect.AnalyzeJSONArray(b, 2000)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	if stats.RowCount != 1 {
		t.Fatalf("RowCount = %d, want 1", stats.RowCount)
	}
	if len(stats.Columns) > 256 {
		t.Fatalf("columns unbounded: got %d, want <= 256 (maxCols)", len(stats.Columns))
	}
}
