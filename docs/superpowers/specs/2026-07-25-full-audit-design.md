# Full audit — bugs + measure-first perf

Branch: `perf/full-audit`. Focus (user decision 2026-07-25): **real bugs +
robustness first**; perf changes kept **only** when benchstat proves a win
(never-worse). Every agent finding is verified by the human-in-the-loop
(Claude, main thread) — agents are advisory, never trusted on their word.

## Protocol

1. **Fan-out (read-only agents, parallel).** Each agent audits one cross-cutting
   lens across the relevant packages and returns structured findings only —
   `{file:line, claim, severity, concrete failure scenario, repro steps}`. No
   agent edits code.
2. **Verify (main thread, NOT agents).** For each finding: reproduce via a
   failing test or a direct code trace. Classify CONFIRMED vs false-positive.
   Only CONFIRMED findings proceed.
3. **Fix (TDD).** Failing regression test → fix → green. Whole suite `-race`
   clean after every fix.
4. **Perf (separate, measure-first).** Agents only *propose* levers. Main thread
   snapshots a benchstat baseline, prototypes ONE lever, runs interleaved
   benchstat (n≥10), keeps only if it clears noise; reverts otherwise. No
   speculative optimization. Opt-in tiers only if they measurably win.
5. **Rounds.** Repeat fan-out until dry — two consecutive rounds with zero new
   CONFIRMED findings. Track in `PROGRESS.md` (this dir).
6. **Output.** Committed fixes (each with a regression test), kept perf wins
   (with before/after benchstat numbers), and a log of rejected findings so
   they are not re-investigated.

## Lenses (round 1)

- **L1 concurrency/race** — daemon serve loop, hookcore MemStore/StateStore,
  cache shared state, analytics rotation lock. (`-race`, atomics, map access.)
- **L2 resource leaks** — fd/goroutine/RAM: daemon conns, file handles, tickers,
  goroutine exit on shutdown, unbounded growth.
- **L3 decode-safety** — attacker/model-controlled sizes: protocol decode,
  detect (JSON/columnar), cache blob/qdf decode, allocation bounded by input.
- **L4 silent-failure** — swallowed errors, wrong fallback, error paths that
  emit empty/incorrect output instead of safe passthrough.
- **L5 zero-copy lifetime** — pooled read buffers (daemon), `qdf.WithNoCopy`
  aliasing, bytesconv S2B/B2S, any retained slice into a reused/freed buffer.
- **L6 pipeline correctness** — summarizers (JSON/gotest/gitlog/bench), delta,
  §ref dedup determinism, squeeze RLE, grep group, glob tree: edge cases,
  never-worse guards, off-by-one, rune boundaries.
- **L7 lifecycle/cmd** — init merge/prune/upgrade idempotence, daemon
  ensure/version-handshake, stats/gc/expand.

## Exit criteria

Two dry rounds; `-race` suite green; all kept perf changes benchstat-backed;
PROGRESS.md reflects every finding's disposition.
