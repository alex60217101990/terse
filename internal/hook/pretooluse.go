package hook

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alex60217101990/terse/internal/analytics"
	"github.com/alex60217101990/terse/internal/bytesconv"
	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/protocol"
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
	if inp.ToolName == "Bash" {
		return handleBashPreToolUse(inp, w)
	}

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

	state := store.LoadSession(ContextKey(inp))
	if state == nil {
		return protocol.EncodePre(w, "allow", "")
	}

	entry, seen := state.Files[ti.FilePath]
	action := "pretool-allow"
	decision := "allow"
	reason := ""
	var bytesIn, bytesOut int

	// Deny only when mtime, size AND ctime all match the cached copy. mtime
	// alone is not sufficient: cp -p, rsync --times, os.Chtimes, tar extraction
	// and coarse-resolution filesystems (NFS, some bind mounts) can leave mtime
	// unchanged across a content change. Requiring the size to match as well is
	// free (already stat'd) and closes most cases, but a same-size content
	// change combined with a forged/rewound mtime (cp -p, touch -r) still slips
	// through mtime+size alone. ctime advances on every content or metadata
	// change and cannot be moved backward from userspace, so it catches that
	// residual window. When ctime is unavailable (0 on either side — Windows, or
	// a pre-CtimeNS cache entry) the deny is NOT taken (see below): a deny can't
	// be recovered, so we never rest it on mtime+size alone.
	sizeMatch := info.Size() == int64(len(entry.Content))
	curCtime := statCtimeNS(info)
	// A deny asserts "unchanged, don't re-read" with no content in hand, so —
	// unlike every other path — it cannot be downgraded to a safe passthrough if
	// wrong. Require a real ctime match: if ctime is unavailable on either side
	// (a pre-CtimeNS cache entry, or a platform without it), do NOT deny — fall
	// through to allow, letting the PostToolUse handler make a content-based,
	// never-worse decision. mtime+size alone can be forged (cp -p, touch -r)
	// across a same-size content change; that residual would serve stale content
	// as authoritative. One extra read of a stale-cache file is strictly safer.
	ctimeMatch := entry.CtimeNS != 0 && curCtime != 0 && entry.CtimeNS == curCtime
	if seen && state.SeenAfterCompact(ti.FilePath) && entry.ModTime == info.ModTime().UnixNano() && sizeMatch && ctimeMatch {
		// mtime+size+ctime unchanged — safe to deny without reading.
		// Register the cached content so the deny is RECOVERABLE. This is the
		// only lossy path in the tool that carried no expand footer, which is
		// what turns a wrong deny into silent context loss instead of one extra
		// round trip (invariant 2: never-lossy-without-recovery).
		refHash := refTokenFor(store, bytesconv.B2S(entry.Content))
		hashHex := cache.ShortHex(entry.Hash[:8])
		var b strings.Builder
		b.Grow(len(hashHex) + len(refHash) + len(ti.FilePath) + 100)
		b.WriteString("§unchanged:")
		b.WriteString(hashHex)
		b.WriteString("§ ")
		b.WriteString(ti.FilePath)
		b.WriteString(" — mtime+size+ctime unchanged, cached at turn ")
		b.WriteString(strconv.Itoa(entry.Turn))
		b.WriteString(". If you do not have this content, run: qdf-hook expand ")
		b.WriteString(refHash)
		reason = b.String()
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
