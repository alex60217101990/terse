package hook

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
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

// HandleRead processes a PostToolUse hook call for the Read tool.
// It reads the hook JSON from r, applies delta/unchanged logic, and writes
// the hook output JSON to w.
func HandleRead(r io.Reader, w io.Writer) error {
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	return handleRead(hookcore.NewDiskStore(), inp, w)
}

// handleRead is the Read logic over an already-decoded input (so Dispatch can
// route without re-decoding).
func handleRead(store hookcore.StateStore, inp *protocol.HookInput, w io.Writer) error {
	start := time.Now()
	if inp.ToolResponse == nil {
		// No tool response (error case from Claude) — pass through.
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Parse file path from tool_input.
	var ti protocol.ReadInput
	if err := json.Unmarshal(inp.ToolInput, &ti); err != nil || ti.FilePath == "" {
		// Cannot parse path — pass through.
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Zero-copy view: content is only read (hashed, diffed, cached) within this
	// call; the cache copies it out on Save. Text() resolves the Read tool's
	// nested file.content (top-level "content" is empty for Read).
	content := bytesconv.S2B(inp.ToolResponse.Text())

	// finish is the single exit: thin the line-number gutter off anything being
	// passed through, then record and encode. Every branch below routes through
	// it, including the windowed-read branch that used to return without
	// recording at all — which is why windowed reads showed up as an unlabelled
	// 169k tokens in the replay report.
	finish := func(out *protocol.HookOutput, action string) error {
		if thinned, ok := thinGutter(out, bytesconv.B2S(content)); ok {
			out, action = thinned, action+"-thin"
		}
		recordRead(inp, start, content, out, action)
		return protocol.EncodeOutput(w, out)
	}

	// A windowed read returns only a slice of the file, not the whole file.
	// Caching that slice as the file's entry would poison the delta/unchanged
	// logic — a later full read (or a Write that caches the raw file) would
	// mismatch the window's hash and emit a bogus "§delta — changes since last
	// read" that actually reflects the window, not a change. Detect a window
	// from the authoritative file metadata (StartLine>1 or NumLines<TotalLines)
	// and, as a fallback when that metadata is absent, from the requested
	// offset/limit. Pass a partial read straight through: don't cache, don't diff.
	partial := ti.Offset != 0 || ti.Limit != 0
	if f := inp.ToolResponse.File; f != nil &&
		(f.StartLine > 1 || (f.TotalLines > 0 && f.NumLines < f.TotalLines)) {
		partial = true
	}
	if partial {
		if out := serveWindowUnchanged(store, inp, &ti); out != nil {
			return finish(out, "unchanged-window")
		}
		return finish(protocol.Passthrough(), "window")
	}

	// Stat the file to capture its modification time and ctime.
	var modTime, ctimeNS int64
	if info, err := os.Stat(ti.FilePath); err == nil {
		modTime = info.ModTime().UnixNano()
		ctimeNS = statCtimeNS(info)
	}

	// Binary content: always pass through, no diff.
	if cache.IsBinaryContent(content) {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Load session state.
	state := store.LoadSession(ContextKey(inp))
	if state == nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}
	state.Turn++

	hash := sha256.Sum256(content)
	entry, seen := state.Files[ti.FilePath]

	var out *protocol.HookOutput
	var action string

	switch {
	case !seen || !state.SeenAfterCompact(ti.FilePath):
		// First read (or first after compaction) — pass the content through
		// unchanged. No token overhead; the cache is still populated below so
		// later reads become §unchanged§/delta. (A previous version prepended a
		// registration header, which made every first read net-negative.)
		out = protocol.Passthrough()
		action = "full"

	case entry.Hash == hash:
		// Same content — serve §unchanged§ marker.
		out = serveUnchanged(ti.FilePath, hash, entry.Turn)
		action = "unchanged"

	default:
		// Content changed — serve §delta§ with unified diff.
		out = serveDelta(ti.FilePath, hash, entry.Content, content)
		action = "delta"
	}

	// Never-worse: if the compact form (unchanged marker / delta) is not
	// actually smaller than the raw content, pass the content through instead —
	// otherwise a tiny file's marker can be longer than the file itself.
	if out.HookSpecificOutput != nil && len(out.HookSpecificOutput.UpdatedToolOutput) >= len(content) {
		out = protocol.Passthrough()
		action = "passthrough"
	}

	// Update cache entry with read tracking fields.
	updatedEntry := state.Files[ti.FilePath]
	updatedEntry.Hash = hash
	updatedEntry.Turn = state.Turn
	updatedEntry.Content = content
	updatedEntry.ReadCount++
	updatedEntry.LastReadAt = time.Now().Unix()
	updatedEntry.ModTime = modTime
	updatedEntry.CtimeNS = ctimeNS
	state.Files[ti.FilePath] = updatedEntry

	store.SaveSession(ContextKey(inp), state)

	return finish(out, action)
}

// thinGutter turns a passthrough into a thinned-gutter replacement, but only
// when that actually costs the model fewer tokens.
//
// It touches passthroughs only. The unchanged/delta markers are already far
// smaller than the content they replace, and a delta is a unified diff whose
// line numbers live in its hunk headers, not in a gutter.
func thinGutter(out *protocol.HookOutput, content string) (*protocol.HookOutput, bool) {
	if out.HookSpecificOutput != nil {
		return out, false
	}
	thinned := detect.ThinLineNumbers(content)
	if thinned == "" || tokens.Count(thinned) >= tokens.Count(content) {
		return out, false
	}
	return protocol.Replace(thinned), true
}

// recordRead writes the analytics event for one Read. Passthrough emits the
// full content, so its emitted size is len(content) — a neutral 0% saving, not
// 0, which would falsely read as 100% saved.
func recordRead(inp *protocol.HookInput, start time.Time, content []byte, out *protocol.HookOutput, action string) {
	emitted := bytesconv.B2S(content)
	if out.HookSpecificOutput != nil {
		emitted = out.HookSpecificOutput.UpdatedToolOutput
	}
	tokensIn := tokens.Count(bytesconv.B2S(content))
	tokensOut := tokensIn
	if len(emitted) != len(content) {
		tokensOut = tokens.Count(emitted)
	}
	_ = analytics.Record(analytics.Event{
		TS:        time.Now().UnixNano(),
		SID:       inp.SessionID,
		Hook:      inp.ToolName, // canonical Claude tool name (e.g. "Read")
		Action:    action,
		BytesIn:   len(content),
		BytesOut:  len(emitted),
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		DurNS:     time.Since(start).Nanoseconds(),
	})
}

func serveUnchanged(path string, hash [32]byte, cachedAtTurn int) *protocol.HookOutput {
	hashHex := cache.ShortHex(hash[:8])
	var sb strings.Builder
	sb.Grow(len(hashHex) + len(path) + 96)
	sb.WriteString("[READ §unchanged:")
	sb.WriteString(hashHex)
	sb.WriteString("§ ")
	sb.WriteString(path)
	sb.WriteString(" — content identical to read at turn ")
	sb.WriteString(strconv.Itoa(cachedAtTurn))
	sb.WriteString(". Full content available if needed.]")
	return protocol.Replace(sb.String())
}

func serveDelta(path string, newHash [32]byte, oldContent, newContent []byte) *protocol.HookOutput {
	// Guard: very large diffs are O((N+M)²) — pass the full content through.
	if bytes.Count(oldContent, []byte("\n"))+bytes.Count(newContent, []byte("\n")) > 10000 {
		return protocol.Passthrough()
	}
	diff := cache.UnifiedDiff(oldContent, newContent, 3)
	hashHex := cache.ShortHex(newHash[:8])

	var sb strings.Builder
	sb.Grow(len(diff) + len(path) + 64)
	sb.WriteString("[READ §delta:")
	sb.WriteString(hashHex)
	sb.WriteString("§ ")
	sb.WriteString(path)
	sb.WriteString(" — showing changes since last read]\n")
	if diff == "" {
		// Shouldn't happen (hashes differ but diff is empty) — serve full.
		sb.Write(newContent)
	} else {
		sb.WriteString(diff)
	}
	return protocol.Replace(sb.String())
}

// serveWindowUnchanged answers a windowed Read from the cached full file when
// NOTHING changed: same mtime AND ctime as the cache entry AND the window's
// bytes equal the corresponding slice of the cached content. Any doubt → nil
// (caller passes through). Never caches, never mutates session state.
func serveWindowUnchanged(store hookcore.StateStore, inp *protocol.HookInput, ti *protocol.ReadInput) *protocol.HookOutput {
	f := inp.ToolResponse.File
	if f == nil || f.StartLine < 1 || f.NumLines <= 0 {
		return nil
	}
	info, err := os.Stat(ti.FilePath)
	if err != nil {
		return nil
	}
	state := store.LoadSession(ContextKey(inp))
	if state == nil {
		return nil
	}
	entry, ok := state.Files[ti.FilePath]
	if !ok || len(entry.Content) == 0 ||
		entry.ModTime != info.ModTime().UnixNano() ||
		entry.CtimeNS == 0 || entry.CtimeNS != statCtimeNS(info) {
		return nil
	}
	window := inp.ToolResponse.Text()
	if window == "" {
		return nil
	}
	// Slice cached content to 1-based lines [StartLine, StartLine+NumLines).
	slice, ok := sliceLines(bytesconv.B2S(entry.Content), f.StartLine, f.NumLines)
	if !ok || slice != window {
		return nil
	}
	hashHex := cache.ShortHex(entry.Hash[:8])
	msg := fmt.Sprintf("[READ §unchanged-window:%s§ %s lines %d–%d — identical to cached read at turn %d. Full content available if needed.]",
		hashHex, ti.FilePath, f.StartLine, f.StartLine+f.NumLines-1, entry.Turn)
	if len(msg) >= len(window) { // never-worse
		return nil
	}
	// Best-effort analytics; mirrors handleRead's event shape.
	_ = analytics.Record(analytics.Event{
		TS:       time.Now().UnixNano(),
		SID:      inp.SessionID,
		Hook:     inp.ToolName,
		Action:   "unchanged-window",
		BytesIn:  len(window),
		BytesOut: len(msg),
	})
	return protocol.Replace(msg)
}

// sliceLines returns the byte-slice of s covering 1-based lines [start,
// start+n). It is a single pass over s that returns a subslice (no allocation,
// no []string materialization). ok=false when s has fewer lines than required,
// so a window running past EOF (or a start beyond EOF) passes through.
func sliceLines(s string, start, n int) (string, bool) {
	pos := 0
	for skipped := 1; skipped < start; skipped++ {
		nl := strings.IndexByte(s[pos:], '\n')
		if nl < 0 {
			return "", false
		}
		pos += nl + 1
	}
	end := pos
	for taken := range n {
		nl := strings.IndexByte(s[end:], '\n')
		if nl < 0 {
			if taken == n-1 && end < len(s) { // last line without a trailing '\n'
				return s[pos:], true
			}
			return "", false
		}
		end += nl + 1
	}
	return s[pos:end], true
}
