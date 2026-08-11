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
	"github.com/alex60217101990/terse/internal/tokens"
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

	// contentTokens is memoised because tokenizing is the expensive part and the
	// same content is measured by every gate in the try-chain below. Computing it
	// per call would tokenise one payload up to nine times.
	contentTokens := -1
	countContent := func() int {
		if contentTokens < 0 {
			contentTokens = tokens.Count(content)
		}
		return contentTokens
	}

	// record takes the emitted TEXT, not its length, because tokens are the real
	// cost and cannot be recovered from a byte count. Passthrough emits the
	// content itself, so it reuses the memoised count instead of paying twice.
	record := func(action string, emitted string) {
		tokensOut := countContent()
		if len(emitted) != len(content) || emitted != content {
			tokensOut = tokens.Count(emitted)
		}
		_ = analytics.Record(analytics.Event{
			TS:        time.Now().UnixNano(),
			SID:       inp.SessionID,
			Hook:      toolName,
			Action:    action,
			BytesIn:   len(content),
			BytesOut:  len(emitted),
			TokensIn:  countContent(),
			TokensOut: tokensOut,
			DurNS:     time.Since(start).Nanoseconds(),
		})
	}

	if len(content) < 256 {
		record("passthrough", content)
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
	// This has to happen before the prefix fold below, which rewrites what the
	// model sees but must not change what the next run diffs against.
	if action == "passthrough" || action == "squeezed" || action == "rerun-delta" {
		store.LastPut(key, content)
	}

	// A file dumped through Bash — cat -n, nl, a loop echoing a header before
	// each file — is the same numbered listing a Read produces, and costs the
	// same redundant gutter tokens, but no summarizer above claims it. Thinning
	// the numbered runs is worth 2.35 points of the corpus; the strict
	// whole-payload check the Read path uses would take only 1.35 of that,
	// because a single header line disqualifies the whole payload.
	//
	// It runs before the fold below, not after: dropping the gutter is what
	// leaves neighbouring lines with a shared head for the fold to find.
	if action == "passthrough" {
		if thinned := detect.ThinLineNumberRuns(content); thinned != "" &&
			tokens.Count(thinned) < countContent()*(100-gutterThinMarginPct)/100 {
			action, replacement = "gutter-thinned", thinned
		}
	}

	// Last resort for everything no summarizer claimed. Command output is
	// line-oriented and incrementally similar — consecutive lines sharing a
	// directory, a package, a log prefix, an indent — and folding those shared
	// heads is worth 11.5% of the Bash output that currently passes through
	// untouched, which is the single largest uncompressed category left.
	//
	// The gate is in tokens, not bytes. A byte-smaller replacement that costs
	// the model more tokens is exactly the trap the Write/Edit marker fell into.
	//
	// It also has to clear a margin, not merely break even. Folding changes the
	// shape of what the model reads, and that is a real cost paid in exchange
	// for tokens; a fold that returns 0.2% is the reader paying for nothing.
	//
	// Passthrough only. Tried on the summarizers' output too — grouped, tree,
	// grep — and it folded exactly nothing: those already emit one line per
	// group with the shared directory hoisted by FoldPathPrefix, so there is no
	// run of near-identical lines left for this to find. The two transforms
	// cover different shapes rather than overlapping.
	if action == "passthrough" || action == "gutter-thinned" {
		src := content
		count := countContent()
		if replacement != "" {
			src, count = replacement, tokens.Count(replacement)
		}
		if folded := detect.FoldLinePrefixes(src); folded != "" &&
			tokens.Count(folded) < count*(100-prefixFoldMarginPct)/100 {
			action, replacement = "prefix-folded", folded
		}
	}

	if replacement != "" {
		record(action, replacement)
		return protocol.EncodeOutput(w, protocol.Replace(replacement))
	}
	record(action, content)
	return protocol.EncodeOutput(w, protocol.Passthrough())
}

// prefixFoldMarginPct is the token saving a prefix fold must clear before it is
// worth reshaping what the model reads. Measured over one project's archive,
// this drops the folds returning under a percent — several MCP readers and a
// WebFetch — while keeping every one that pays double digits.
const prefixFoldMarginPct = 5

// gutterThinMarginPct is the same idea for the line-number thin: numbers a
// reader may want to count from are only worth dropping for a real saving. The
// runs that clear it in practice return around 20%.
const gutterThinMarginPct = 5

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

// dedupWithStore is the store-backed dedup: it replaces content that was
// already emitted (this or an earlier session, byte-identical) with a compact
// §ref token, or registers it and returns ("", false) so the caller emits it in
// full this first time. minSize gates tiny outputs where a ~60-byte token would
// not pay off. Blobs are keyed by cache.RefHashOf (sha256[:16] hex), the same
// content-address the ref store uses everywhere.
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
	hash := refTokenFor(store, raw)
	var sb strings.Builder
	sb.Grow(len(summary) + len(hash) + 32) // + "[full output: qdf-hook expand " + "]\n"
	sb.WriteString(summary)
	sb.WriteString("[full output: qdf-hook expand ")
	sb.WriteString(hash)
	sb.WriteString("]\n")
	return sb.String()
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
