// Package cache is qdf-hook's persistence layer.
//
// It owns everything under ~/.qdf-hook: per-session read state (which files the
// model has seen, their hashes and mtimes), a content-addressed §ref blob store
// for deduplicating byte-identical repeated output, Myers unified diffing for
// changed files, and utility-based eviction.
//
// Every store is a rebuildable cache. Reads decode defensively: a corrupt or
// truncated file degrades to a fresh, empty state rather than an error, so a
// crash or a racing writer can only ever cause a cache miss, never wrong
// content. Writes are plain (no tmp+rename) for the same reason.
//
// Serialization uses github.com/alex60217101990/qdf in its balanced mode
// (~38x faster decode than encoding/json at equal wire size), decoded zero-copy
// where the backing buffer safely outlives the values.
//
// Key entry points: [Load]/[Save] (session state), [Dedup]/[RefGet]/[expand]
// (the §ref store), [UnifiedDiff] (delta), [RunGC] (eviction), and [ShortHex]
// (stack-allocated hex encoding used across the hooks).
package cache
