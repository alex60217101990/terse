package hook

import (
	"io"

	"github.com/alex60217101990/terse/internal/bytesconv"
	"github.com/alex60217101990/terse/internal/detect"
	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/summary"
)

// Summary gates: only replace tool output when the summary is at most this
// fraction of the original.
//   - minSummaryRatio (strict, 0.5): very lossy transforms that discard whole
//     rows/records (columnar JSON). Demand at least a 2x win to justify the loss.
//   - minSummaryRatioLoose (0.75): mildly lossy transforms that only truncate
//     long fields / context (git log). A 25%+ win is worth keeping; the old 0.5
//     gate silently threw away real 25–49% reductions.
const (
	minSummaryRatio      = 0.5
	minSummaryRatioLoose = 0.75
)

// worth reports whether a summary is non-empty and small enough to replace the
// original at the given ratio.
func worth(summary, content string, ratio float64) bool {
	return summary != "" && float64(len(summary)) <= float64(len(content))*ratio
}

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
	if !worth(s, content, minSummaryRatio) {
		return ""
	}
	return s
}

func tryGoTest(content string) string {
	if !detect.IsGoTestOutput(content) {
		return ""
	}
	s := detect.SummarizeGoTest(content)
	if !worth(s, content, minSummaryRatio) {
		return ""
	}
	return s
}

func tryGitLog(content string) string {
	if !detect.IsGitLogOutput(content) {
		return ""
	}
	s := detect.SummarizeGitLog(content)
	if !worth(s, content, minSummaryRatioLoose) {
		return ""
	}
	return s
}

func tryBench(content string) string {
	if !detect.IsGoBenchOutput(content) {
		return ""
	}
	s := detect.SummarizeGoBench(content)
	if !worth(s, content, minSummaryRatioLoose) {
		return ""
	}
	return s
}

// tryGrep compresses grep/ripgrep "file:line:text" output run via Bash (or any
// non-Grep tool). buildGrepSummary already powers the Grep tool; this wires the
// same content-mode compressor into the generic try-chain so `rg`/`grep` in
// Bash compress too. Only "grouped" (content-mode) output is accepted — bare
// path lists ("tree") are too easily confused with ordinary `ls`/`find` output
// to fold blindly here.
func tryGrep(content string) string {
	s, action := buildGrepSummary(content)
	if action != "grouped" {
		return ""
	}
	if !worth(s, content, minSummaryRatio) {
		return ""
	}
	return s
}

// tryGitDiff folds long unchanged-context runs in unified diff output.
func tryGitDiff(content string) string {
	s := detect.SummarizeGitDiff(content)
	if !worth(s, content, minSummaryRatioLoose) {
		return ""
	}
	return s
}
