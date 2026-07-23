package hook

import (
	"io"

	"github.com/alex60217101990/qdf-hook/internal/bytesconv"
	"github.com/alex60217101990/qdf-hook/internal/detect"
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
	"github.com/alex60217101990/qdf-hook/internal/summary"
)

// minSummaryRatio: only replace tool output if the summary is at most this
// fraction of the original. Below this, compression isn't worth the overhead.
const minSummaryRatio = 0.5

// HandleBash is retained for backward compatibility. PostToolUse routing now
// goes through Dispatch, which handles Bash (and every non-Read/Write tool) via
// the generic pipeline.
func HandleBash(r io.Reader, w io.Writer) error { return Dispatch(hookcore.NewDiskStore(), r, w) }

// The try* detectors are content-sniffed and tool-agnostic: the generic pipeline
// runs them for any tool whose output matches the shape.

func tryJSON(content string) string {
	if !detect.IsJSONArray(content) {
		return ""
	}
	// Zero-copy view: AnalyzeJSONArray only reads the bytes.
	stats, err := detect.AnalyzeJSONArray(bytesconv.S2B(content), 2000)
	if err != nil || stats.RowCount < 5 {
		return ""
	}
	s := summary.ColumnarSummary("(tool output)", stats)
	if float64(len(s)) > float64(len(content))*minSummaryRatio {
		return ""
	}
	return s
}

func tryGoTest(content string) string {
	if !detect.IsGoTestOutput(content) {
		return ""
	}
	s := detect.SummarizeGoTest(content)
	if float64(len(s)) > float64(len(content))*minSummaryRatio {
		return ""
	}
	return s
}

func tryGitLog(content string) string {
	if !detect.IsGitLogOutput(content) {
		return ""
	}
	s := detect.SummarizeGitLog(content)
	if float64(len(s)) > float64(len(content))*minSummaryRatio {
		return ""
	}
	return s
}

func tryBench(content string) string {
	if !detect.IsGoBenchOutput(content) {
		return ""
	}
	s := detect.SummarizeGoBench(content)
	if float64(len(s)) > float64(len(content))*minSummaryRatio {
		return ""
	}
	return s
}
