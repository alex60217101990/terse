# qdf-hook v2: Analytics, Utility-Based Eviction, PreToolUse Interception, Extended Hook Coverage

## Goal

Extend qdf-hook from a PostToolUse-only compressor into a full-coverage Claude Code token
optimizer that intercepts Read tool calls before they execute, caches Bash output, compresses
Glob and Write/Edit responses, tracks per-invocation analytics, and provides a `stats`
subcommand with real token-savings accounting.

## Background

qdf-hook v1 (feat/qdf-hook branch) already implements:
- PostToolUse Read: unchanged/delta/full modes with Myers diff
- PostToolUse Bash: JSON array, go test, git log, benchmark summarizers
- PreCompact/PostCompact/SessionStart: file manifest injection after compaction

v2 adds seven new capabilities described below.

---

## Capability 1: FileEntry Utility Fields

`internal/cache/state.go` — extend `FileEntry` with three new fields:

```go
type FileEntry struct {
    Hash       [32]byte `json:"hash"`
    Turn       int      `json:"turn"`
    Content    []byte   `json:"content"`
    ModTime    int64    `json:"mtime"`  // unix nanoseconds from os.Stat
    ReadCount  int      `json:"rc"`     // incremented on every read (pre or post)
    LastReadAt int64    `json:"lra"`    // unix seconds of last read
}
```

These fields power both the PreToolUse mtime fast-path and the LRU utility score.

---

## Capability 2: Utility-Score Eviction (Exponential Decay)

`internal/cache/evict.go` — eviction by combined recency+frequency score:

```go
func utilityScore(entry FileEntry, nowSec int64) float64 {
    ageHours := float64(nowSec-entry.LastReadAt) / 3600.0
    lambda := decayLambda() // from QDF_DECAY_LAMBDA env, default 0.1
    return float64(entry.ReadCount) * math.Exp(-lambda*ageHours)
}
```

`decayLambda` reads `QDF_DECAY_LAMBDA` env var (float64, default 0.1). At λ=0.1:
- File read 50× an hour ago: score ≈ 45
- File read 50× two days ago: score ≈ 0.4
- File read 1× an hour ago: score ≈ 0.9

`Evict(state *SessionState, maxFiles int)` — called by `Save` automatically when
`len(state.Files) > maxFiles` (default 200). Sorts by score ascending, drops bottom 20%.

`GCSessionFiles(maxAgeDays int, lambda float64)` — scans all `.qdf` session files,
computes a per-session utility score from total reads and session age, removes sessions
below threshold. Called by `qdf-hook gc`.

---

## Capability 3: PreToolUse Read Interceptor

New subcommand `qdf-hook pretooluse`. Registered as a PreToolUse hook in settings.json
with matcher `"Read"`.

Protocol (Claude Code PreToolUse):
- stdout: `{"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "deny"|"allow", "permissionDecisionReason": "..."}}`
- exit 0 always

Algorithm:
1. Read stdin (session_id, tool_input.file_path)
2. `os.Stat(path)` for mtime — no content read
3. Load session state
4. If `entry.ModTime == stat.ModTime.UnixNano() && state.SeenAfterCompact(path)`:
   - Increment `entry.ReadCount`, update `entry.LastReadAt`, save state
   - Append analytics event (action="pretool-unchanged")
   - Return `deny` + `permissionDecisionReason: "§unchanged:HASH§ PATH — mtime unchanged, cached at turn N. No re-read needed."`
5. Else: return `allow` — PostToolUse Read will handle it and update mtime

`internal/protocol/protocol.go` — add `EncodePre(w io.Writer, decision, reason string) error`.

---

## Capability 4: Analytics

`internal/analytics/analytics.go` — append-only JSONL writer.

