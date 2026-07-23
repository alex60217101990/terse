# qdf-hook — Bash re-run delta + output noise-strip (design)

Date: 2026-07-23
Status: approved (brainstorm)

## Problem

The two biggest *non-obvious* token sinks in a real session:

1. **Re-run near-duplication.** Commands like `go test`, `git status`, `ls`,
   `go build` are re-run constantly; their output changes only slightly, but the
   §ref cache only collapses **byte-identical** repeats, so each re-run is
   re-ingested in full.
2. **Shell/build noise.** Every command may carry junk lines that are pure
   presentation, never signal: gvm's `zsh: command not found: _encode …`,
   `go: downloading …`, `npm notice`/funding banners, docker layer-pull
   progress, deprecation spam.

## Goals

- **Bash delta:** on a re-run of the same command+cwd, emit only a unified diff
  against the previous run when that is smaller.
- **Noise-strip:** drop known junk lines losslessly (they carry no information)
  before compression, always never-worse.

## Design

### Noise-strip (`internal/detect`)

`StripNoise(content string) string` — single pass over lines; drop a line when it
matches a known-noise predicate; return the input **unchanged** (same string
header, no copy) when nothing matched, so callers gate on identity/len.

- Predicates are cheap byte-level `strings.HasPrefix`/`Contains` checks on a
  small fixed table — **no regexp** (regex alloc/backtrack not worth it for fixed
  prefixes). Table covers: `zsh: command not found: _`, `go: downloading `,
  `npm notice`, `npm warn `, `npm fund`, docker `Pulling `/`Waiting`/`Download
  complete`/`Pull complete`, `go: finding `.
- Perf: one `strings.Builder` sized to `len(content)`; iterate lines via manual
  `IndexByte('\n')` scan (no `strings.Split` slice allocation); append kept
  lines by slicing the original (zero-copy substrings). Bail to the original
  string if zero lines dropped (no allocation in the common case).
- Applied first in `HandleBash`, before detectors/§ref/squeeze.

### Bash delta (`internal/cache` + `internal/hook`)

Per-command last-output store: `~/.qdf-hook/bashlast/<key>.blob`, `key =
sha256(command + 0x00 + cwd)[:16]`, payload a qdf `OptBalanced`
`bashLastEntry{ Output string; TS int64 }`.

- `cache.BashLastGet(cmd, cwd) (string, bool)` — `os.ReadFile` + `qdf.Unmarshal`
  with `WithNoCopy` (decoded output aliases the read buffer).
- `cache.BashLastPut(cmd, cwd, out string)` — qdf `OptBalanced` + plain
  `os.WriteFile` (no tmp+rename; rebuildable cache).
- Key hashing uses a zero-copy `unsafe.String` view of `cmd`/`cwd`; no
  intermediate concat allocation beyond the hashed bytes.

`HandleBash` flow (order):
1. `content = StripNoise(Text())`.
2. `< 256` → passthrough (record).
3. structural detectors → summary.
4. §ref (byte-identical repeat) — unchanged.
5. **bash delta:** `prev, ok := BashLastGet(cmd,cwd)`; if `ok && prev != content`,
   `d := UnifiedDiff([]byte(prev), []byte(content), 3)`; if
   `len(header+d) < len(content)` emit `[BASH §rerun-delta§ cmd — changes since
   last run]\n` + d (action `rerun-delta`). Always `BashLastPut(cmd,cwd,content)`
   afterwards (store current for next time).
6. squeeze → passthrough.

`bi` (command+cwd) parsing returns to `HandleBash` (was removed with the old
cache) — parse `tool_input` once.

## Performance (measure-first, hard requirement)

Apply every proven technique: qdf `OptBalanced`, plain write, `WithNoCopy`
decode, zero-copy `unsafe.String` hashing, stack scratch (`[64]byte` hex in
`ShortHex`), single-pass `strings.Builder` with pre-sized `Grow`, substring
slicing instead of copies, **no regexp** in the hot path.

Benchmarks (interleaved benchstat, n≥12, keep/revert):
- `BenchmarkStripNoise_Clean` (no junk → must be ~0 alloc, returns input).
- `BenchmarkStripNoise_Dirty` (drops N lines).
- `BenchmarkBashDelta_Hit` (re-run, small change → diff).
- `BenchmarkBashLastPut/Get`.
Targets: StripNoise clean < 20 µs & **0 allocs**; delta hit dominated by
`UnifiedDiff` (already benched); Put/Get < 250 µs (plain-write floor).

Never-worse everywhere: emit delta / stripped form only when strictly smaller.

## Testing

- StripNoise: drops each junk pattern; keeps signal lines; clean input returned
  byte-identical (same content, verified by content equality + never-worse).
- Bash delta: first run stores + passthrough; second run with a 1-line change →
  `§rerun-delta` smaller than full; identical re-run still handled by §ref;
  tiny output → never-worse passthrough.
- gc prunes `bashlast/` by mtime TTL (reuse `RefTTLHours`).

## Out of scope

Block-level cross-tool dedup; MCP (`mcp__*`) output hook; base64/binary → §ref;
giant-output truncate+ref. Separate follow-up specs.
