package hook

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// HandleRead processes a PostToolUse hook call for the Read tool.
// It reads the hook JSON from r, applies delta/unchanged logic, and writes
// the hook output JSON to w.
func HandleRead(r io.Reader, w io.Writer) error {
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

	switch {
	case !seen || !state.SeenAfterCompact(ti.FilePath):
		// First read (or first read after compaction) — serve full content.
		out = serveFullContent(ti.FilePath, hash, content, inp.ToolResponse.Content)

	case entry.Hash == hash:
		// Same content — serve §unchanged§ marker.
		out = serveUnchanged(ti.FilePath, hash, entry.Turn)

	default:
		// Content changed — serve §delta§ with unified diff.
		out = serveDelta(ti.FilePath, hash, entry.Content, content)
	}

	// Update cache entry.
	state.Files[ti.FilePath] = cache.FileEntry{
		Hash:    hash,
		Turn:    state.Turn,
		Content: content,
	}

	_ = cache.Save(inp.SessionID, state)
	return protocol.EncodeOutput(w, out)
}

func serveFullContent(path string, hash [32]byte, _ []byte, original string) *protocol.HookOutput {
	// First read: return original content with a cache-registration header.
	hashHex := fmt.Sprintf("%x", hash[:8]) // first 8 bytes = 16 hex chars
	header := fmt.Sprintf("[READ §ref:%s§ %s — CACHED for delta tracking]\n", hashHex, path)
	return protocol.Replace(header + original)
}

func serveUnchanged(path string, hash [32]byte, cachedAtTurn int) *protocol.HookOutput {
	hashHex := fmt.Sprintf("%x", hash[:8])
	msg := fmt.Sprintf("[READ §unchanged:%s§ %s — content identical to read at turn %d. Full content available if needed.]",
		hashHex, path, cachedAtTurn)
	return protocol.Replace(msg)
}

func serveDelta(path string, newHash [32]byte, oldContent, newContent []byte) *protocol.HookOutput {
	diff := cache.UnifiedDiff(oldContent, newContent, 3)
	hashHex := fmt.Sprintf("%x", newHash[:8])

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
