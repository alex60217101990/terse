package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// HandlePreToolUse intercepts Read tool calls before execution.
// If the file's mtime matches the cached mtime and the file was seen after
// the last compaction, it denies the read with an §unchanged§ marker —
// the file is not read at all, saving I/O and output tokens.
func HandlePreToolUse(r io.Reader, w io.Writer) error {
	start := time.Now()

	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return protocol.EncodePre(w, "allow", "")
	}

	var ti protocol.ReadInput
	if err := json.Unmarshal(inp.ToolInput, &ti); err != nil || ti.FilePath == "" {
		return protocol.EncodePre(w, "allow", "")
	}

	// Fast path: stat only (no content read).
	info, err := os.Stat(ti.FilePath)
	if err != nil {
		// File doesn't exist or unreadable — let the tool handle it.
		return protocol.EncodePre(w, "allow", "")
	}

	state, err := cache.Load(inp.SessionID)
	if err != nil {
		return protocol.EncodePre(w, "allow", "")
	}

	entry, seen := state.Files[ti.FilePath]
	action := "pretool-allow"
	decision := "allow"
	reason := ""

	if seen && state.SeenAfterCompact(ti.FilePath) && entry.ModTime == info.ModTime().UnixNano() {
		// mtime unchanged — safe to deny without reading.
		hashHex := fmt.Sprintf("%x", entry.Hash[:8])
		reason = fmt.Sprintf(
			"§unchanged:%s§ %s — mtime unchanged, cached at turn %d. No re-read needed.",
			hashHex, ti.FilePath, entry.Turn,
		)
		decision = "deny"
		action = "pretool-unchanged"

		// Update usage stats.
		entry.ReadCount++
		entry.LastReadAt = time.Now().Unix()
		state.Files[ti.FilePath] = entry
		if err := cache.Save(inp.SessionID, state); err != nil {
			fmt.Fprintf(os.Stderr, "qdf-hook: save state: %v\n", err)
		}
	}

	_ = analytics.Record(analytics.Event{
		TS:     time.Now().UnixNano(),
		SID:    inp.SessionID,
		Hook:   "pretooluse",
		Action: action,
		DurNS:  time.Since(start).Nanoseconds(),
	})

	return protocol.EncodePre(w, decision, reason)
}
