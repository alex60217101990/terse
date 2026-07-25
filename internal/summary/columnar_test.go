package summary_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
	"github.com/alex60217101990/terse/internal/summary"
)

// TestColumnarSummary_BoolNullNotCountedFalse pins the fix for null rows being
// double-counted as false in a bool column.
func TestColumnarSummary_BoolNullNotCountedFalse(t *testing.T) {
	stats := &detect.ArrayStats{
		RowCount: 2,
		Columns: []detect.ColStats{{
			Name:      "active",
			Kind:      detect.KindBool,
			Observed:  2, // one true, one null
			BoolTrue:  1,
			NullCount: 1,
			Nullable:  true,
		}},
	}
	out := summary.ColumnarSummary("x.json", stats)
	if !strings.Contains(out, "true=1 false=0") {
		t.Errorf("null must not count as false; want 'true=1 false=0', got:\n%s", out)
	}
}
