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
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// Dispatch is the single PostToolUse entry point. It decodes the event once and
// routes by tool name: Read and Write keep their file-specific handlers; every
// other tool (Bash, Glob, Grep, mcp__*, anything new) flows through the generic
// pipeline so it gets dedup / delta / noise-strip / squeeze / structural
// summaries for free — no per-tool hardcoding. store is the pipeline's session
// and cache backend (the on-disk cache for the CLI, an in-memory store for a
// future daemon).
func Dispatch(store hookcore.StateStore, r io.Reader, w io.Writer) error {
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	return routeInput(store, inp, w)
}

// DispatchBytes is Dispatch for callers that already hold the whole request in
// memory (the daemon, via a pooled read buffer): it decodes with a single
// json.Unmarshal instead of wrapping the bytes in a json.Decoder, avoiding
// that decoder's extra buffering/streaming layer on the hottest path. Pool
// safety is unchanged — json.Unmarshal, like the Decoder, copies every string
// out of req, so nothing routeInput retains aliases the pooled bytes.
func DispatchBytes(store hookcore.StateStore, req []byte, w io.Writer) error {
	inp, err := protocol.DecodeInputBytes(req)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	return routeInput(store, inp, w)
}

func routeInput(store hookcore.StateStore, inp *protocol.HookInput, w io.Writer) error {
	switch inp.ToolName {
	case "Read":
		return handleRead(store, inp, w)
	case "Write", "Edit", "MultiEdit":
		return handleWrite(store, inp, w)
	default:
		return handleGeneric(store, inp.ToolName, inp, w)
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
		for t := range strings.SplitSeq(env, ",") {
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
func handleGeneric(store hookcore.StateStore, toolName string, inp *protocol.HookInput, w io.Writer) error {
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
		} else if tok, ok := dedupWithStore(store, content, 256); ok {
			action, replacement = "ref", tok
		} else if d, ok := tryRerunDelta(store, toolName, key, content); ok {
			action, replacement = "rerun-delta", d
		} else if sq := detect.SqueezeOutput(content); len(sq) < len(content)*9/10 {
			action, replacement = "squeezed", sq
		} else {
			action = "passthrough"
		}
	}

	// Remember this output for the next run's delta on the unstructured paths.
	if action == "passthrough" || action == "squeezed" || action == "rerun-delta" {
		store.LastPut(key, content)
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
func tryRerunDelta(store hookcore.StateStore, toolName, key, content string) (string, bool) {
	prev, ok := store.LastGet(key)
	if !ok || prev == content {
		return "", false
	}
	// Guard: UnifiedDiff (Myers) is O((N+M)²) in line count — skip it for very
	// large reruns rather than risk gigabytes of transient allocation or a hang
	// that would take down every daemon connection. Mirrors read.go's
	// serveDelta guard; the caller then falls back to squeeze/passthrough.
	if strings.Count(prev, "\n")+strings.Count(content, "\n") > 10000 {
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

// dedupWithStore is the store-backed equivalent of cache.Dedup: it replaces
// content that was already emitted (this or an earlier session, byte-
// identical) with a compact §ref token, or registers it and returns ("",
// false) so the caller emits it in full this first time. minSize gates tiny
// outputs where a ~60-byte token would not pay off. The token format and hash
// (cache.RefHashOf, the same sha256[:16] hex cache.Dedup uses) match
// cache.Dedup exactly for parity.
func dedupWithStore(store hookcore.StateStore, content string, minSize int) (token string, deduped bool) {
	if len(content) < minSize {
		return "", false
	}
	hash := cache.RefHashOf(content)
	if store.RefSeen(hash) {
		return fmt.Sprintf("§ref:%s§ (%d bytes, identical to earlier output — qdf-hook expand %s)",
			hash, len(content), hash), true
	}
	store.RefPut(hash, content)
	return "", false
}
