# qdf-hook — resident daemon (qdf-hookd) design

Date: 2026-07-23
Status: approved (brainstorm)

## Problem

Per hook event Claude Code exec's the Go binary (~1–6 ms spawn + runtime init)
and every stateful hook does disk syscalls (Load/Save). These two dominate — the
actual compression work is tens of µs. The only way below the per-process floor
is to stop spawning a Go process per event and stop touching disk per event.

## Goal

A long-lived `qdf-hookd` that holds all state in RAM and answers hook requests
over a unix socket, fronted by a near-zero-startup client. Eliminate per-event
Go spawn (non-Go client) and per-event disk (in-RAM state, lazy flush), while
degrading gracefully to today's per-process path whenever the daemon is absent.

## Architecture

### Client transport — hybrid (approved)

Hook command: `nc -U ~/.qdf-hook/d.sock 2>/dev/null || <exe> post`

- Daemon up + `nc` present → `nc` pipes stdin→socket, relays reply→stdout in
  **sub-ms, no Go runtime, no disk**.
- Daemon down / socket missing / no `nc` → `nc` fails at **connect** (stdin
  untouched) → `||` runs the Go `post` fallback (today's local processing).
- Connection-per-request, **EOF-delimited** (no framing): client half-closes
  after piping stdin; daemon reads to EOF, processes, writes reply, closes.

Requires the hook `command` to run under a shell (Claude Code's default). `init`
emits this hybrid command; detection/idempotency updated to recognize it.

### Daemon

- Listens on `~/.qdf-hook/d.sock` (unix stream). Accepts connections
  concurrently (goroutine per conn).
- State lives as **live Go structs in RAM** — no per-request decode. Same
  pipeline as `Dispatch` (Read/Write handlers + generic pipeline), operating on
  the in-RAM state instead of Load/Save.
- **Persistence:** load from disk on startup; flush dirty sessions on a ~5 s
  timer and on SIGTERM / idle-exit. qdf `OptBalanced`. Crash loses a few seconds
  of a rebuildable cache — acceptable.
- **Lifecycle:** SessionStart runs `qdf-hook daemon --ensure` (starts detached
  if the socket is dead; restarts if the running daemon's version is older via a
  version handshake). Idle-exit after N minutes with no requests.
- **Concurrency:** per-session state guarded by a sharded mutex (session id →
  shard); refs/lastout stores guarded independently. Readers don't block across
  sessions.

### Fallback correctness

Every failure path (no daemon, no nc, crash mid-request, decode error) resolves
to either the Go local pipeline or a plain passthrough. The hook must never
block or emit garbage.

## Server-side performance (long-lived — these techniques now pay off)

Unlike the one-shot CLI (where pools/interning are useless — the process exits
before reuse), the daemon amortizes them across thousands of requests. Apply all:

- **Buffer reuse / pools:** `sync.Pool` (or per-conn) read buffers, response
  builders, and diff scratch — reused across requests, cutting steady-state GC.
- **String interning:** intern repeated strings the daemon sees across requests
  — file paths, tool names, session ids, commands — so the in-RAM state stores
  one copy each (flat open-addressed table; see the go-string-interning-tuning
  approach — plain map[string]string first, measure before anything fancier).
- **Per-request arena:** a bump arena for transient request scratch (parsed
  bits, diff output), `Reset` after each response. **Prototype + benchstat vs a
  sync.Pool of buffers** — arena is not a guaranteed win (go-arena-alloc-pitfalls);
  gate on the realistic request mix, never keep it on faith.
- **Stack / small fixed buffers:** hex encoding, key hashing (ShortHex stays
  stack; hashing via zero-copy views).
- **Zero-copy strings:** `internal/bytesconv` S2B/B2S on the request path
  (read-only, lifetime-bound) — no `[]byte(string)` copies.
- **Generics, no boxing:** typed request/response structs and any shared
  container (e.g. a generic sharded store) use type parameters, not `interface{}`
  — no boxing on the hot path.
- **qdf:** only for the disk snapshot/load (OptBalanced), not per request (state
  is already live in RAM). `qdf.Arena`/`WithArena`/`StreamDecoder` do NOT fit an
  nc/JSON wire and are not used per request.

## Measure-first (hard requirement)

- Baseline the current per-process event cost (spawn + Load/Save) vs the daemon
  path (nc round-trip + in-RAM handle) end-to-end. Keep the daemon only if it
  clears the noise by a wide margin.
- Every server-perf lever (pool, intern, arena) is `benchstat`-gated (n≥12) on a
  realistic request stream; revert losers with a recorded reason.
- Parity: the daemon must produce byte-identical hook output to the local
  pipeline for the same input (shared handler code + a differential test).

## Phasing (for the plan)

1. Extract the pipeline behind an interface that works on an in-memory state
   provider (so daemon and CLI share exactly one implementation).
2. Daemon core: socket listen, conn handling, in-RAM state, lazy flush.
3. Lifecycle: `daemon --ensure`, version handshake, idle-exit, SessionStart wire.
4. Hybrid client command + `init` migration + detection update.
5. Server-perf pass (pools/intern/arena) — each measured & gated.
6. Differential parity tests + end-to-end latency benchmark vs per-process.

## Out of scope

Cross-machine/remote daemon; Windows named pipes (unix socket only for now).
