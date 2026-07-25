# Full audit — progress

Legend: 🔎 reported · ✅ CONFIRMED+fixed · ❌ false-positive (rejected) · ⏳ open · 💨 perf-lever

## Round 1 — 7 lenses (L1-L7), all returned. Verified by main thread.

| # | lens | finding | file | sev | disposition |
|---|------|---------|------|-----|-------------|
| L6#1 | pipeline | unified diff overlapping hunks (dup context, @@ overruns file) | cache/delta.go | **High** | ✅ fixed `ee03204` (merge gap ≤ 2·ctx) + regr test |
| L6#2 | pipeline | gotest crash/timeout/panic summarized as PASS | detect/gotest.go | **High** | ✅ fixed `6ce0068` (detect crash → passthrough) + tests |
| L3#1 | decode-safety | JSON columns unbounded (only rows capped) → mem amplification | detect/json.go | Med | ✅ fixed `6ce0068` (maxCols=256, lazy strFreq) + test |
| L4-F4 | silent-fail | PreToolUse deny rests on mtime+size when ctime==0 → stale content | hook/pretooluse.go | Med | ✅ fixed `89234f6` (require real ctime; else allow) + tests |
| L7#1 | lifecycle | hook command paths unquoted → space breaks exec + idempotence | cmd/init.go | Med | ✅ fixed `34f6148` (shquote + shFields) + tests |
| L4-F1 | silent-fail | daemon swallows DispatchBytes error, no observability | daemon/daemon.go | Med | ✅ fixed `436dfc3` (log to daemon.log) |
| L4-F2 | silent-fail | client deadline 2s < daemon 10s → partial reply to stdout | daemon/client.go | Med | ✅ fixed `436dfc3` (exchange ≥ 15s) + regr test |
| L4-F3 | silent-fail | misleading "fallback covers" handler comment | daemon/daemon.go | Low | ✅ fixed `436dfc3` (comment) |
| L7#2 | lifecycle | gc --dry-run omits blob eviction preview | cmd/gc.go | Low | ✅ fixed `34f6148` |
| L6#3 | pipeline | gotest panic/stack detail dropped from FAILURES block | detect/gotest.go | Med | ✅ subsumed by L6#2 (crash → full passthrough) |
| L3#2 | decode-safety | Myers diff O((N+M)²) transient at the 10k-line guard | cache/delta.go | Low/Med | ⏳ OPEN — tighten guard to bound N·M/bytes |
| L6#4 | pipeline | JSON ConstVal uses rowCount not observed (sparse cols) | detect/json.go | Low | ⏳ OPEN — cheap fix (lost signal only) |
| L6#5 | pipeline | noise-strip "Waiting" prefix over-broad + 4 prefixes unreachable | detect/noise.go | Low | ⏳ OPEN — narrow "Waiting", add discriminators |
| L6#6 | pipeline | §ref dedup confirms blob existence, not decodability (torn blob) | cache/ref.go | Low | ❌ loud expand error, not silent/wrong (L4-verified acceptable) |
| L1#1 | concurrency | startTime plain pkg var (write Serve / read writeStats) | daemon/daemon.go | Low | ❌ prod happens-before; -race baseline clean (multi-Serve is test-only) |
| L2#1 | leak | profile temp file not removed | cmd/profile.go | Low | ❌ by design — exec'd `go tool pprof` needs the file |
| L3#3 | decode-safety | ShortHex panics if b>32 bytes | cache/ref.go | Low | ❌ not input-reachable (all callers pass 8B) |
| L7#3 | lifecycle | entryAllSuperseded won't prune legacy+foreign co-located entry | cmd/init.go | Low | ❌ needs hand-merged entry; old installer never wrote one |
| L5 | zero-copy | — | — | — | ✅ NONE (verified vs stdlib json copy semantics) |

**Round 1 result: 8 confirmed bugs fixed (2 High, 5 Med, 1 Low), 3 Low still open, 5 rejected. Full -race suite green, golangci-lint 0 issues.**

## Still to do
- Finish the 3 open Lows (L3#2 delta guard, L6#4 ConstVal, L6#5 noise).
- Perf round (measure-first, benchstat-gated) — separate phase.
- Round 2 fan-out (until 2 dry rounds).

## Kept perf wins (benchstat-backed)
_(perf round pending)_
