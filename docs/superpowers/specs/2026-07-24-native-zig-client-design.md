# Native Zig-Built Client, Windows Build, Containerization, CI/CD & Daemon Tuning — Design

**Date:** 2026-07-24
**Status:** Approved (design), pending implementation plan
**Scope:** Replace the external `nc` dependency in the hook fast-path with a purpose-built client, make the binary build on every platform, add minimal container images, a full best-practice Makefile + CI/CD pipeline, daemon profiling, and fix the `stats` impact-bar visualization.

---

## 1. Motivation & Measured Findings

The hook fast-path currently shells out to `nc` (netcat) to reach the resident daemon: `nc <flags> -U <sock> 2>/dev/null || <exe> post`. This has three problems:

1. **External dependency.** `nc` flags are platform-specific (macOS BSD `nc` uses `-U`, no `-N`; Linux/BSD OpenBSD `nc` needs `-N -U`). This already caused a silent production bug: passing `-N` on macOS errored, so the daemon path always fell through to the slow CLI spawn.
2. **Not portable.** The binary does **not build on Windows at all** — `internal/daemon/daemon.go` uses `syscall.SysProcAttr{Setsid: true}`, a Unix-only field, so `GOOS=windows go build` fails to compile.
3. `nc` is not present by default on Windows.

### 1.1 What the measurements proved (measure-first)

All figures below are medians on the dev machine (Apple silicon), payload ~2.3 KB, N=250 per candidate, interleaved harness. The machine's fork/exec floor is inflated (~5–6 ms for `/usr/bin/true`); **read the relative deltas, not absolutes** — on normal hardware every number shrinks proportionally but the ranking holds.

