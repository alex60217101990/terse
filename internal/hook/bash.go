package hook

import (
	"fmt"
	"io"

	"github.com/alex60217101990/qdf-hook/internal/detect"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
	"github.com/alex60217101990/qdf-hook/internal/summary"
)

// minSummaryRatio: only replace tool output if the summary is at most this fraction of the original.
// Below this threshold, compression isn't worth the overhead.
const minSummaryRatio = 0.5

// HandleBash processes a PostToolUse hook call for the Bash tool.
// It tries each detector in priority order; falls back to pass-through.
func HandleBash(r io.Reader, w io.Writer) error {
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	if inp.ToolResponse == nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	content := inp.ToolResponse.Content
	if len(content) < 256 {
		// Too small to bother compressing.
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Try detectors in priority order.
	if s := tryJSON(content); s != "" {
		return protocol.EncodeOutput(w, protocol.Replace(s))
	}
	if s := tryGoTest(content); s != "" {
		return protocol.EncodeOutput(w, protocol.Replace(s))
	}
	if s := tryGitLog(content); s != "" {
		return protocol.EncodeOutput(w, protocol.Replace(s))
	}
	if s := tryBench(content); s != "" {
		return protocol.EncodeOutput(w, protocol.Replace(s))
	}

	// No detector matched — pass through.
	return protocol.EncodeOutput(w, protocol.Passthrough())
}

func tryJSON(content string) string {
	if !detect.IsJSONArray(content) {
		return ""
	}
	stats, err := detect.AnalyzeJSONArray([]byte(content), 2000)
	if err != nil || stats.RowCount < 5 {
		return ""
	}
	s := summary.ColumnarSummary("(bash output)", stats)
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
