package hook

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// HandleRead processes a PostToolUse hook call for the Read tool.
// It reads the hook JSON from r, applies delta/unchanged logic, and writes
// the hook output JSON to w.
func HandleRead(r io.Reader, w io.Writer) error {
	start := time.Now()

	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
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

	// Stat the file to capture its modification time.
	var modTime int64
	if info, err := os.Stat(ti.FilePath); err == nil {
		modTime = info.ModTime().UnixNano()
	}

	content := []byte(inp.ToolResponse.Content)

	// Binary content: always pass through, no diff.
	if cache.IsBinaryContent(content) {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Load session state.
	state, err := cache.Load(inp.SessionID)
	if err != nil {
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

	// Update cache entry with read tracking fields.
	updatedEntry := state.Files[ti.FilePath]
	updatedEntry.Hash = hash
	updatedEntry.Turn = state.Turn
	updatedEntry.Content = content
	updatedEntry.ReadCount++
	updatedEntry.LastReadAt = time.Now().Unix()
	updatedEntry.ModTime = modTime
	state.Files[ti.FilePath] = updatedEntry

	if err := cache.Save(inp.SessionID, state); err != nil {
		fmt.Fprintf(os.Stderr, "qdf-hook: save state: %v\n", err)
	}

	// Record analytics (best-effort — never block the hook). Passthrough emits
	// the full content, so its emitted size is len(content) (a neutral 0%
	// saving) — not 0, which would falsely read as 100% saved.
	bytesOut := len(content)
	if out.HookSpecificOutput != nil {
		bytesOut = len(out.HookSpecificOutput.UpdatedToolOutput)
	}
	_ = analytics.Record(analytics.Event{
		TS:       time.Now().UnixNano(),
		SID:      inp.SessionID,
		Hook:     "read",
		Action:   action,
		BytesIn:  len(content),
		BytesOut: bytesOut,
		DurNS:    time.Since(start).Nanoseconds(),
	})

	return protocol.EncodeOutput(w, out)
}


func serveUnchanged(path string, hash [32]byte, cachedAtTurn int) *protocol.HookOutput {
	hashHex := cache.ShortHex(hash[:8])
	msg := fmt.Sprintf("[READ §unchanged:%s§ %s — content identical to read at turn %d. Full content available if needed.]",
		hashHex, path, cachedAtTurn)
	return protocol.Replace(msg)
}

func serveDelta(path string, newHash [32]byte, oldContent, newContent []byte) *protocol.HookOutput {
	// Guard: very large diffs are O((N+M)²) — pass the full content through.
	if bytes.Count(oldContent, []byte("\n"))+bytes.Count(newContent, []byte("\n")) > 10000 {
		return protocol.Passthrough()
	}
	diff := cache.UnifiedDiff(oldContent, newContent, 3)
	hashHex := cache.ShortHex(newHash[:8])

	var sb strings.Builder
	fmt.Fprintf(&sb, "[READ §delta:%s§ %s — showing changes since last read]\n", hashHex, path)
	if diff == "" {
		// Shouldn't happen (hashes differ but diff is empty) — serve full.
		sb.Write(newContent)
	} else {
		sb.WriteString(diff)
	}
	return protocol.Replace(sb.String())
}
