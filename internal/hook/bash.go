package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/cache"
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

	// Parse command and cwd from tool_input for cache lookup.
	var bi protocol.BashInput
	_ = json.Unmarshal(inp.ToolInput, &bi)

	content := inp.ToolResponse.Content

	// Bash cache fast-path for read-only commands: if the output is identical
	// to a recent cached run, return a compact §bash-unchanged§ token instead
	// of running the full detector pipeline.
	if cache.IsReadOnlyCommand(bi.Command) {
		if cached, ok := cache.BashCacheGet(bi.Command, bi.Cwd); ok {
			if cached == content {
				hashHex := cache.BashOutputHash(content)
				compact := fmt.Sprintf("§bash-unchanged:%s§ [%s] — output identical (< 30s ago)",
					hashHex, bi.Command)
				_ = analytics.Record(analytics.Event{
					TS:       time.Now().UnixNano(),
					SID:      inp.SessionID,
					Hook:     "bash",
					Action:   "bash-unchanged",
					BytesIn:  len(content),
					BytesOut: len(compact),
					DurNS:    time.Since(start).Nanoseconds(),
				})
				return protocol.EncodeOutput(w, protocol.Replace(compact))
			}
		}
		// Output changed or first run — update cache for next call.
		cache.BashCacheSet(bi.Command, bi.Cwd, content)
	}

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
