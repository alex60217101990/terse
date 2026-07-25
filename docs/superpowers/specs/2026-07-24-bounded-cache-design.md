# Bounded Cache Growth Control — Design

**Date:** 2026-07-24
**Status:** approved (brainstorming), pending implementation plan

## Problem

qdf-hook's on-disk cache grows without bound and the daemon leaks memory:

1. **`refs/` and `last/` are unbounded.** Every unique tool output ≥256 B
   becomes a content-addressed `.blob` under `refs/` (dedup) or a rerun-delta
   base under `last/`. Neither has a TTL, size cap, or prune. `qdf-hook gc`
   (`RunGC`) evicts only whole *session* files and is manual-only. Over weeks
   this is unbounded disk growth.
2. **The daemon never garbage-collects.** `cmd/qdf-hook/main.go`'s `init()`
   runs `runtime.GOMAXPROCS(1)` + `debug.SetGCPercent(-1)` for *every*
   subcommand — correct for the one-shot CLI, catastrophic for the long-lived
   `daemon --serve`: GC is disabled, so every request's transient allocations
   accumulate for the daemon's whole life, and a single P serializes concurrent
   connections.
3. **The daemon pins whole read buffers in RAM.** `MemStore` holds every ref's
   `Content` in a `map[string]string`, loaded at startup and decoded with
   `qdf.WithNoCopy()`, so decoded strings alias whole `os.ReadFile` buffers that
   stay live for the daemon's lifetime.

Session `Files` are already bounded (auto `Evict(s,200)` in `cache.Save`); this
design leaves that untouched and bounds the rest.

## Goals

- Bound `refs/` + `last/` on disk by **size + age + usage** (3-criteria
  utility eviction).
- Run gc/prune **automatically**, not only on manual invocation.
- Fix the daemon's disabled GC / single-P runtime and stop holding ref/last
  content in RAM.
- Preserve correctness: bit-identical hook output (the Task-6 parity test
  guards it); a missed/evicted entry only costs a re-cache, never wrong content.

## Non-goals

- Changing session `Files` eviction (already bounded).
- Changing the hook↔Claude JSON protocol.
- No qdf column-projection or arena: the dedup hot path needs only a hash's
  existence (the `§ref` token's byte count comes from the *current* output the
  caller already holds, not from stored metadata), so the daemon never decodes
  a blob on the hot path — there is nothing to project and no bulk decode to
  arena. (Evaluated per request; both are genuinely useful for wide multi-field
  records, which a single-`Content` blob is not — YAGNI here.)

## Architecture

Six components.

### 1. Disk layout (blobs unchanged)

`refEntry`/`lastEntry` blobs keep their current schema (`Content`, `TS`). No
migration. A blob's **creation time** is its file mtime (blobs are written once
and never rewritten) and its **size** is the dir-entry size — both free from a
directory scan, no decode.

Usage counters live in a separate sidecar, not in the blobs (§3), so a dedup
hit never rewrites a blob.

### 2. Daemon RAM: seen-set, not content

`MemStore` stops holding ref/last `Content` in RAM. Instead:

- `refs` becomes a `map[string]struct{}` **seen-set** of hashes, built from one
  `refs/` directory scan at startup and updated on `RefPut`. `RefSeen(hash)` is
  a map lookup; no `Content`, so no `WithNoCopy` buffer pinning.
- `RefPut(hash, content)` writes the blob to disk (lazily, as today) and adds
  the hash to the set.
- `RefGet(hash)` — used only by `qdf-hook expand` — reads the blob from disk on
  demand.
- `last/` is handled the same way: a `map[key]struct{}` of known keys; the
  previous `Output` is read from disk only when a rerun-delta actually needs it.

Daemon RAM for the cache is now O(number of hashes) map keys, not O(total blob
bytes).

### 3. Usage sidecar

A single sidecar index tracks usage for eviction: `hash → {Hits uint32,
LastUsed int64}` (and the same for `last/` keys), serialized with qdf
(`OptBalanced`, matching the blob writers).

- **Daemon:** bumps `Hits`/`LastUsed` in an in-RAM copy of the sidecar on each
  dedup hit; flushed lazily on the existing 5 s ticker. Free on the read path.
- **CLI fallback:** updates the sidecar in batch at end of invocation (never on
  a per-blob basis).
- Missing/corrupt sidecar → treated as empty (all entries `Hits=0`,
  `LastUsed=0`, i.e. coldest); worst case is coarser eviction, never wrong
  content. Cross-process daemon+CLI updates are last-writer-wins on advisory
  counters — acceptable.