| Fact | Evidence |
|---|---|
| **Process spawn dominates hook latency (~94%); socket I/O is ~0.8%** | nc spawn+roundtrip = 8501 µs; in-process socket roundtrip (bench) = 69 µs = 0.8% |
| **No pure-Go one-shot client can beat `nc`** | best Go recipe (`net.Dial` + `GOMAXPROCS=1`) = +23% vs nc; even min trails ~500 µs |
| **The Go runtime bring-up alone ≈ nc's entire roundtrip** | `G_noop` (start runtime, exit) = +2.6 ms over exec floor, already slower than nc's full roundtrip. This is why Go loses structurally. |
| **Dropping `net` for raw `syscall` gives nothing** | `G_syscall` ≈ `G_net` within noise; binary size is not the lever — fixed runtime bring-up is |
| **`GOMAXPROCS=1` is the only measurable Go lever** | ~2% faster startup (fewer P's spun up) |
| **Buffer pools / unsafe / zero-copy are useless in a one-shot client** | the process lives one roundtrip then exits — nothing to reuse; those tricks optimize the 0.8% I/O path |
| **Only a native (no-runtime) client beats nc** | tiny C client `qc` = −11% median / −22% min vs nc, near the practical ceiling (only `/usr/bin/true` is faster, and it does nothing) |
| **cgo cannot help** | Go 1.26 sped up cgo *call overhead* ~30%, but that is the boundary on the 0.8% path; a cgo client still pays full Go bring-up first, then adds the boundary — strictly worse than pure Go. cgo also breaks cross-compilation and produces dynamically-linked binaries. |
| **`zig cc` delivers C-floor speed AND painless one-machine cross-compilation** | `qc.c` built with `zig cc` ≡ native `cc` build within noise, both below nc; one darwin/arm64 host cross-built correct static musl binaries for linux/amd64 and linux/arm64 in one command each, no sysroot, no QEMU |
| **`qc.c` (POSIX sockets) cannot target Windows** | `zig cc -target x86_64-windows` fails: `sys/socket.h` not found — Windows uses winsock2. Windows is therefore covered by the pure-Go client, not the native one. |

### 1.2 Design conclusion

- **Unix (darwin, linux, *bsd): a purpose-built native C client `qdf-hookc`**, cross-compiled with `zig cc` — C-floor speed (−11% vs nc), static, zero external dependency.
- **Windows and universal fallback: the pure-Go client built into `qdf-hook post`** — dials the socket itself, falls back to inline dispatch if the daemon is down. Correctness guaranteed everywhere; speed layered on top where the native client exists.
- The `nc` dependency, `ncArgs()`, and `QDF_NC_ARGS` are removed entirely.

The absolute win of the native client over nc is small on a spawn-bound path; the durable wins are **zero external dependency**, **single self-contained toolchain story**, **Windows support**, and **removal of a fragile flag that already broke silently**. This was chosen with the tradeoff understood.

---

## 2. Global Constraints

Every task inherits these.

- **Go 1.26** (`go.mod` is the version source of truth).
- **Max performance is a *measured* principle, not a license for unsafe code.** Every performance-motivated change passes the measure-first gate: `benchstat` over n≥12 interleaved base-vs-head runs, keep/revert decision, reverts recorded with the reason. Behavior-identical changes are test-gated only. The performance toolbox (stack allocation, `sync.Pool` **only in the long-lived daemon**, zero-alloc hot paths, `strings.SplitSeq`/iterators, pre-sized buffers, `bytesconv`, avoiding reflection) is applied **where measurement shows a win on the real workload** — never speculatively. The one-shot client is explicitly **not** a target for these tricks (proven useless: startup dominates).
- **Go binary is `CGO_ENABLED=0`** (static, no libc, fast load, trivial cross-compile). The native client is C compiled by `zig cc`.
- **The native client output is byte-identical to the current `nc` path** and to the daemon's own reply. Never-worse: a client failure always falls through to a correct inline dispatch.
- **Conventional Commits**; **never** a `Co-Authored-By` trailer.
- Chat in Russian; all code, comments, docs, commit messages in English.
- Security suite, release credentials, ghcr push, and cosign signing run in CI via repository secrets — not from a local dev session.

---

## 3. Architecture

### 3.1 Client (two layers)

```
Claude Code hook invocation
        │
        ▼
Unix:    qdf-hookc <sock> 2>/dev/null  ──socket OK──▶  daemon (warm, ~C-floor)
        │  (native, zig-built)
        │ nonzero (daemon down / client absent)
        ▼
        qdf-hook post   ── socket-first ──▶ daemon
                        └─ dial fails ────▶ inline dispatch (correct, slow path)

Windows: qdf-hook post  (native client absent; post does socket-or-inline)
```

**`qdf-hookc`** (`client/qc.c`, ~30 lines, POSIX): `socket(AF_UNIX) → connect → read all of stdin → write to socket → shutdown(SHUT_WR) → copy socket→stdout until EOF → exit`. The daemon's completion signal is EOF on the read side, so the half-close (`shutdown(SHUT_WR)`) is mandatory — verified against the daemon protocol (no newline/length framing; `readRequest` reads to EOF). Exit non-zero on any connect/IO error so the shell `||` falls through.

**`qdf-hook post` / `pretooluse` (Go).** New behavior: try `net.DialUnix` to `daemon.SockPath()`; on success write the request, `CloseWrite()`, `io.Copy` the reply to stdout, exit. On dial failure, run the existing inline dispatch (`hook.Dispatch…`). This is:
- the fallback when the daemon is down (was: `|| qdf-hook post` re-spawn — now one process instead of two),
- the Windows path,
- the universal correctness guarantee if `qdf-hookc` is missing.

The Go client uses idiomatic `net.DialUnix` (raw `syscall` measured to give nothing) and is invoked with `GOMAXPROCS=1` in the env where it matters (the ~2% lever), set in the hook command or via `runtime.GOMAXPROCS` — decision deferred to the plan after a confirming micro-measurement.

### 3.2 Windows build fix

Split the daemon's detached-exec into build-tagged files:
- `internal/daemon/detach_unix.go` (`//go:build !windows`): `SysProcAttr{Setsid: true}`.
- `internal/daemon/detach_windows.go` (`//go:build windows`): `SysProcAttr{CreationFlags: CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS}` (or a documented no-op stub if full daemonization on Windows is deferred).

Goal: `GOOS=windows go build ./...` compiles clean. Full Windows daemon lifecycle is best-effort; the guaranteed Windows story is the inline `qdf-hook post` path.

### 3.3 Daemon profiling & tuning

- **`--pprof <addr>` flag / `QDF_PPROF` env** on `daemon --serve`: when set (default off), start `net/http/pprof` bound to a **loopback address only** (security). Off by default so production daemons expose nothing.
- **`STATS` control word** over the socket (peer of `PING`/`QUIT`): daemon replies with a compact snapshot — `runtime.MemStats` highlights (HeapAlloc, HeapInuse, NumGC, PauseTotalNs), `NumGoroutine`, live-connection and served-request counters, uptime. Lightweight tuning without a full profile.
- **`qdf-hook profile [cpu|heap|goroutine|allocs|mutex|block] [-o <file>] [-d <dur>]`** subcommand: fetches the named profile from the running daemon's pprof endpoint (starting it transiently if needed via a control word, or requiring `--pprof` — decided in the plan), writes it to a file, or launches `go tool pprof` interactively when no `-o` is given. `cpu`/`block`/`mutex` honor `-d`.
- **`daemon --ping`** flag: connect, send `PING`, print the version reply, exit 0/1 — used by the container HEALTHCHECK (scratch image has no shell).

### 3.4 Stats visualization fix

`internal/analytics/stats.go` renders the per-hook "impact" bar as `meter(saved/maxSaved, 12)` where `maxSaved` is the single largest hook's absolute saved bytes. Under skew (one hook dominates total savings — e.g. Read at 23 MB vs Edit at 898 KB), every other bar floors to zero cells even at 99% per-hook efficiency, so the column goes blank for all but the top hook.

**Fix:** perceptual scaling — `meter(sqrt(saved)/sqrt(maxSaved), width)` — with a **floor of 1 filled cell for any non-zero saved** (and 0 cells only for exactly 0). The dominant hook still reads as the fullest bar (honest), but meaningful mid contributors remain visible. Pin with a golden test over a deliberately skewed distribution.

---

## 4. Containerization (`deploy/docker/`)

Two minimal, best-practice, `scratch`-final images.

```
deploy/docker/
  daemon/Dockerfile     # Go daemon + CLI
  client/Dockerfile     # native zig-built client
  .dockerignore
```

**`daemon/Dockerfile`** — multi-stage:
- builder `golang:1.26` → `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.appVersion=$VERSION" -o /qdf-hook ./cmd/qdf-hook`
- final `FROM scratch`: `COPY` the binary; `USER 65532:65532` (numeric, no `/etc/passwd` needed); `ENV HOME=/data`; `VOLUME /data` (socket + cache); `HEALTHCHECK` via `["/qdf-hook","daemon","--ping"]`; OCI image labels; `ENTRYPOINT ["/qdf-hook"]`; `CMD ["daemon","--serve"]`.
- No CA certs (qdf-hook makes no network calls), no shell, no package manager.

**`client/Dockerfile`** — multi-stage:
- builder installs pinned Zig, runs `zig cc -O2 -target $ARCH-linux-musl -o /qdf-hookc client/qc.c && strip /qdf-hookc`
- final `FROM scratch`: `COPY /qdf-hookc`; `ENTRYPOINT ["/qdf-hookc"]`.
- Static musl → runs on scratch with zero deps.

**Multi-arch:** buildx with QEMU (`docker/setup-qemu-action`) for `linux/amd64,linux/arm64`, as requested. Note recorded in docs: because Zig and `CGO_ENABLED=0` Go both cross-compile natively, a QEMU-free variant (cross-compile then `buildx imagetools create` a manifest) is available and faster; QEMU is the default for conventional robustness.

**macOS caveat (documented):** daemon-in-container + host-hook-over-socket relies on a bind-mounted `~/.qdf-hook` ↔ `/data`. Works on Linux. On macOS Docker Desktop, AF_UNIX over the VM bind-mount boundary is frequently broken — native daemon recommended on macOS; containers are best on Linux.

---

## 5. Build Automation — Makefile

Best-practice targets. Local variants of `modernize`/`align` mutate; the `check` target runs them in non-mutating report mode for CI.

| Target | Command |
|---|---|
| `build` | `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.appVersion=$(VERSION)" ./cmd/qdf-hook` |
| `client` | `zig cc -O2 -o bin/qdf-hookc client/qc.c` (host target) |
| `client-all` | `zig cc` cross for `darwin/{amd64,arm64}`, `linux/{amd64,arm64}` + `strip` |
| `modernize` | `go run golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@latest ./... --fix` |
| `align` | `go run github.com/dkorunic/betteralign/cmd/betteralign@latest -apply ./...` |
| `fix` | `go fix ./...` |
| `fmt` | `gofumpt -w .` (+ `goimports`) |
| `tidy` | `go mod tidy` |
| `vet` | `go vet ./...` |
| `lint` | `golangci-lint run --timeout=5m` |
| `test` | `go test -race -count=1 -timeout=10m -coverprofile=coverage.out -covermode=atomic ./...` |
| `bench` | `go test -bench=. -benchmem -count=12 ./...` (benchstat-ready) |
| `cover` | coverage → HTML |
| `check` | `fmt` + `vet` + `lint` + `modernize`(report) + `align`(report) + `test` — the CI gate, non-mutating |
| `docker-daemon` / `docker-client` / `docker-all` / `docker-push` | buildx image builds/push |

Version is injected via `-ldflags -X main.appVersion` (already an overridable `var appVersion`), stamped from `git describe`.

---

## 6. CI/CD (`.github/`) — adapted from the qdf project

The qdf project's `.github` is the blueprint. Its security and hygiene suite is project-agnostic and copied nearly verbatim; qdf is a *library* (hand-rolled release, no binaries, no Docker, no Makefile, no GoReleaser), so the release, container, and Zig portions are net-new for this project.

**Copied ~as-is:** `codeql.yml` (security-extended, weekly + push/PR), `govulncheck.yml` (daily + push/PR), `scorecard.yml` (OpenSSF weekly), `dependabot.yml` (gomod weekly, actions monthly). Conventions: SHA-pinned actions with `# vN`, top-level `permissions: contents: read` + per-job elevation, `actions/setup-go@v6` with `go-version-file: go.mod` + cache, `concurrency` cancel-in-progress.

**`ci.yml`** (net-new specifics):
- `test` matrix over **native runners** `[ubuntu-latest, ubuntu-24.04-arm, macos-latest]`, `fail-fast: false`; warm-module-cache retry step; `go vet ./...`; `go test -race -count=1 -timeout=10m -coverprofile -covermode=atomic ./...`; `codecov/codecov-action` (ubuntu only, non-blocking).
- `lint` job: `golangci-lint` v2 using qdf's config as the base plus this project's build tags and any C-interop path exclusions.
- `client` job: install Zig (`goto-bus-stop/setup-zig` or pinned download), `make client-all`, run the `qdf-hookc`↔daemon parity test.

**`release.yml` — GoReleaser** (net-new; qdf has none):
- `builds`: Go daemon/CLI (`-trimpath`, `-ldflags="-s -w -X main.appVersion={{.Version}}"`, `CGO_ENABLED=0`) for darwin/linux (amd64, arm64) and windows/amd64; plus `builder: prebuilt` entries for the Zig-cross-built `qdf-hookc` binaries produced in a preceding CI step.
- `archives`, `checksums`, **SBOM via syft**, **cosign signing**, `changelog`.
- `dockers` + `docker_manifests`: the two scratch images, multi-arch, pushed to `ghcr.io`.
- `SOURCE_DATE_EPOCH` for reproducible builds. Trigger on `v*.*.*` tags; validate semver; re-verify tests at the tagged commit before publishing.

**Docker CI:** `docker/setup-qemu-action` + `docker/setup-buildx-action` + `docker/build-push-action` with `platforms: linux/amd64,linux/arm64`, provenance + SBOM attestations, ghcr login via `GITHUB_TOKEN`.

**Enforcement note:** `modernize`/`betteralign`/`gofumpt` are run as **gates** here (qdf keeps them local-only). CI runs them in non-mutating mode and fails on a diff; a `dependabot-auto-tidy`-style job may auto-apply `go mod tidy` on dependency-bump PRs.

---

## 7. Task Breakdown (features first, docs last)

1. **Stats-viz fix** — sqrt scaling + min-1-cell floor + golden test. Self-contained.
2. **Go socket-client** in `post`/`pretooluse` (socket-first → inline fallback) + tests + byte-parity.
3. **Windows build** — `detach_unix.go`/`detach_windows.go` build tags; `GOOS=windows go build ./...` green.
4. **`qdf-hookc`** — `client/qc.c` in repo + Makefile `zig cc` cross targets + `qc`≡`nc` parity test.
5. **init rework** — remove `nc`/`ncArgs`/`QDF_NC_ARGS`, platform-branch command, invoke `qdf-hookc`, upgrade stale entries + tests.
6. **Profiling** — `--pprof`/`QDF_PPROF`, `STATS` control word, `qdf-hook profile` subcommand, `daemon --ping`.
7. **Dockerfiles** — `deploy/docker/{daemon,client}/Dockerfile` (scratch, best-practice) + `.dockerignore`.
8. **Makefile** — all targets in §5.
9. **CI/CD** — security suite + `ci.yml` (+Zig job) + GoReleaser release + docker buildx/QEMU.
10. **Parity + perf finalization** — benchstat vs nc, `-race` ×3, apply `modernize`/`align` across the base (each change measure-first-gated).
11. **Docs (LAST)** — README/ARCHITECTURE update + `docs/CONTAINERS.md` full tutorial + `docs/PROFILING.md` + stats section.

---

## 8. Testing Strategy

- **Parity:** `qdf-hookc` output byte-identical to `nc -U` and to the daemon's own reply, over representative payloads (Read/Edit/Write/Bash shapes). Table test invoking both against a test daemon.
- **Fallback:** with the daemon down, `qdf-hook post` produces the same result via inline dispatch as the socket path would.
- **Cross-compile smoke:** CI builds `qdf-hookc` for all Unix targets and runs the native-arch ones against a daemon.
- **Windows compile:** `GOOS=windows go build ./...` in CI (compile-only gate).
- **Stats golden:** skewed distribution renders visible bars for all non-zero hooks; dominant hook is fullest.
- **Perf:** benchstat gate per §2 for any perf change; native client vs nc recorded.
- **Race:** `go test -race` ×3 clean for daemon/client/hook packages.

---

## 9. Out of Scope

- A winsock (`winsock2.h`) native client for Windows — Windows uses the Go client. Revisit only if a Windows fast-path is demanded.
- Standalone C toolchain matrices (mingw etc.) — `zig cc` is the single cross-compiler.
- CGO anywhere in the Go binary — measured strictly worse.
- Raw-syscall / unsafe / buffer-pool micro-optimization of the one-shot client — measured useless.
- Real registry pushes, cosign key material, and branch protection — operated via CI secrets / the maintainer's authenticated session, not this design's implementation.
