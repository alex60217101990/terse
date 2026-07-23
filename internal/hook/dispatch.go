package hook

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/bytesconv"
	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/detect"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// Dispatch is the single PostToolUse entry point. It decodes the event once and
// routes by tool name: Read and Write keep their file-specific handlers; every
// other tool (Bash, Glob, Grep, mcp__*, anything new) flows through the generic
// pipeline so it gets dedup / delta / noise-strip / squeeze / structural
// summaries for free — no per-tool hardcoding.
func Dispatch(r io.Reader, w io.Writer) error {
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	switch inp.ToolName {
	case "Read":
		return handleRead(inp, w)
	case "Write", "Edit", "MultiEdit":
		return handleWrite(inp, w)
	default:
		return handleGeneric(inp.ToolName, inp, w)
	}
}

// skipTool reports whether a tool's output must be passed through verbatim (the
// model needs it exactly). Extend via QDF_SKIP_TOOLS (comma-separated).
func skipTool(tool string) bool {
	switch tool {
	case "TodoWrite", "ExitPlanMode":
		return true
	}
	if env := os.Getenv("QDF_SKIP_TOOLS"); env != "" {
		for _, t := range strings.Split(env, ",") {
			if strings.TrimSpace(t) == tool {
				return true
			}
		}
	}
	return false
}

// handleGeneric applies the tool-agnostic compression pipeline to any tool's
// output. All steps are never-worse (emit the compact form only when strictly
// smaller).
func handleGeneric(toolName string, inp *protocol.HookInput, w io.Writer) error {
	start := time.Now()
	if inp.ToolResponse == nil || skipTool(toolName) {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	content := detect.StripNoise(inp.ToolResponse.Text())

	record := func(action string, bytesOut int) error {
		_ = analytics.Record(analytics.Event{
			TS:       time.Now().UnixNano(),
			SID:      inp.SessionID,
			Hook:     toolName,
			Action:   action,
			BytesIn:  len(content),
			BytesOut: bytesOut,
			DurNS:    time.Since(start).Nanoseconds(),
		})
		return nil
	}

	if len(content) < 256 {
		_ = record("passthrough", len(content))
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	key := cache.LastOutputKey(toolName, inp.ToolInput)

	var action, replacement string
	switch {
	// Format-specific summarizers keyed by tool name.
	case toolName == "Glob":
		if t := buildGlobTree(content); t != "" && len(t) < len(content) {
			action, replacement = "tree", t
		}
	case toolName == "Grep":
		if g, a := buildGrepSummary(content); g != "" && len(g) < len(content) {
			action, replacement = a, g
		}
	}

	if replacement == "" {
		if s := tryJSON(content); s != "" {
			action, replacement = "summary", s
		} else if s := tryGoTest(content); s != "" {
			action, replacement = "summary", s
		} else if s := tryGitLog(content); s != "" {
			action, replacement = "summary", s
		} else if s := tryBench(content); s != "" {
			action, replacement = "summary", s
		} else if tok, ok := cache.Dedup(content, 256); ok {
			action, replacement = "ref", tok
		} else if d, ok := tryRerunDelta(toolName, key, content); ok {
			action, replacement = "rerun-delta", d
		} else if sq := detect.SqueezeOutput(content); len(sq) < len(content)*9/10 {
			action, replacement = "squeezed", sq
		} else {
			action = "passthrough"
		}
	}

	// Remember this output for the next run's delta on the unstructured paths.
	if action == "passthrough" || action == "squeezed" || action == "rerun-delta" {
		cache.LastOutputPut(key, content)
	}

	if replacement != "" {
		_ = record(action, len(replacement))
		return protocol.EncodeOutput(w, protocol.Replace(replacement))
	}
	_ = record(action, len(content))
	return protocol.EncodeOutput(w, protocol.Passthrough())
}

// tryRerunDelta returns a unified diff of content against the previous run under
// key, when strictly smaller. Diff inputs are zero-copy views (read-only).
func tryRerunDelta(toolName, key, content string) (string, bool) {
	prev, ok := cache.LastOutputGet(key)
	if !ok || prev == content {
		return "", false
	}
	diff := cache.UnifiedDiff(bytesconv.S2B(prev), bytesconv.S2B(content), 3)
	if diff == "" {
		return "", false
	}
	out := "[§rerun-delta§ " + toolName + " — changes since last run]\n" + diff
	if len(out) >= len(content) {
		return "", false
	}
	return out, true
}
