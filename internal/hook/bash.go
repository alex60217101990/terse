package hook

import (
	"fmt"
	"io"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
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
	start := time.Now()

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

	// Try detectors in priority order; collect action and replacement.
	var action string
	var replacement string

	if s := tryJSON(content); s != "" {
		action, replacement = "summary", s
	} else if s := tryGoTest(content); s != "" {
		action, replacement = "summary", s
	} else if s := tryGitLog(content); s != "" {
		action, replacement = "summary", s
	} else if s := tryBench(content); s != "" {
		action, replacement = "summary", s
	} else {
		action = "passthrough"
	}

	var out *protocol.HookOutput
	var bytesOut int
	if action == "summary" {
		out = protocol.Replace(replacement)
		bytesOut = len(replacement)
	} else {
		out = protocol.Passthrough()
		bytesOut = len(content)
	}

	// Record analytics (best-effort — never block the hook).
	_ = analytics.Record(analytics.Event{
		TS:       time.Now().UnixNano(),
		SID:      inp.SessionID,
		Hook:     "bash",
		Action:   action,
		BytesIn:  len(content),
		BytesOut: bytesOut,
		DurNS:    time.Since(start).Nanoseconds(),
	})

	return protocol.EncodeOutput(w, out)
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
