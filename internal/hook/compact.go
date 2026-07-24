package hook

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// HandlePreCompact runs before Claude Code performs context compaction.
// It records CompactedAt in the session state so subsequent reads know
// to serve full content (since compaction erased the context).
func HandlePreCompact(r io.Reader, w io.Writer) error {
	start := time.Now()

	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}

	state, err := cache.Load(inp.SessionID)
	if err != nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	state.Turn++
	state.CompactedAt = state.Turn
	if err := cache.Save(inp.SessionID, state); err != nil {
		fmt.Fprintf(os.Stderr, "qdf-hook: save state: %v\n", err)
	}

	_ = analytics.Record(analytics.Event{
		TS:     time.Now().UnixNano(),
		SID:    inp.SessionID,
		Hook:   "precompact",
		Action: "precompact",
		DurNS:  time.Since(start).Nanoseconds(),
	})

	// PreCompact hook doesn't replace tool output — just updates state.
	return protocol.EncodeOutput(w, protocol.Passthrough())
}

// HandlePostCompact runs after compaction completes.
// It injects a manifest of previously-read files into the fresh context,
// so Claude knows what's been read without re-reading.
func HandlePostCompact(r io.Reader, w io.Writer) error {
	start := time.Now()

	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}

	state, err := cache.Load(inp.SessionID)
	if err != nil || len(state.Files) == 0 {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	manifest := buildManifest(state)

	_ = analytics.Record(analytics.Event{
		TS:       time.Now().UnixNano(),
		SID:      inp.SessionID,
		Hook:     "postcompact",
		Action:   "postcompact",
		BytesOut: len(manifest),
		DurNS:    time.Since(start).Nanoseconds(),
	})

	return protocol.EncodeOutput(w, protocol.Replace(manifest))
}

// HandleSessionStart runs at session start (after compaction or fresh session).
// If there is an existing state file from a previous compacted session,
// it injects the file manifest. Otherwise it's a no-op.
func HandleSessionStart(r io.Reader, w io.Writer) error {
	start := time.Now()

	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}

	// Best-effort automatic cache prune, throttled to at most once/24h. This
	// runs on EVERY session start (before the no-manifest early return below),
	// so it fires reliably even for sessions that never compact. It is
	// blob-only (refs/ + last/) via SweepBlobs — it never touches session .qdf
	// state, so it cannot evict the very session this request is about to load.
	// Never blocks or fails the hook.
	if cache.ShouldRunGC(time.Now().Unix()) {
		cache.SweepBlobs(time.Now().Unix())
	}

	state, err := cache.Load(inp.SessionID)
	if err != nil || len(state.Files) == 0 || state.CompactedAt == 0 {
		// No prior session or no compaction — no manifest to inject.
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	manifest := buildManifest(state)

	_ = analytics.Record(analytics.Event{
		TS:     time.Now().UnixNano(),
		SID:    inp.SessionID,
		Hook:   "sessionstart",
		Action: "session-start",
		DurNS:  time.Since(start).Nanoseconds(),
	})

	return protocol.EncodeOutput(w, protocol.Replace(manifest))
}

// buildManifest creates a compact text manifest of all cached files.
func buildManifest(state *cache.SessionState) string {
	var sb strings.Builder
	sb.WriteString("[qdf-hook SESSION RESTORE — previously read files]\n")
	sb.WriteString("These files are tracked. Re-reads will return delta only.\n\n")

	// Sort for deterministic output.
	paths := make([]string, 0, len(state.Files))
	for p := range state.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, path := range paths {
		entry := state.Files[path]
		fmt.Fprintf(&sb, "  §ref:%s§  %s  (read at turn %d, %d bytes)\n",
			cache.ShortHex(entry.Hash[:8]), path, entry.Turn, len(entry.Content))
	}
	sb.WriteString("\n[Subsequent reads of these files will show delta or §unchanged§]\n")
	return sb.String()
}
