# Task 2 Report: in-RAM `StateStore` (MemStore) + lazy disk flush

## Status: DONE

## Commit
`<see below>` — `feat(hookcore): in-RAM sharded StateStore with lazy flush`

## Files Created
- `internal/hookcore/memstore.go` — `MemStore`, `NewMemStore`, `(*MemStore) StateStore`, `(*MemStore) FlushDirty`, and all `StateStore` interface methods.
- `internal/hookcore/memstore_test.go` — round-trip+flush test (from the brief, verbatim), a last-output round-trip+flush test, a missing-session-returns-empty test, and a `-race` concurrent test hammering two sessions plus the ref/last maps from 50 goroutines x 200 iterations each.

## Design

- **Sharded sessions**: `[16]*sessionShard`, each holding `map[string]*cache.SessionState` guarded by its own `sync.RWMutex`. Shard picked by `fnv.New32a()` hash of the session id mod 16.
- **Refs / last-output**: plain `map[string]string` each behind its own `sync.RWMutex` (no sharding needed — brief only asked for shard on sessions).
- **Dirty tracking**: a single `dirtyMu sync.Mutex` guarding three sets (`dirtySessions`, `dirtyRefs`, `dirtyLast` as `map[string]struct{}`), populated by every `Save*`/`*Put`. `FlushDirty` swaps each set out for a fresh empty map under the lock, then does the actual disk I/O (`cache.Save`, `cache.RefPut`, `cache.LastOutputPut`) without holding `dirtyMu` — so writes that land mid-flush go into the new set instead of blocking on I/O or racing the flush loop.
- **Constructor load**: `NewMemStore` walks `cache.StateDir()` (`*.qdf` files → `cache.Load(id)`), `cache.RefsDir()` (`*.blob` → `cache.RefGet(hash)`), and `cache.LastOutDir()` (`*.blob` → `cache.LastOutputGet(key)`), decoding each into the RAM maps. Missing directories or decode failures are skipped (best-effort), matching `cache.Load`'s own "corrupt → fresh state" behavior.
- **`LoadSession` aliasing**: returns the live `*cache.SessionState` pointer under the shard's read lock (not a deep copy) — documented in the type doc comment. This is safe because a given session id is only ever driven by one in-flight hook invocation; MemStore's actual concurrency guarantee is across *different* session ids, which the shard-per-id + per-shard-mutex design makes race-free.
- **No disk fallback in accessors**: once constructed, all reads (`LoadSession`, `RefGet`, `RefSeen`, `LastGet`) only ever touch RAM — disk is read once at construction and written only from `FlushDirty`.

## Test Results

```
GOWORK=off go test -race ./internal/hookcore/...
ok  github.com/alex60217101990/qdf-hook/internal/hookcore  1.9s

GOWORK=off go test -race ./...
ok  all packages (cmd/qdf-hook, internal/analytics, internal/cache, internal/detect, internal/hook, internal/hookcore, internal/protocol, internal/summary)
```

`go vet ./...` and `gofmt -l` on the new files: clean.