### 4. Utility eviction (3 criteria)

Extends the exponential-decay utility model already in
`internal/cache/evict.go`.

- **Score** per entry = `Hits × decay(now − LastUsed)` — the same recency ×
  frequency curve `RunGC` uses for sessions. `Hits=0` (never re-hit) scores ~0
  → evicted first.
- **Hard size cap** (trigger): total bytes of `refs/` + `last/` from the dir
  scan. When over the cap, evict lowest-score entries until at **80 %** of the
  cap (matches `Evict`'s existing 80 % target).
- **TTL floor** (age): entries whose `LastUsed` (or mtime, if absent from the
  sidecar) is older than the TTL are dropped unconditionally, before the score
  pass.

Eviction deletes the disk blob and removes the hash from the seen-set + sidecar.

### 5. Automatic gc/prune

`RunGC` is extended to also sweep `refs/` + `last/` (currently sessions only),
via a dir scan joined with the sidecar. It runs:

- **SessionStart, throttled:** at most once per 24 h, gated by a timestamp
  marker (`~/.qdf-hook/.gc-stamp`); off the hot path, covers CLI-only users.
  Config comes from env on this path (§7).
- **Daemon ticker:** a periodic sweep folded into the daemon's flush loop
  (e.g. every N flushes), since the daemon is long-lived and accumulates
  between SessionStarts.
- **Manual:** `qdf-hook gc` stays, now covering refs/last too.

### 6. Daemon runtime restoration

`daemon --serve` (in `cmd/qdf-hook/daemon.go`, before `daemon.Serve`) restores
the runtime the CLI `init()` detuned:

```go
runtime.GOMAXPROCS(runtime.NumCPU())
debug.SetGCPercent(100)
```

One-shot subcommands keep the `init()` tuning; only the daemon undoes it. This
is the single most important memory fix — without it, GC never runs in the
daemon and nothing else bounds its RAM.

### 7. Configuration

Cap and TTL are configurable, with defaults:

- `daemon --serve` and `qdf-hook gc` flags: `--cache-max-size` (bytes, default
  **128 MiB**) and `--cache-ttl` (duration, default **720h** = 30 days).
- The SessionStart-triggered auto-gc (a hook, not a flagged invocation) reads
  `QDF_CACHE_MAX_SIZE` / `QDF_CACHE_TTL`, falling back to the same defaults.

## Data flow

- **RefPut:** write blob (`Content`+`TS`) to disk → add hash to seen-set.
- **Dedup hit (`RefSeen` true):** bump sidecar `Hits`/`LastUsed` (RAM, lazy
  flush / CLI batch) → emit `§ref` token from the hash + the current output's
  length (no blob read).
- **expand:** `RefGet` reads the full blob from disk.
- **gc sweep:** dir-scan refs/last → TTL-floor drop → if over size cap, evict
  lowest-score (sidecar-joined) until 80 % → delete blobs + set/sidecar entries.

## Error handling

- Best-effort throughout (matches existing cache): a failed prune/flush is
  logged (daemon → `daemon.log`) and retried next cycle; it never fails a hook.
- Missing/corrupt sidecar → empty (coldest); coarser eviction, never wrong
  content.
- Config parse failure (bad flag/env) → fall back to defaults, log once.

## Testing

- Eviction unit tests: over-cap triggers eviction to 80 %; TTL floor drops old;
  score orders by `Hits × recency` (a hot recent entry survives a cold old one).
- Daemon seen-set: `RefSeen` is true after `RefPut` without holding Content;
  `RefGet` still returns the blob (disk read); a large blob does not grow the
  daemon's resident cache maps beyond the hash key.
- Daemon runtime restoration: after `daemon --serve` start, `GOMAXPROCS` =
  NumCPU and `GCPercent` = 100 (unit test of the restore step, or a hook that
  reports them).
- Auto-gc throttle: two SessionStarts within 24 h → one sweep; stamp respected.
- Config: flag/env override changes the effective cap/TTL; defaults when unset.
- Parity unchanged: the Task-6 daemon-vs-CLI byte-identical parity test still
  passes.

## Risks

- **Cross-process usage races:** daemon + CLI both updating the sidecar →
  last-writer-wins on advisory counters; acceptable (they only affect eviction
  ordering, never correctness).
- **gc dir-scan cost on huge caches:** a scan of thousands of blob files each
  sweep. Bounded because the cap keeps the file count in check, and the sweep
  runs at most every 24 h (SessionStart) or on the daemon's slow ticker, never
  on the hot path.
