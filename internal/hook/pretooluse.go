package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// HandlePreToolUse intercepts Read tool calls before execution.
// If the file's mtime matches the cached mtime and the file was seen after
// the last compaction, it denies the read with an §unchanged§ marker —
// the file is not read at all, saving I/O and output tokens.
func HandlePreToolUse(r io.Reader, w io.Writer) error {
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return protocol.EncodePre(w, "allow", "")
	}
	return handlePreToolUse(hookcore.NewDiskStore(), inp, w)
}

// handlePreToolUse is the PreToolUse logic over a decoded input and a state
// store, so both the CLI (disk store) and the daemon (in-RAM store, via
// routeInput) share it. Routing through the daemon lets a repeated Read hit
// the in-memory session instead of a fresh process + disk decode.
func handlePreToolUse(store hookcore.StateStore, inp *protocol.HookInput, w io.Writer) error {
	start := time.Now()

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

	state := store.LoadSession(inp.SessionID)
	if state == nil {
		return protocol.EncodePre(w, "allow", "")
	}

	entry, seen := state.Files[ti.FilePath]
	action := "pretool-allow"
	decision := "allow"
	reason := ""
	var bytesIn, bytesOut int

	// Deny only when BOTH mtime and size match the cached copy. mtime alone is
	// not sufficient: cp -p, rsync --times, os.Chtimes, tar extraction and
	// coarse-resolution filesystems (NFS, some bind mounts) can leave mtime
	// unchanged across a content change. Requiring the size to match as well is
	// free (already stat'd) and closes the common cases. Residual risk — an
	// edit that preserves both mtime and size — is rare and only costs a missed
	// compression, never wrong content beyond that window.
	sizeMatch := info.Size() == int64(len(entry.Content))
	if seen && state.SeenAfterCompact(ti.FilePath) && entry.ModTime == info.ModTime().UnixNano() && sizeMatch {
		// mtime + size unchanged — safe to deny without reading.
		hashHex := cache.ShortHex(entry.Hash[:8])
		reason = fmt.Sprintf(
			"§unchanged:%s§ %s — mtime+size unchanged, cached at turn %d. No re-read needed.",
			hashHex, ti.FilePath, entry.Turn,
		)
		decision = "deny"
		action = "pretool-unchanged"
		// Credit the saving: the whole cached file would have been re-read and
		// re-ingested; instead Claude receives only the short deny reason.
		bytesIn = len(entry.Content)
		bytesOut = len(reason)

		// Deliberately do NOT Save here. This is the hottest path (every
		// repeated read of an unchanged file) and the only mutation would be a
		// ReadCount/LastReadAt bump for gc's utility score — not worth an
		// atomic full-state rewrite (~160µs of syscalls) on every re-read. The
		// cost is only slightly staler eviction scoring: a file that is only
		// ever re-read (never re-served by the PostToolUse handler) may be
		// evicted sooner, which just causes a one-off re-cache, never wrong
		// content.
	}

	_ = analytics.Record(analytics.Event{
		TS:       time.Now().UnixNano(),
		SID:      inp.SessionID,
		Hook:     "pretooluse",
		Action:   action,
		BytesIn:  bytesIn,
		BytesOut: bytesOut,
		DurNS:    time.Since(start).Nanoseconds(),
	})

	return protocol.EncodePre(w, decision, reason)
}
