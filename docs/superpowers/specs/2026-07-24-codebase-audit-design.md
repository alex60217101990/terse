# Whole-Codebase Audit + Improvement Pass — Design

**Date:** 2026-07-24
**Status:** approved (brainstorming), audit-then-plan

## Goal

A deep, measured pass over the entire qdf-hook codebase that (a) finds and
fixes real bugs, (b) captures perf optimizations, and (c) modernizes the code
with Go 1.2x syntax — **without changing observable behavior**. Every agent
finding is independently re-verified; every perf change is benchstat-gated;
nothing merges that breaks a test.

## Scope

- **In:** all of `internal/*` (cache, hookcore, daemon, hook, detect, protocol,
  analytics, bytesconv, summary) and `cmd/qdf-hook`, including `_test.go` files.
- **Out:** the `qdf` dependency (separate repo). No API redesign that breaks
  callers. No new features.
- **Go floor:** 1.26 (all target syntax available).

## Audit dimensions (three parallel read-only agents, findings-only)

Each agent sweeps the whole codebase for its dimension and returns a ranked
list — `file:line`, a one-line claim, and a concrete failure scenario (bugs) or
qualitative win + hot/cold (perf) or behavior-identical flag (modernization).
Agents do NOT edit code. Large packages may be split across agent invocations.

1. **Bugs / correctness.** Incorrect logic, off-by-one, empty/boundary/edge
   crashes, data races, resource (fd/goroutine) leaks, error-swallowing that
   hides a real failure, unsafe/zero-copy hazards (a `bytesconv`/`WithNoCopy`
   value retained past its backing array), a compressor that can emit larger
   output without a never-worse guard, mis-parse of tool output.

2. **Perf / logic.** Allocations that could be stack/pooled; redundant
   syscalls / copies / re-hashes / re-decodes; interface boxing in hot loops;
   O(n²) where linear works; a map where a slice suffices (or vice-versa);
   `fmt.Sprintf` where `strconv`/`hex.Encode`/byte-appends work; missing
   zero-copy where sound (and misuse where unsound). Tag hot (per-hook /
   per-request) vs cold (startup / rare).

3. **Go 1.2x modernization.** `for range N` (range-over-int); `min`/`max`
   builtins; `slices`/`maps` stdlib (`slices.Contains/Sort/SortFunc/Clone`,
   `maps.Copy/Clone/Keys`); iterators (`range`-over-func, `strings.SplitSeq`
   etc.) where natural; generics to collapse duplicated non-boxing code;
   constant-size stack arrays. **Each finding tagged behavior-identical (yes/no)**
   — a `no` finding (changes output/semantics) is rejected outright per the
   no-behavior-change rule.

## Verification protocol (I re-verify everything; agents are not trusted)

- **Triage:** every finding is re-checked against the actual code before it
  enters the plan. False/speculative findings are logged with the reason and
  dropped. Findings that contradict a deliberate design choice are logged, not
  applied.
- **Behavior-identical changes** (modernization, safe refactors): gated by the
  full `-race` suite staying green, plus byte-identical output where a golden/
  parity test applies. No benchstat needed.
- **Perf-changing changes:** benchstat **n≥12 interleaved** base-vs-head,
  keep only if it clears noise; every reverted idea recorded with why. An
  alloc reduction with wall-clock within noise is keepable when it cuts
  GC pressure under sustained/concurrent load (documented, e.g. the daemon
  path), never on a synthetic-only win.
- **Bug fixes:** a regression test that reproduces the bug first, then the fix.
- **Never break behavior:** the Task-6 daemon-vs-CLI byte-identical parity test
  and the whole `-race` suite are the gate on every task. `-gcflags=-m` is used
  to confirm claimed stack/escape wins.

## Deliverables

- This spec (methodology + acceptance criteria).
- After the audit: a detailed SDD implementation plan whose tasks are the
  triaged, confirmed findings — grouped by package/theme, bite-sized, TDD,
  each independently testable and reviewable.
- A progress ledger at `.superpowers/sdd/audit-progress.md`.
- A reverts/decisions log (kept in the ledger) recording every dropped finding
  and every benchstat revert, with reasons.

## Acceptance

- Zero behavior regressions: full `-race` suite + parity test green at every
  task and at the end.
- Every kept perf change has a recorded benchstat delta clearing noise.
- Every applied bug fix has a reproducing regression test.
- Modernization changes are behavior-identical and test-covered.
- Dropped/reverted findings are logged with reasons (no silent discards).

## Sequencing note

The plan cannot be written before the audit runs — its tasks *are* the
confirmed findings. Order: commit this spec → run the three audit agents →
triage/verify findings → writing-plans from confirmed findings → SDD execution.
