package summary

import (
	"fmt"
	"strings"

	"github.com/alex60217101990/terse/internal/detect"
)

// ColumnarSummary returns a compact human-readable columnar summary of a JSON
// array: schema plus per-column statistics in minimal tokens. Individual row
// values are elided — this is a lossy summary. The only production caller is
// the generic tool-output pipeline (Bash/MCP), which appends a "qdf-hook expand
// <hash>" recovery pointer so the full array remains retrievable; the summary
// therefore does not, on its own, claim any particular way to fetch rows (a
// previous version wrongly told the model to "Read with offset/limit", which is
// meaningless for a command's output). The path parameter is unused and kept
// only for call-site clarity.
func ColumnarSummary(_ string, stats *detect.ArrayStats) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "[COLUMNAR SUMMARY (%d rows) — per-column stats; row values elided]\n", stats.RowCount)

	// Schema line: one compact view of all column types.
	sb.WriteString("SCHEMA: {")
	for i, col := range stats.Columns {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%s:%s", col.Name, col.Kind)
	}
	sb.WriteString("}\n")

	// Per-column detail lines.
	for _, col := range stats.Columns {
		label := col.Name + ":"
		fmt.Fprintf(&sb, "%-14s", label)
		switch col.Kind {
		case detect.KindString:
			if col.ConstVal != "" {
				fmt.Fprintf(&sb, "%q × %d", col.ConstVal, col.ConstCount)
			} else if len(col.TopVals) > 0 {
				sb.WriteString(strings.Join(col.TopVals, ", "))
				if col.Cardinality > 5 {
					fmt.Fprintf(&sb, " ... (%d distinct)", col.Cardinality)
				}
			}
		case detect.KindInt, detect.KindFloat:
			if col.Min == col.Max {
				fmt.Fprintf(&sb, "const=%.4g", col.Min)
			} else {
				fmt.Fprintf(&sb, "[%.4g..%.4g] mean=%.4g p95=%.4g", col.Min, col.Max, col.Mean, col.P95)
			}
		case detect.KindBool:
			// Observed includes rows whose value was JSON null; exclude them
			// so nulls aren't double-counted as false.
			falseCount := col.Observed - col.BoolTrue - col.NullCount
			fmt.Fprintf(&sb, "true=%d false=%d", col.BoolTrue, falseCount)
		case detect.KindNull:
			sb.WriteString("null (all rows)")
		case detect.KindMixed:
			sb.WriteString("(mixed types)")
		}
		if col.Nullable {
			fmt.Fprintf(&sb, " (nullable, %d nulls)", col.NullCount)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}
