package hook

import (
	"io"
	"strings"

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

// tryJSONObject summarizes one large JSON object (config/API dump) to a key
// schema with scalar values. Strict gate + recovery footer (rows are elided).
func tryJSONObject(content string) string {
	s := detect.SummarizeJSONObject(content)
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

// grepShapeProbeLines caps how many non-empty leading lines looksLikeGrep
// inspects before deciding. Bounding the scan keeps the guard O(1)-ish on
// huge payloads instead of walking the whole content.
const grepShapeProbeLines = 8

// looksLikeGrep is a zero-alloc pre-guard for tryGrep: it scans up to the
// first grepShapeProbeLines non-empty lines with parseGrepLine (which takes
// no allocations) and requires at least half of them to parse as
// "file:line:text" content lines. This lets tryGrep bail out before calling
// buildGrepSummary, which otherwise falls into its tree fallback ->
// buildGlobTree -> topGlobDir -> strings.SplitN for ordinary (non-grep)
// payloads — a real cost paid on every generic Bash/try-chain call.
func looksLikeGrep(content string) bool {
	checked, matched := 0, 0
	for ln := range strings.SplitSeq(strings.TrimSpace(content), "\n") {
		if ln == "" {
			continue
		}
		checked++
		if _, _, _, ok := parseGrepLine(ln); ok {
			matched++
		}
		if checked >= grepShapeProbeLines {
			break
		}
	}
	return checked > 0 && matched*2 >= checked
}

// tryGrep compresses grep/ripgrep "file:line:text" output run via Bash (or any
// non-Grep tool). buildGrepSummary already powers the Grep tool; this wires the
// same content-mode compressor into the generic try-chain so `rg`/`grep` in
// Bash compress too. Only "grouped" (content-mode) output is accepted — bare
// path lists ("tree") are too easily confused with ordinary `ls`/`find` output
// to fold blindly here.
//
// looksLikeGrep gates the call to buildGrepSummary: for non-grep-shaped
// content (the common case for generic Bash output) this returns "" without
// ever reaching buildGrepSummary's tree fallback. The tree fallback stays
// reachable for the dedicated Grep tool dispatch (dispatch.go case "Grep"),
// which calls buildGrepSummary directly and is unaffected by this guard.
func tryGrep(content string) string {
	if !looksLikeGrep(content) {
		return ""
	}
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

// tryStackTrace elides the middle frames of long stack traces.
func tryStackTrace(content string) string {
	s := detect.SummarizeStackTrace(content)
	if !worth(s, content, minSummaryRatioLoose) {
		return ""
	}
	return s
}

// tryTable compresses tabular output (docker/kubectl/ls/CSV/TSV) to header +
// head/tail rows. Strict gate: it drops rows, so demand a 2x win.
func tryTable(content string) string {
	s := detect.SummarizeTable(content)
	if !worth(s, content, minSummaryRatio) {
		return ""
	}
	return s
}
