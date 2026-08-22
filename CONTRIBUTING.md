# Contributing to qdf-hook

Thanks for helping make Claude Code cheaper to run. This project has a small,
strict set of conventions — following them keeps the tool fast and correct.

## Development loop

```bash
go test -race ./...                      # all tests must pass, race-clean
go test -bench=. -benchmem -count=6 ./... # benchmarks
gofmt -w .                                # formatting
go vet ./...
```

Go 1.27+ is required (the code relies on `unsafe.String`, `b.Loop()`, the
`encoding/json/jsontext` decoder, and the 1.26+ allocator behavior).

## The one hard rule: measure before you optimize

Every performance change is **gated by measurement**, not intuition. Roughly
half of plausible optimizations turn out to be neutral or worse once measured.

1. **Interleaved `benchstat`, n ≥ 12.** Build two test binaries and alternate
   them in one loop — never run all of base then all of head (a warming laptop
   manufactures false regressions):

   ```bash
   go test -c -o /tmp/base ./internal/cache/    # on main
   # apply change
   go test -c -o /tmp/head ./internal/cache/
   cd internal/cache   # run from the package dir — benchmarks use relative testdata paths
   for i in $(seq 12); do
     /tmp/base -test.bench=BenchmarkX -test.benchmem -test.count=1 >> b.txt
     /tmp/head -test.bench=BenchmarkX -test.benchmem -test.count=1 >> h.txt
   done
   benchstat b.txt h.txt
   ```

2. **Keep only wins that clear the noise.** On a laptop, thermal/scheduling
   noise is ±15 %. "Within noise" is a NO-GO regardless of the mean.

3. **Record reverts.** If an idea loses, note *why* in the PR so nobody retries
   it. (Example already in the tree: a `json.Decoder.Token()` rewrite of the
   JSON analyzer lost to interface boxing — +14 % slower, +120 % allocs.)

## Serialization: use qdf

All on-disk state uses [`qdf`](https://github.com/alex60217101990/qdf), not
`encoding/json`. The option picked per store matters and isn't uniform:

- **`OptSpeed`** for session state (`internal/cache/store.go`) — it's written
  on *every* hook, so minimizing per-Marshal allocs beats a smaller wire size;
  `OptBalanced`/`OptCompression` add codec overhead (Dense/QPack/ShapeIntern,
  rANS/FSST/Gorilla) that doesn't pay off at this write frequency.
- **`OptBalanced`** for `§ref` blobs, the rerun-delta `last/` store, and the
  usage sidecar index (`internal/cache/ref.go`, `lastout.go`, `usage.go`) —
  the repetitive-payload default, ~38× faster decode than JSON at equal wire
  size; these are read far more often than written.
- **`WithNoCopy`** to decode string/`[]byte` fields as aliases into the read
  buffer — only where that buffer outlives the value and is never mutated.
- **Plain `os.WriteFile`** (no tmp+rename) for the rebuildable caches — a torn
  write just fails to decode and is treated as a cache miss.
- `OptCompression` is for **cold archives only** (~38× slower encode) — never a
  hot path.

## Correctness rules for hooks

- Hooks must **never change what a tool does** — they run in `PostToolUse`
  (after execution) or only *deny* a redundant re-read in `PreToolUse`.
- All persisted state is a **rebuildable cache**: a decode failure must always
  degrade to a fresh, correct result, never wrong content or a crash.
- Any compressor must be **never-worse**: only replace output when the
  replacement is strictly smaller than the original.
- Handle each tool's **real `tool_response` shape**, not an assumed one: Read
  nests content under `file.content` (with window metadata), Edit/MultiEdit
  have no plain-text field at all (`originalFile`/`oldString`/`newString`/
  `structuredPatch`, and `originalFile` can be the whole pre-edit file), Bash
  uses `stdout`/`stderr`. A windowed Read (`offset`/`limit`, or a
  `startLine`/`numLines` partial) must pass through uncached — never fed into
  delta/unchanged tracking, which assumes it's looking at the whole file.

## Correctness rules for the daemon

`qdf-hookd` (`internal/daemon`) is long-lived and handles concurrent
connections, which is a different risk profile from the one-shot CLI:

- **Never let a connection hang the daemon.** Every connection gets a bounded
  deadline (`connDeadlineNS`) covering read + dispatch + write; an unbounded
  `io.ReadAll` on a client that never half-closes would otherwise leak the
  handler goroutine and block clean shutdown via `Serve`'s `wg.Wait()`.
- **A panic in one connection must never take down the daemon.** `handleConn`
  recovers and just closes the connection — the hybrid client's CLI fallback
  covers the dropped request.
- **The daemon and the CLI must produce byte-identical hook output** for the
  same input. Both call the same `hook.Dispatch`/`DispatchBytes` — the only
  difference is which `hookcore.StateStore` backs the call
  (`DiskStore` for the CLI, `MemStore` for the daemon). New hook logic goes in
  `internal/hook` against the `hookcore.StateStore` interface, never
  hardcoded against one backend.
- **The daemon must restore normal runtime behavior.** `main.init()`'s
  `GOMAXPROCS(1)` + `SetGCPercent(-1)` (fast one-shot CLI startup) would leak
  and serialize connections if left in place for a long-lived process —
  `daemon --serve` calls `restoreDaemonRuntime()` first.
- **Flush and sweep on a schedule, not just on exit.** `Serve` flushes dirty
  state every 5 s and sweeps `refs/`/`last/` to their bounds every 10 minutes,
  so a crash or a long session never loses more than one interval's worth of
  state or lets the cache grow unbounded.

## Commits

- [Conventional Commits](https://www.conventionalcommits.org/): `feat`, `fix`,
  `perf`, `docs`, `test`, `refactor`, `chore`, with an optional scope
  (`feat(cache): ...`).
- One logical change per commit. Perf commits should cite the `benchstat` delta.

## Adding a compressor

1. Put detection + summarization in `internal/detect` (pure, table-testable).
2. Wire it into the relevant handler in `internal/hook`. If it's a tool-agnostic
   shape (not a new Read/Write-style handler), wire it into `handleGeneric`
   (`internal/hook/dispatch.go`) instead — every tool already flows through
   `Dispatch`/`post`, so it needs no new subcommand or `settings.json` entry.
3. Add a benchmark and a never-worse test (summary strictly shorter; malformed
   input passes through unchanged).
4. Register a new subcommand in `cmd/qdf-hook/main.go` only if it's a genuinely
   new hook *event* (not just a new tool shape) — most additions don't need one.
