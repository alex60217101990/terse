package hook

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alex60217101990/terse/internal/analytics"
	"github.com/alex60217101990/terse/internal/bytesconv"
	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/detect"
	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/protocol"
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
	// PreToolUse (the Read mtime fast-path) is distinguished by the event name
	// Claude sends in the payload — the daemon multiplexes both events over one
	// socket, unlike the CLI where the subcommand picked the handler. Anything
	// else is a PostToolUse event, routed by tool name.
	if inp.HookEventName == "PreToolUse" {
		return handlePreToolUse(store, inp, w)
	}
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
	if inp.ToolResponse == nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}
	if skipTool(toolName) {
		// Passthrough by policy — still record it so stats see the invocation.
		raw := inp.ToolResponse.Text()
		_ = analytics.Record(analytics.Event{
			TS: time.Now().UnixNano(), SID: inp.SessionID, Hook: toolName,
			Action: "skip", BytesIn: len(raw), BytesOut: len(raw),
			DurNS: time.Since(start).Nanoseconds(),
		})
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	content := detect.StripNoise(inp.ToolResponse.Text())

	record := func(action string, bytesOut int) {
		_ = analytics.Record(analytics.Event{
			TS:       time.Now().UnixNano(),
			SID:      inp.SessionID,
			Hook:     toolName,
			Action:   action,
			BytesIn:  len(content),
			BytesOut: bytesOut,
			DurNS:    time.Since(start).Nanoseconds(),
		})
	}

	if len(content) < 256 {
		record("passthrough", len(content))
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	key := cache.LastOutputKey(toolName, inp.ToolInput)

	var action, replacement string
	// Format-specific summarizers keyed by tool name.
	switch toolName {
	case "Glob":
		if t := buildGlobTree(content); t != "" && len(t) < len(content) {
			action, replacement = "tree", t
		}
	case "Grep":
		if g, a := buildGrepSummary(content); g != "" && len(g) < len(content) {
			// "grouped" content-mode summaries elide per-file lines past
			// grepFileCap — same lossy shape as the generic tryGrep branch, so
			// make them recoverable too. "tree" (files_with_matches) drops
			// nothing and needs no footer.
			if a == "grouped" {
				g = withRecovery(store, g, content)
			}
			action, replacement = a, g
		}
	}

	if replacement == "" {
		if s := tryJSON(content); s != "" {
			// Columnar is the most lossy transform (all rows dropped). Register
			// the raw array so the model can recover it, and point at it — the
			// summary alone is otherwise unrecoverable off the Read path.
			action, replacement = "columnar", withRecovery(store, s, content)
		} else if s := tryJSONObject(content); s != "" {
			action, replacement = "jsonobject", withRecovery(store, s, content)
		} else if s := tryGoTest(content); s != "" {
			action, replacement = "summary", s
		} else if s := tryGrep(content); s != "" {
			// buildGrepSummary elides per-file lines past grepFileCap, so the
			// summary is lossy: register the raw output and point at it. This
			// is the safety net for colon-delimited config lines that pass the
			// path-char test (e.g. "db.host:5432:desc") — nothing is
			// unrecoverably dropped.
			action, replacement = "grep", withRecovery(store, s, content)
		} else if s := tryGitDiff(content); s != "" {
			action, replacement = "gitdiff", withRecovery(store, s, content)
		} else if s := tryGitLog(content); s != "" {
			action, replacement = "summary", s
		} else if s := tryBench(content); s != "" {
			action, replacement = "summary", s
		} else if s := tryStackTrace(content); s != "" {
			action, replacement = "stacktrace", withRecovery(store, s, content)
		} else if s := tryTable(content); s != "" {
			action, replacement = "table", withRecovery(store, s, content)
		} else if tok, ok := dedupWithStore(store, content, 256); ok {
			action, replacement = "ref", tok
		} else if d, ok := tryRerunDelta(store, toolName, key, content); ok {
			action, replacement = "rerun-delta", d
		} else if sq := detect.SqueezeOutput(detect.FoldRepeatedBlocks(content)); len(sq) < len(content)*9/10 {
			// Fold non-adjacent duplicate blocks (e.g. an MCP batch that
			// re-dumps the same section under several query headers), then
			// run-length/ANSI squeeze the result. Both are never-worse.
			action, replacement = "squeezed", sq
		} else {
			action = "passthrough"
		}
	}

	// Path-dense outputs: fold the shared directory prefix (lossless).
	if action == "tree" || action == "grep" || action == "grouped" {
		if folded := detect.FoldPathPrefix(replacement); len(folded) < len(replacement) {
			replacement = folded
		}
	}

	// Remember this output for the next run's delta on the unstructured paths.
	// "grep" is included so a rerun of the same grep can delta against the
	// prior raw output (the summary itself is lossy and can't be diffed).
	if action == "passthrough" || action == "squeezed" || action == "rerun-delta" || action == "grep" {
		store.LastPut(key, content)
	}

	if replacement != "" {
		record(action, len(replacement))
		return protocol.EncodeOutput(w, protocol.Replace(replacement))
	}
	record(action, len(content))
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
// refTokenFor registers content in the ref store (idempotently) and returns its
// hash, so a lossy summary can point the model at the recoverable original via
// `qdf-hook expand <hash>`. Unlike dedupWithStore it always stores and returns a
// hash — it is for making a summary recoverable, not for skipping duplicate
// output.
func refTokenFor(store hookcore.StateStore, content string) string {
	hash := cache.RefHashOf(content)
	if !store.RefSeen(hash) {
		store.RefPut(hash, content)
	}
	return hash
}

// withRecovery appends the standard lossy-summary recovery footer: the raw
// output is registered in the ref store and the summary points at it.
func withRecovery(store hookcore.StateStore, summary, raw string) string {
	return summary + fmt.Sprintf("[full output: qdf-hook expand %s]\n", refTokenFor(store, raw))
}

func dedupWithStore(store hookcore.StateStore, content string, minSize int) (token string, deduped bool) {
	if len(content) < minSize {
		return "", false
	}
	hash := cache.RefHashOf(content)
	if store.RefSeen(hash) {
		store.RefHit(hash) // record usage for eviction
		return fmt.Sprintf("§ref:%s§ (%d bytes, identical to earlier output — qdf-hook expand %s)",
			hash, len(content), hash), true
	}
	store.RefPut(hash, content)
	return "", false
}