Event struct:
```go
type Event struct {
    TS     int64  `json:"ts"`    // unix nanoseconds
    SID    string `json:"sid"`   // first 16 chars of session_id
    Hook   string `json:"hook"`  // pretooluse|read|bash|glob|write|precompact|postcompact
    Action string `json:"action"`// pretool-unchanged|full|unchanged|delta|summary|passthrough|compressed|tree
    BytesIn  int  `json:"bi"`    // original content size in bytes
    BytesOut int  `json:"bo"`    // emitted content size in bytes
    DurNS  int64  `json:"dur"`   // hook duration in nanoseconds
}
```

`Record(e Event) error` — opens `~/.qdf-hook/analytics.jsonl` with `O_APPEND|O_CREATE`,
writes one JSON line + newline, closes. POSIX append is atomic for writes < PIPE_BUF
(4096 bytes); our records are < 200 bytes. Returns error only; never blocks the hook.

Rotation: if file exceeds 10MB on open, rename to `analytics.jsonl.1` first.

Token savings formula: `saved_tokens = (BytesIn - BytesOut) / 4` (4 bytes per token,
standard Claude approximation).

---

## Capability 5: Bash Output Caching

`internal/hook/bash.go` — extend existing `HandleBash` with a caching layer before the
detector pipeline.

Bash cache: `~/.qdf-hook/bash-cache/<sha256(command+"\x00"+cwd)[:16]>.entry`
Each entry: `{"hash": "<sha256(output)>", "output": "...", "ts": <unix-sec>}`
TTL: 30 seconds (configurable via `QDF_BASH_CACHE_TTL_SEC`, default 30).

Read-only command whitelist (prefix match):
```
git status, git log, git diff, git branch, git show,
ls, find, cat, pwd, which, go env, go version,
echo, printenv, env
```

Algorithm in `HandleBash`:
1. Parse command from `tool_input.command`
2. If command matches whitelist AND cache entry exists AND age < TTL:
   - Hash current output (from tool_response) and compare to cached hash
   - If match: return `§bash-unchanged:HASH§ [command] — output identical to N seconds ago`
   - If mismatch: update cache, proceed to detector pipeline
3. If command not in whitelist or no cache: proceed to detector pipeline as before
4. After detector pipeline: if whitelisted command, update/create cache entry

---

## Capability 6: Glob Output Compressor

New file `internal/hook/glob.go` + new subcommand `qdf-hook glob`.

PostToolUse matcher: `"Glob"`.

Input: tool_response.content is a newline-separated list of file paths.
Output: compact tree representation.

Algorithm:
1. Parse paths, group by first two directory levels
2. For groups ≤ 5 files: list inline `dir/ (N files: a.go b.go ...)`
3. For groups > 5 files: `dir/ (N files)` — no inline listing
4. Append `[N total files, M dirs]` footer

Example output (50 files → ~8 lines):
```
internal/hook/     6 files (bash.go bash_test.go compact.go glob.go read.go read_test.go)
internal/cache/    5 files (delta.go delta_test.go evict.go state.go store.go)
internal/detect/   8 files
internal/analytics/2 files (analytics.go analytics_test.go)
cmd/qdf-hook/      2 files (main.go stats.go)
[23 total files, 5 dirs]
```

Only compress if output > 256 bytes AND tree is shorter than original (ratio < 0.8).

---

## Capability 7: Write/Edit Result Compressor

New file `internal/hook/write.go` + new subcommand `qdf-hook write`.

PostToolUse matcher: `"Write|Edit|MultiEdit"`.

Claude just wrote this content — no need to see it again. Replace tool response with:
```
[WRITE §ref:HASH§ /path/to/file.go — 127 lines written, cached for delta tracking]
```

Where HASH = first 8 bytes of SHA-256 of content (16 hex chars).
Store in session state as FileEntry (enables delta on next read).
Only compress if content > 256 bytes.

---

## Capability 8: Stats Subcommand

`cmd/qdf-hook/stats.go` — `qdf-hook stats [--json] [--days N] [--session SID]`.

Reads `~/.qdf-hook/analytics.jsonl` (and `.1` if exists), filters by days, computes:

