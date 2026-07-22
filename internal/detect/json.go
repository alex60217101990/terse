package detect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
)

// ColKind represents the inferred type of a JSON column.
type ColKind string

const (
	KindString ColKind = "string"
	KindInt    ColKind = "int"
	KindFloat  ColKind = "float"
	KindBool   ColKind = "bool"
	KindNull   ColKind = "null"
	KindMixed  ColKind = "mixed"
)

// ColStats holds per-column aggregate statistics.
type ColStats struct {
	Name        string
	Kind        ColKind
	ConstVal    string   // non-empty if all non-null string values are equal
	TopVals     []string // up to 5 most frequent values (strings)
	NullCount   int
	BoolTrue    int
	Observed    int     // rows where this key was present (sparse-column support)
	ConstCount  int     // how many rows have ConstVal
	Min, Max    float64 // for numeric kinds
	Mean        float64
	P95         float64 // approximate 95th percentile
	Cardinality int     // distinct count (capped at maxDistinct)
	Nullable    bool
}

// ArrayStats holds aggregate statistics for a JSON array of objects.
type ArrayStats struct {
	Columns  []ColStats
	RowCount int
}

const maxDistinct = 64 // cap cardinality tracking to save memory

// IsJSONArray returns true if s begins with '[' and its first element starts with '{'.
// This is an O(1) heuristic — it does not fully validate the JSON.
func IsJSONArray(s string) bool {
	t := strings.TrimSpace(s)
	if len(t) == 0 || t[0] != '[' {
		return false
	}
	// Scan past '[' for the first non-whitespace character.
	for i := 1; i < len(t); i++ {
		c := t[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		return c == '{'
	}
	return false
}

// AnalyzeJSONArray parses up to maxRows rows from data and returns column statistics.
// data must be a JSON array of objects (IsJSONArray check is the caller's responsibility).
func AnalyzeJSONArray(data []byte, maxRows int) (*ArrayStats, error) {
	dec := json.NewDecoder(bytes.NewReader(data))

	// Read opening '['.
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("expected '[': %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected '[', got %v", tok)
	}

	// Per-column accumulators, keyed by name; order preserved via colOrder.
	type colAcc struct {
		strFreq  map[string]int
		nums     []float64
		nulls    int
		boolTrue int
		observed int
		hasStr   bool
		hasInt   bool
		hasFloat bool
		hasBool  bool
		hasNull  bool
	}

	accs := make(map[string]*colAcc)
	var colOrder []string
	rowCount := 0

	for dec.More() && rowCount < maxRows {
		var row map[string]json.RawMessage
		if err := dec.Decode(&row); err != nil {
			return nil, fmt.Errorf("row %d: %w", rowCount, err)
		}
		rowCount++

		for name, v := range row {
			acc, exists := accs[name]
			if !exists {
				acc = &colAcc{strFreq: make(map[string]int)}
				accs[name] = acc
				colOrder = append(colOrder, name)
			}
			acc.observed++

			if len(v) == 0 {
				continue
			}
			switch v[0] {
			case '"':
				acc.hasStr = true
				var s string
				if err := json.Unmarshal(v, &s); err == nil {
					// Cap cardinality at maxDistinct distinct values.
					if len(acc.strFreq) < maxDistinct {
						acc.strFreq[s]++
					} else if _, ok := acc.strFreq[s]; ok {
						acc.strFreq[s]++
					}
				}
			case 't', 'f':
				acc.hasBool = true
				if v[0] == 't' {
					acc.boolTrue++
				}
			case 'n':
				acc.hasNull = true
				acc.nulls++
			default:
				var f float64
				if err := json.Unmarshal(v, &f); err == nil {
					if f == math.Trunc(f) {
						acc.hasInt = true
					} else {
						acc.hasFloat = true
					}
					acc.nums = append(acc.nums, f)
				}
			}
		}
	}

	stats := &ArrayStats{RowCount: rowCount, Columns: make([]ColStats, 0, len(colOrder))}
	for _, name := range colOrder {
		acc := accs[name]
		cs := ColStats{
			Name:      name,
			Nullable:  acc.nulls > 0,
			NullCount: acc.nulls,
			BoolTrue:  acc.boolTrue,
			Observed:  acc.observed,
		}

		// Determine kind; mixed if multiple base types observed.
		kinds := 0
		if acc.hasStr {
			kinds++
			cs.Kind = KindString
		}
		if acc.hasInt {
			kinds++
			cs.Kind = KindInt
		}
		if acc.hasFloat {
			kinds++
			cs.Kind = KindFloat
		}
		if acc.hasBool {
			kinds++
			cs.Kind = KindBool
		}
		if acc.hasNull && kinds == 0 {
			cs.Kind = KindNull
		}
		if kinds > 1 {
			// Pure numeric (int + float together) is still KindFloat, not Mixed.
			if !acc.hasStr && !acc.hasBool {
				cs.Kind = KindFloat
			} else {
				cs.Kind = KindMixed
			}
		}

		// String stats: cardinality, constant value, top-5 by frequency.
		if acc.hasStr && len(acc.strFreq) > 0 {
			cs.Cardinality = len(acc.strFreq)
			nonNullRows := rowCount - acc.nulls
			for val, cnt := range acc.strFreq {
				if cnt == nonNullRows {
					cs.ConstVal = val
					cs.ConstCount = cnt
				}
			}
			// Sort by frequency descending, take top 5.
			type kv struct {
				k string
				v int
			}
			top := make([]kv, 0, len(acc.strFreq))
			for k, v := range acc.strFreq {
				top = append(top, kv{k, v})
			}
			slices.SortFunc(top, func(a, b kv) int {
				return b.v - a.v
			})
			limit := min(5, len(top))
			for _, pair := range top[:limit] {
				cs.TopVals = append(cs.TopVals, fmt.Sprintf("%q×%d", pair.k, pair.v))
			}
		}

		// Numeric stats: min, max, mean, p95.
		if len(acc.nums) > 0 {
			slices.Sort(acc.nums)
			cs.Min = acc.nums[0]
			cs.Max = acc.nums[len(acc.nums)-1]

			var sum float64
			for _, n := range acc.nums {
				sum += n
			}
			cs.Mean = sum / float64(len(acc.nums))

			// P95: ceiling-index into sorted slice.
			p95idx := min(max(int(math.Ceil(0.95*float64(len(acc.nums))))-1, 0), len(acc.nums)-1)
			cs.P95 = acc.nums[p95idx]
		}

		stats.Columns = append(stats.Columns, cs)
	}
	return stats, nil
}
