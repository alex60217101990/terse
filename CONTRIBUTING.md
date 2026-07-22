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

Go 1.26+ is required (the code relies on `unsafe.String`, `b.Loop()`, and the
1.26 allocator behavior).

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
`encoding/json`:

- **`OptBalanced`** for hot read paths — repetitive-payload default, ~38× faster
  decode than JSON at equal wire size.
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

## Commits

- [Conventional Commits](https://www.conventionalcommits.org/): `feat`, `fix`,
  `perf`, `docs`, `test`, `refactor`, `chore`, with an optional scope
  (`feat(cache): ...`).
- One logical change per commit. Perf commits should cite the `benchstat` delta.

## Adding a compressor

1. Put detection + summarization in `internal/detect` (pure, table-testable).
2. Wire it into the relevant handler in `internal/hook`.
3. Add a benchmark and a never-worse test (summary strictly shorter; malformed
   input passes through unchanged).
4. Register the subcommand in `cmd/qdf-hook/main.go` if it's a new hook.
