// Package hookcore defines the storage abstraction behind the compression
// pipeline (internal/hook) so it can be driven by either the on-disk cache
// (CLI, one-shot per invocation) or an in-memory store (a future long-lived
// daemon). It depends on internal/cache but never on internal/hook, so hook
// can safely import hookcore without an import cycle.
package hookcore

import "github.com/alex60217101990/terse/internal/cache"

// StateStore is the storage surface the compression pipeline needs: session
// state (per-file read/write tracking) plus the two content-addressed caches
// (dedup refs and per-tool-call last-output, used for delta encoding).
type StateStore interface {
	// LoadSession returns the session state for id (an empty state if none
	// exists yet).
	LoadSession(id string) *cache.SessionState
	// SaveSession persists the session state for id.
	SaveSession(id string, s *cache.SessionState)

	// RefSeen reports whether content addressed by hash was already stored.
	RefSeen(hash string) bool
	// RefPut stores content under hash.
	RefPut(hash, content string)
	// RefGet returns the content stored under hash.
	RefGet(hash string) (string, bool)
	// RefHit records a dedup hit against a stored ref (usage bump for eviction).
	RefHit(hash string)

	// LastGet returns the previous tool output stored under key.
	LastGet(key string) (string, bool)
	// LastPut stores the current tool output under key.
	LastPut(key, content string)
}