```
qdf-hook  (last 7 days · 3,847 invocations)

TOKEN SAVINGS
  Original:  284.7 MB  (~71.2M tokens)
  Emitted:     6.1 MB  (~1.5M tokens)
  Saved:     278.6 MB   97.8%  (~69.7M tokens)

LATENCY  p50 / p95 / p99
  pretooluse/unchanged:  0.1ms /  0.3ms /  0.5ms
  read/unchanged:        0.4ms /  0.9ms /  1.4ms
  read/delta:            1.2ms /  3.5ms /  6.1ms
  bash/summary:         18ms  / 29ms  / 38ms
  bash/unchanged:        0.2ms /  0.5ms /  0.9ms
  write/compressed:      0.3ms /  0.8ms /  1.2ms
  glob/tree:             0.5ms /  1.2ms /  2.0ms

BY HOOK
  pretooluse: 1,204  (unchanged: 1,204)
  read:         891  (full: 203  unchanged: 580  delta: 108)
  bash:       1,421  (summary: 847  unchanged: 412  passthrough: 162)
  write/edit:   331  (compressed: 331)
  glob:           0

STATE
  Sessions:  12  (2.1 MB)
  Files tracked: 156
  Run 'qdf-hook gc' to clean stale sessions
```

`--json` outputs a JSON object with the same fields.

---

## Capability 9: GC Subcommand

`cmd/qdf-hook/gc.go` — `qdf-hook gc [--dry-run] [--days N] [--lambda F]`.

Per-session utility score:
```go
sessionScore = totalReads * math.Exp(-lambda * sessionAgeHours)
```

Sessions with score < threshold (default 0.01) are deleted.
`--dry-run` prints what would be deleted without deleting.
`--days N` forces deletion of sessions older than N days regardless of score.

Also runs `Evict` on each surviving session state.

---

## Capability 10: Profiling Flags

`cmd/qdf-hook/main.go` — persistent flags on root Cobra command:

```
--cpuprofile FILE   write CPU profile to FILE on exit
--memprofile FILE   write heap profile to FILE on exit
```

In `main()`:
```go
if cpuprofile != "" {
    f, _ := os.Create(cpuprofile)
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
}
```

---

## Analytics Wiring (all existing hooks)

Every handler records one `analytics.Event` after processing:
- `read.go`: action = "full"|"unchanged"|"delta"
- `bash.go`: action = "summary"|"unchanged"|"passthrough"
- `compact.go` (PreCompact): action = "precompact"
- `compact.go` (PostCompact): action = "postcompact"
- `pretooluse`: action = "pretool-unchanged"|"pretool-allow"
- `glob.go`: action = "tree"|"passthrough"
- `write.go`: action = "compressed"|"passthrough"

`BytesIn` = len of original tool response content.
`BytesOut` = len of what we returned (or same as BytesIn for passthrough).
`DurNS` = `time.Since(start).Nanoseconds()`.

---

## Settings.json (full configuration)

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Read",
        "hooks": [{"type": "command", "command": "qdf-hook pretooluse"}]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Read",
        "hooks": [{"type": "command", "command": "qdf-hook read"}]
      },
      {
        "matcher": "Bash",
        "hooks": [{"type": "command", "command": "qdf-hook bash"}]
      },
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [{"type": "command", "command": "qdf-hook write"}]
      },
      {
        "matcher": "Glob",
        "hooks": [{"type": "command", "command": "qdf-hook glob"}]
      }
    ],
    "PreCompact": [
      {"matcher": ".*", "hooks": [{"type": "command", "command": "qdf-hook precompact"}]}
    ],
    "PostCompact": [
      {"matcher": ".*", "hooks": [{"type": "command", "command": "qdf-hook postcompact"}]}
    ]
  }
}
```

---

## Non-Goals

- WebFetch HTML compression (content too varied, risk of info loss)
- MCP tool output compression (unknown schemas)
- PreToolUse Bash blocking (Claude Code v2.1.x ignores updatedInput.command)
- Daemon/server mode (hooks are short-lived; analytics file is the store)
