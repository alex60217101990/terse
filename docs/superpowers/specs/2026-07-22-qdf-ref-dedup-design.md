# qdf-hook — Universal §ref output dedup (design)

Date: 2026-07-22
Status: approved (brainstorm)

## Problem

Every tool result Claude receives is re-emitted in full even when byte-identical
to something already seen this session (re-run tests, repeated `git status`,
re-grep, unchanged logs). The existing bash cache dedups only whitelisted
read-only commands, per-command, single-session. Repetition elsewhere — any
command, any tool, across sessions — is uncompressed.

## Goal

Any tool output already seen (byte-identical) → a compact `§ref:HASH§` token
instead of the bytes. Content-addressed, cross-tool, cross-session-safe.
Generalizes and **replaces** the current bash cache.

## Design

### Content-addressed blob store

- `~/.qdf-hook/refs/<hash>.blob` holds one previously-emitted output.
- `hash = sha256(content)[:16]` as hex (32 chars). Collision probability
  negligible; wrong-content-on-collision is the only failure mode and is
  astronomically unlikely at 128 bits.
- Registry = the filesystem itself: existence is one `os.Stat` on the blob path.
  No separate index, nothing added to the hot SessionState.
- Blob payload is a qdf-serialized `refEntry{ Content string; TS int64 }`
  (qdf `OptBalanced` — the proven repetitive-payload default: ~38× faster
  decode than JSON at equal wire, and the payload here is exactly log/command
  text). `TS` (unix seconds) drives gc; no reliance on fs mtime.

### Cross-session safety (by construction)

A `§ref:HASH§` is emitted only after hashing the **current** output and finding
its blob. The blob content therefore equals the current bytes — staleness is
impossible. Cross-session reuse (same content cached by an earlier session) is a
free bonus, not a hazard.

### Dedup entry point

`cache.Dedup(content string, minSize int) (token string, deduped bool)`:
1. `len(content) < minSize` → `("", false)` (a `§ref` token costs ~13 tokens;
   not worth it below ~256 B).
2. `hash := sha256(content)[:16]`; hashing reads the input via `unsafe.String`
   (zero-copy).
3. `RefSeen(hash)` (one `os.Stat`) → emit `§ref:HASH§ (N bytes, seen before —
   qdf-hook expand HASH)`, `deduped=true`.
4. else `RefPut(hash, content)` (plain write, no tmp+rename — proven −63%) and
   `("", false)` (caller emits the content this first time).

Pure function over its inputs + the blob store; reused later by Grep / generic
squeezers.

### Wiring

- `HandleBash`: on the **passthrough** path (structural detectors did not
  compress), call `Dedup`. Byte-identical repeat of ANY command → `§ref`. This
  supersedes the whitelist + TTL `§bash-unchanged` path.
- **Remove `internal/cache/bashcache.go`** (and its `bashCache*` API,
  `readOnlyPrefixes`, `IsReadOnlyCommand`) — §ref is a strict superset. Keep
  `ShortHex` (used elsewhere) by moving it to a small shared file.
- New subcommand `qdf-hook expand <hash>`: reads the blob, qdf-decodes with
  `WithNoCopy` (zero-copy), prints `Content`. UX mirrors `sqz expand`.
- `gc`: extend to prune `refs/` by `TS` age (same utility/TTL approach as the
  removed bash cache). `--dry-run` lists.

### Components (isolation)

- `internal/cache/ref.go` — content-addressed blob I/O only: `RefPath`,
  `RefSeen`, `RefPut`, `RefGet`, `refEntry`. Independently testable.
- `cache.Dedup` — pure dedup decision over `ref.go`.
- `cmd/qdf-hook/expand.go` — the subcommand.
- `HandleBash` — one added call on the passthrough branch.

## Performance (measured, not assumed)

Apply the session's proven techniques: qdf `OptBalanced`, plain write (no
rename), `WithNoCopy` decode, zero-copy `unsafe.String` hashing.

Benchmarks (interleaved benchstat, n≥12, keep/revert):
- `BenchmarkDedup_Hit` — Stat + emit token. Target < 100 µs.
- `BenchmarkDedup_Miss` — Stat + qdf encode + plain write. Target < 250 µs
  (the plain-write disk floor).
- `BenchmarkExpand` — read + WithNoCopy decode. Target < 100 µs.

A `§ref` hit must never emit more bytes than passthrough (never-worse: only
emit `§ref` when `len(token) < len(content)`, guaranteed by `minSize`).

## Testing

- `ref_test.go`: RefPut→RefSeen→RefGet round-trip; qdf blob decodes; missing
  hash; corrupt blob → treated as absent.
- `Dedup`: first call writes + returns not-deduped; second identical call
  returns `§ref`; sub-minSize returns not-deduped.
- `expand` round-trips content written by `Dedup`.
- Bash integration: repeated identical unstructured output → `§ref` on 2nd.
- gc prunes aged ref blobs (dry-run lists, real removes).

## Out of scope (later specs)

- Block-level / partial-repeat dedup.
- Grep compressor; generic unstructured-output squeezer (RLE, head+tail, ANSI
  strip). These shrink *first* occurrences; §ref shrinks *repeats* — complementary.

## Risks

- Blob store growth → gc mandatory (in this spec).
- Removing bashcache changes the §bash-unchanged token wording; update any test
  asserting the old string.
