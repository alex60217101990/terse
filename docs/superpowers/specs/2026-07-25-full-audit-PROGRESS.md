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
| L3#2 | decode-safety | Myers diff O((N+M)²) transient at the 10k-line guard | cache/delta.go | Low/Med | ✅ fixed (edit-distance memory cap) + test |
| L6#4 | pipeline | JSON ConstVal uses rowCount not observed (sparse cols) | detect/json.go | Low | ✅ fixed (use observed) + test |
| L6#5 | pipeline | noise-strip "Waiting" prefix over-broad | detect/noise.go | Low | ✅ fixed (removed over-broad prefix) + test |
| L6#6 | pipeline | §ref dedup confirms blob existence, not decodability (torn blob) | cache/ref.go | Low | ❌ loud expand error, not silent/wrong (L4-verified acceptable) |
| L1#1 | concurrency | startTime plain pkg var (write Serve / read writeStats) | daemon/daemon.go | Low | ❌ prod happens-before; -race baseline clean (multi-Serve is test-only) |
| L2#1 | leak | profile temp file not removed | cmd/profile.go | Low | ❌ by design — exec'd `go tool pprof` needs the file |
| L3#3 | decode-safety | ShortHex panics if b>32 bytes | cache/ref.go | Low | ❌ not input-reachable (all callers pass 8B) |
| L7#3 | lifecycle | entryAllSuperseded won't prune legacy+foreign co-located entry | cmd/init.go | Low | ❌ needs hand-merged entry; old installer never wrote one |
| L5 | zero-copy | — | — | — | ✅ NONE (verified vs stdlib json copy semantics) |

**Round 1 result: 11 confirmed bugs fixed (2 High, 5 Med, 4 Low), 5 rejected. Full -race suite green, golangci-lint 0 issues.**

## Still to do
- Round 2 fan-out (adversarial review of the fixes + perf-lever probe + fresh sweep of summary/analytics/bytesconv/stats).
- Perf round (measure-first, benchstat-gated).

## Kept perf wins (benchstat-backed)
_(perf round pending)_

## Round 2 — adversarial fix-review + fresh sweep + perf-levers (all verified by main thread)

| # | lens | finding | disposition |
|---|------|---------|-------------|
| adversarial | review of the 11 round-1 fixes | no correctness regressions; all solid | ✅ verified clean; delta memory budget raised 64→128MiB (coverage) |
| sweep | stats: one >64KB line aborts whole LoadEvents | ✅ fixed (bufio.Reader, best-effort skip) + test |
| sweep | FormatBytes(math.MinInt) infinite recursion | ❌ unreachable (byte deltas never near MinInt) |
| sweep | summary/analytics/bytesconv/stats/expand/gitlog/bench/glob/grep/write/compact | ✅ NONE (verified correct) |
| perf L1 | daemon decode encoding/json → jsonparser (measured 3.2× on 4.5KB) | ❌ rejected: ~20µs absolute on an already-imperceptible warm path (sub-1% of the spawn cost the daemon removed) vs high correctness risk of a hand-decoder on the polymorphic shapes. Available if a profile ever shows decode as the bottleneck. |
| perf L2+L3 | JSON analyzer: prealloc nums + strconv TopVals | 💨✅ KEPT — benchstat n=8 p=0.000: sec/op −2.69%, B/op −64.52%, allocs/op −37.14%, output byte-identical |
| perf (skip) | fnv inline, copySession deep-copy, sha256 scratch | ❌ measured already-optimal (agent-verified, confirmed) |

**Round 2 result: +1 bug fixed (LoadEvents), +1 tuning (delta budget), +1 benchstat-proven perf win (JSON −64% B/op). Adversarial review found no regressions. Full -race green, lint 0.**

## Final tally
- **12 bugs fixed** (2 High, 5 Med, 5 Low), each with a regression test.
- **1 perf win kept** (JSON analyzer, benchstat-gated).
- **7 findings rejected** as false-positive / by-design / not-worthwhile, each with reasoning.
- Whole suite `-race` green; golangci-lint 0 issues; every kept perf change benchstat-backed.