## TDD Trail
1. Wrote `memstore_test.go` (brief's round-trip+flush test plus 3 more) referencing `hookcore.NewMemStore` before it existed.
2. Confirmed compile failure: `undefined: hookcore.NewMemStore` (6 occurrences).
3. Implemented `internal/hookcore/memstore.go`.
4. `go test -race ./internal/hookcore/...` → pass.
5. `go test -race ./...` → whole suite pass.

## Concerns / Deviations
- None. Signatures match the brief exactly (`NewMemStore() *MemStore`, `(*MemStore) FlushDirty()`, `(*MemStore) StateStore() StateStore`). Task 1's `StateStore` interface and `diskStore` were not modified.
- Note: this file previously held an unrelated report (for `internal/cache` session state load/save, a different earlier task also numbered "2" in an older plan iteration) — that content has been superseded here since the current plan's Task 2 is the in-RAM MemStore described above.

---

# Follow-up: Critical concurrency bug fix + test gap (2026-07-23)

## Status: DONE

## Bug

`LoadSession` returned the **live** `*cache.SessionState` pointer stored in
the shard map. Real callers (`internal/hook/read.go`,
`internal/hook/write.go`) do load → mutate `state.Files` in place → save.
Two concurrent hook handlers for the *same* session id therefore shared one
`Files` map:

- concurrent writes to it → `fatal error: concurrent map writes`
- a concurrent `FlushDirty` (which ranges `Files` via `cache.Save` →
  `qdf.Marshal`) racing a handler's write → `fatal error: concurrent map
  iteration and map write` / data race under `-race`.

The type doc comment asserted a "single writer per session" invariant that
is not actually enforced anywhere and cannot be relied on.

## TDD Trail

1. **RED** — added `TestMemStore_ConcurrentLoadMutateSaveSameSession` to
   `internal/hookcore/memstore_test.go`: 50 goroutines all driving the same
   session id via `LoadSession → mutate .Files → SaveSession` in a loop,
   plus a concurrent goroutine looping `FlushDirty()`, all on one id. Ran it
   against the pre-fix code:

   ```
   GOWORK=off go test -race -run TestMemStore_ConcurrentLoadMutateSaveSameSession ./internal/hookcore/ -v
   ```

   Result: multiple `WARNING: DATA RACE` reports (map read/write races
   between goroutines) followed by `fatal error: concurrent map writes` and
   test binary abort — confirmed the test reproduces both the race and the
   crash before any fix.

2. **GREEN** — fixed `LoadSession` in `internal/hookcore/memstore.go` to
   return a **copy**: a new `*cache.SessionState` with copied scalar fields
   (`Turn`, `CompactedAt`) and a freshly allocated `Files` map populated from
   the stored session's entries, built entirely under the shard's `RLock`.
   `SaveSession` is unchanged — it remains the only path that mutates what's
   stored, via an atomic pointer swap under the shard's write lock — so a
   stored session is never mutated in place once stored, making
   `FlushDirty`'s unlocked read-then-marshal safe. Updated the `MemStore` and
   `LoadSession` doc comments to describe the copy-on-load contract and
   explicitly document last-writer-wins semantics instead of the
   unenforceable single-writer claim.

3. Also fixed the Minor doc inaccuracy in `NewMemStore`'s comment: a corrupt
   session file decodes via `cache.Load` to a fresh **empty** state with
   `err == nil` (not an error), so `loadFromDisk` was never actually
   "skipping" corrupt sessions — corrected the comment to describe what
   really happens (corrupt file → empty state loaded into RAM, same as
   `cache.Load`'s own contract).

## Test Results

```
$ GOWORK=off go test -race -run TestMemStore_ConcurrentLoadMutateSaveSameSession ./internal/hookcore/ -v
# (pre-fix) WARNING: DATA RACE (several) then:
fatal error: concurrent map writes
FAIL	github.com/alex60217101990/qdf-hook/internal/hookcore	0.872s

$ GOWORK=off go test -race ./internal/hookcore/ -v
=== RUN   TestDiskStore_RoundTrip
--- PASS: TestDiskStore_RoundTrip (0.00s)
=== RUN   TestMemStore_RoundTripAndFlush
--- PASS: TestMemStore_RoundTripAndFlush (0.00s)
=== RUN   TestMemStore_LastOutputRoundTripAndFlush
--- PASS: TestMemStore_LastOutputRoundTripAndFlush (0.00s)
=== RUN   TestMemStore_LoadSessionMissingReturnsEmpty
--- PASS: TestMemStore_LoadSessionMissingReturnsEmpty (0.00s)
=== RUN   TestMemStore_ConcurrentSessionsAndRefs
--- PASS: TestMemStore_ConcurrentSessionsAndRefs (0.14s)
=== RUN   TestMemStore_ConcurrentLoadMutateSaveSameSession
--- PASS: TestMemStore_ConcurrentLoadMutateSaveSameSession (0.11s)
PASS
ok  	github.com/alex60217101990/qdf-hook/internal/hookcore	2.414s

$ GOWORK=off go clean -testcache && GOWORK=off go test -race ./...
ok  	github.com/alex60217101990/qdf-hook/cmd/qdf-hook	2.264s
ok  	github.com/alex60217101990/qdf-hook/internal/analytics	2.807s
?   	github.com/alex60217101990/qdf-hook/internal/bytesconv	[no test files]
ok  	github.com/alex60217101990/qdf-hook/internal/cache	3.182s
ok  	github.com/alex60217101990/qdf-hook/internal/detect	1.404s
ok  	github.com/alex60217101990/qdf-hook/internal/hook	3.584s
ok  	github.com/alex60217101990/qdf-hook/internal/hookcore	4.059s
ok  	github.com/alex60217101990/qdf-hook/internal/protocol	2.446s
ok  	github.com/alex60217101990/qdf-hook/internal/summary	1.748s
```

## Files Changed
- `internal/hookcore/memstore.go` — `LoadSession` now copies under the shard
  `RLock` instead of returning the live pointer; doc comments on `MemStore`,
  `LoadSession`, and `NewMemStore` updated/corrected.
- `internal/hookcore/memstore_test.go` — added
  `TestMemStore_ConcurrentLoadMutateSaveSameSession`.

## Constraints Honored
- `StateStore` interface and Task-1 files untouched.
- `cache.Save`'s use of `OptSpeed` for session state left as-is (out of
  scope).
- All `go` commands run with `GOWORK=off`.
