# Architecture

`qdf-hook` is a single static binary invoked by Claude Code as a hook command,
once per tool event. Each invocation reads a JSON event on stdin, decides how to
compress the tool's output, and writes a hook response on stdout. There is no
daemon and no shared process — all coordination happens through small files
under `~/.qdf-hook/`.

## Data flow

```
Claude Code event ──stdin(JSON)──▶ qdf-hook <subcommand> ──stdout(JSON)──▶ Claude Code
                                          │
                                          ├── reads/writes ~/.qdf-hook/sessions/<id>.qdf
                                          ├── reads/writes ~/.qdf-hook/refs/<hash>.blob
                                          └── appends      ~/.qdf-hook/analytics.jsonl
```

## Packages

| Package | Responsibility |
| --- | --- |
| `cmd/qdf-hook` | CLI: one subcommand per hook; profiling flags; `stats`/`gc`/`expand`. |
| `internal/protocol` | Hook wire types; decode stdin / encode stdout. |
| `internal/hook` | One handler per tool (`HandleRead`, `HandleBash`, …). Orchestration only. |
| `internal/detect` | Pure output analysis: JSON columnar stats, `go test`/`git log`/bench summarizers, `SqueezeOutput`. No I/O. |
| `internal/summary` | Human-readable rendering of `detect` stats. |
| `internal/cache` | Persistence: session state, content-addressed `§ref` store, unified diff, eviction. |
| `internal/analytics` | Append-only event log + `stats` aggregation. |

The split keeps `detect` pure and table-testable, `cache` responsible for all
disk I/O, and `hook` a thin orchestration layer.

## Core mechanisms

### 1. PreToolUse read interception (the biggest lever)

Re-reading a file is the #1 token sink. Before a `Read` executes, the
`pretooluse` handler stats the file and compares mtime **and** size against the
cached entry. If both match (and the file was seen after the last compaction),
it returns `permissionDecision: "deny"` with a `§unchanged§` reason — **the file
is never read**, so the tokens are never spent.

mtime alone is insufficient (`cp -p`, `rsync --times`, coarse NFS clocks can
preserve it across a change), so the size check — free, already in the stat —
closes the common holes. The residual case (same mtime *and* size, different
content) only costs a missed compression, never wrong content.

This path deliberately does **not** persist anything: bumping a usage counter is
not worth a full state rewrite (~160 µs of syscalls) on the hottest path.

### 2. Content-addressed §ref dedup

Any tool output byte-identical to something already emitted collapses to a
`§ref:HASH§` token. `HASH = sha256(output)[:16]`; the blob lives at
`~/.qdf-hook/refs/<hash>.blob`. The filesystem *is* the registry — one `stat`
tells us whether we've seen this output.

Safe by construction: a `§ref` is emitted only after hashing the **current**
output and finding its blob, so the blob always equals the current bytes. No
staleness, works across sessions. `qdf-hook expand <hash>` reconstructs the full
content on demand.

This generalizes and replaces the older read-only bash cache (which was
whitelist-gated, single-session, and byte-identical only).

### 3. Delta tracking

The `read` handler caches each file's content. On a re-read where the content
changed, it computes a Myers unified diff (`internal/cache/delta.go`) and returns
`[§delta:HASH§]` + the diff instead of the whole file. A hard guard falls back
to full content when the diff would be pathologically large (O((N+M)²) Myers).

### 4. Structural summarizers

`internal/detect` recognizes and compresses common shapes: JSON arrays →
columnar schema + per-column stats; `go test -v` → pass/fail counts + only the
failures; `git log` → compact commit table; `go test -bench` → aligned table.
Each is gated so it only fires when it at least halves the output.

### 5. Generic squeeze

Output that no structural detector matched, and that isn't a `§ref` repeat, goes
through `SqueezeOutput`: ANSI/VT escape stripping + run-length collapse of
identical consecutive lines (`line ⨯N`). Self-describing, gated at ≥10 % win.

## Persistence choices

- **Format:** [`qdf`](https://github.com/alex60217101990/qdf), not JSON. Session
  state and `§ref` blobs use `OptBalanced` — ~38× faster decode than JSON at
  equal wire size, which matters because Load runs on every hook.
- **Zero-copy decode:** `qdf.WithNoCopy()` aliases decoded strings/bytes into the
  read buffer (safe: the buffer outlives the value and is never mutated),
  cutting decode allocations ~74 %.
- **Writes:** plain `os.WriteFile`, no tmp+rename. Every state file is a
  rebuildable cache; a torn write from a crash or a concurrent hook simply fails
  to decode on the next Load, which returns a fresh empty state. The atomic
  rename bought nothing and cost a syscall on every hook.
- **No `fsync`:** cross-process visibility rides the page cache (correct, since
  the next hook is a new process reading the same file); durability is not
  required for a cache.

## Eviction (`gc`)

`gc` prunes two stores: session files by a utility score (read frequency ×
recency with exponential decay, `internal/cache/evict.go`), and `§ref` blobs by
age (`QDF_REF_TTL_HOURS`, default 168 h). `--dry-run` reports without deleting.

## Why it's safe

- Hooks never mutate files or change command behavior — they run after execution
  (or only deny a redundant read).
- Every cache is content-addressed or rebuildable; the failure mode is always a
  cache miss (a correct, fresh read), never wrong content or a crash.
- Every compressor is never-worse: it emits the compact form only when it is
  strictly smaller than the original.
