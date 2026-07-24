package hook

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/bytesconv"
	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

const writePassthroughThreshold = 256

// HandleWrite compresses Write/Edit/MultiEdit tool responses.
// Claude just wrote this content — no need to see the full file again.
// Also caches the written content for delta tracking on next Read.
func HandleWrite(r io.Reader, w io.Writer) error {
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	return handleWrite(hookcore.NewDiskStore(), inp, w)
}

// handleWrite is the Write/Edit logic over an already-decoded input.
func handleWrite(store hookcore.StateStore, inp *protocol.HookInput, w io.Writer) error {
	start := time.Now()
	if inp.ToolResponse == nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Only the length is needed on the common (file-readable) path; the raw
	// bytes are materialized lazily below, in the fallback branch that is the
	// sole reader of them. Avoids copying the whole response string per call.
	contentLen := len(inp.ToolResponse.Content)
	if contentLen <= writePassthroughThreshold {
		_ = analytics.Record(analytics.Event{
			TS:       time.Now().UnixNano(),
			SID:      inp.SessionID,
			Hook:     "write",
			Action:   "passthrough",
			BytesIn:  contentLen,
			BytesOut: contentLen,
			DurNS:    time.Since(start).Nanoseconds(),
		})
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Parse file path from tool_input (Write and Edit both have file_path).
	var ti protocol.ReadInput // reuse: same {file_path} shape
	_ = json.Unmarshal(inp.ToolInput, &ti)

	// Cache the ACTUAL file bytes for delta tracking — not inp.ToolResponse
	// .Content, which is only the tool's confirmation/snippet (a diff for Edit,
	// a summary for Write). Caching the response would make the next Read hash
	// the real file, mismatch the cached snippet, and emit a bogus delta.
	var (
		hash      [32]byte
		hashHex   string
		lineCount int
	)
	fileBytes, ferr := os.ReadFile(ti.FilePath)
	if ferr == nil {
		hash = sha256.Sum256(fileBytes)
		hashHex = cache.ShortHex(hash[:8])
		lineCount = bytes.Count(fileBytes, []byte("\n"))
		if inp.SessionID != "" {
			state := store.LoadSession(inp.SessionID)
			if state != nil {
				state.Turn++
				var modTime int64
				if info, e := os.Stat(ti.FilePath); e == nil {
					modTime = info.ModTime().UnixNano()
				}
				state.Files[ti.FilePath] = cache.FileEntry{
					Hash:       hash,
					Turn:       state.Turn,
					Content:    fileBytes,
					ModTime:    modTime,
					LastReadAt: time.Now().Unix(),
					ReadCount:  1,
				}
				store.SaveSession(inp.SessionID, state)
			}
		}
	} else {
		// Can't read the file back — emit a marker from the response but do not
		// cache, so we never poison delta tracking with non-file bytes. S2B is
		// a read-only, call-scoped view (no copy); the bytes don't outlive it.
		content := bytesconv.S2B(inp.ToolResponse.Content)
		hash = sha256.Sum256(content)
		hashHex = cache.ShortHex(hash[:8])
		lineCount = bytes.Count(content, []byte("\n"))
	}

	compressed := fmt.Sprintf("[WRITE §ref:%s§ %s — %d lines, cached for delta tracking]",
		hashHex, ti.FilePath, lineCount)

	// Never-worse guard: if the marker isn't actually shorter than the
	// original response, pass the response through unchanged rather than grow
	// what Claude sees (a long file path can make the marker exceed a small
	// response). Mirrors read.go's guard and every dispatch.go branch.
	action, bytesOut := "compressed", len(compressed)
	replace := len(compressed) < contentLen
	if !replace {
		action, bytesOut = "passthrough", contentLen
	}

	_ = analytics.Record(analytics.Event{
		TS:       time.Now().UnixNano(),
		SID:      inp.SessionID,
		Hook:     "write",
		Action:   action,
		BytesIn:  contentLen,
		BytesOut: bytesOut,
		DurNS:    time.Since(start).Nanoseconds(),
	})

	if !replace {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}
	return protocol.EncodeOutput(w, protocol.Replace(compressed))
}
