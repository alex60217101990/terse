package hook

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

const writePassthroughThreshold = 256

// HandleWrite compresses Write/Edit/MultiEdit tool responses.
// Claude just wrote this content — no need to see the full file again.
// Also caches the written content for delta tracking on next Read.
func HandleWrite(r io.Reader, w io.Writer) error {
	start := time.Now()
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	if inp.ToolResponse == nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	content := []byte(inp.ToolResponse.Content)
	if len(content) <= writePassthroughThreshold {
		_ = analytics.Record(analytics.Event{
			TS:      time.Now().UnixNano(),
			SID:     inp.SessionID,
			Hook:    "write",
			Action:  "passthrough",
			BytesIn: len(content),
			BytesOut: len(content),
			DurNS:   time.Since(start).Nanoseconds(),
		})
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Parse file path from tool_input (Write and Edit both have file_path).
	var ti protocol.ReadInput // reuse: same {file_path} shape
	_ = json.Unmarshal(inp.ToolInput, &ti)

	hash := sha256.Sum256(content)
	hashHex := fmt.Sprintf("%x", hash[:8])
	lineCount := bytes.Count(content, []byte("\n"))

	compressed := fmt.Sprintf("[WRITE §ref:%s§ %s — %d lines written, cached for delta tracking]",
		hashHex, ti.FilePath, lineCount)

	// Cache written content for delta tracking on next Read.
	if ti.FilePath != "" && inp.SessionID != "" {
		state, serr := cache.Load(inp.SessionID)
		if serr == nil {
			state.Turn++
			state.Files[ti.FilePath] = cache.FileEntry{
				Hash:       hash,
				Turn:       state.Turn,
				Content:    content,
				LastReadAt: time.Now().Unix(),
				ReadCount:  1,
			}
			_ = cache.Save(inp.SessionID, state)
		}
	}

	_ = analytics.Record(analytics.Event{
		TS:       time.Now().UnixNano(),
		SID:      inp.SessionID,
		Hook:     "write",
		Action:   "compressed",
		BytesIn:  len(content),
		BytesOut: len(compressed),
		DurNS:    time.Since(start).Nanoseconds(),
	})

	return protocol.EncodeOutput(w, protocol.Replace(compressed))
}
