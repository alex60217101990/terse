package hook

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/alex60217101990/terse/internal/analytics"
	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/protocol"
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
	paths := slices.Sorted(maps.Keys(state.Files))

	for _, path := range paths {
		entry := state.Files[path]
		fmt.Fprintf(&sb, "  §ref:%s§  %s  (read at turn %d, %d bytes)\n",
			cache.ShortHex(entry.Hash[:8]), path, entry.Turn, len(entry.Content))
	}
	sb.WriteString("\n[Subsequent reads of these files will show delta or §unchanged§]\n")
	return sb.String()
}
