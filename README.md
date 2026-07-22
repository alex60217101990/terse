# qdf-hook

Claude Code hook binary that reduces token consumption by 50–99% on:
- **Repeated file reads** — a `PreToolUse` interceptor denies re-reading an unchanged file (mtime fast-path); the `PostToolUse` handler returns `§unchanged§` or a unified diff instead of full content
- **Structured Bash output** — collapses JSON arrays, `go test -v`, `git log`, benchmarks into compact summaries; any repeated byte-identical output (any command, any tool, across sessions) collapses to a `§ref:HASH§` token
- **Write/Edit echoes** — suppresses the full-file echo Claude just wrote, caching it for delta tracking on the next read
- **Glob listings** — compresses long flat file lists into a directory tree
- **Post-compaction re-reads** — injects a file-read manifest after compaction so Claude doesn't start from scratch

## Install

```bash
go install github.com/alex60217101990/qdf-hook/cmd/qdf-hook@latest
```

Or from source:
```bash
git clone https://github.com/alex60217101990/qdf-hook
cd qdf-hook
go build -ldflags="-s -w" -o "$(go env GOPATH)/bin/qdf-hook" ./cmd/qdf-hook/
```

## Configure Claude Code

Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {"matcher": "Read", "hooks": [{"type": "command", "command": "qdf-hook pretooluse"}]}
    ],
    "PostToolUse": [
      {"matcher": "Read", "hooks": [{"type": "command", "command": "qdf-hook read"}]},
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "qdf-hook bash"}]},
      {"matcher": "Write|Edit|MultiEdit", "hooks": [{"type": "command", "command": "qdf-hook write"}]},
      {"matcher": "Glob", "hooks": [{"type": "command", "command": "qdf-hook glob"}]}
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

> **Note:** If you already run sqz for Bash, the two overlap — both compress structured Bash output. Put `qdf-hook bash` *after* sqz in the Bash hooks array, or pick one. qdf-hook's Bash handler is a superset of sqz for structured-data detection and additionally §ref-dedups any repeated byte-identical output.

## Subcommands

| Subcommand | Hook | Purpose |
|------------|------|---------|
| `pretooluse` | PreToolUse (Read) | Deny re-read of an unchanged file (mtime fast-path) |
| `read` | PostToolUse (Read) | `§unchanged§` / unified-diff compression of file content |
| `bash` | PostToolUse (Bash) | Structured-output summaries + §ref dedup of repeated output |
| `write` | PostToolUse (Write/Edit/MultiEdit) | Suppress content echo, prime delta cache |
| `glob` | PostToolUse (Glob) | Compress flat file list into a tree |
| `precompact` | PreCompact | Mark session so next reads serve full content |
| `postcompact` | PostCompact | Inject file-read manifest into fresh context |
| `sessionstart` | SessionStart | Initialize session state |
| `stats` | — | Print token-savings analytics (`--json` for machine output) |
| `gc` | — | Evict low-utility cached sessions + prune old §ref blobs (`--dry-run`) |
| `expand` | — | Print the full content behind a `§ref:HASH§` token |
| `version` | — | Print version |

Global flags: `--cpuprofile FILE`, `--memprofile FILE` write pprof profiles.

## State files

- Per-session read cache: `~/.qdf-hook/sessions/{session_id}.qdf` (qdf-compressed).
- Content-addressed §ref blobs: `~/.qdf-hook/refs/{hash}.blob` (qdf OptBalanced; TTL `QDF_REF_TTL_HOURS`, default 168h).
- Analytics event log: appended JSONL, surfaced via `qdf-hook stats`.

All are safe to delete; qdf-hook recreates them. `qdf-hook gc` prunes stale sessions by a utility score (recency × read frequency with exponential decay).

## How it works

### Read (PreToolUse + PostToolUse)
1. **PreToolUse:** if the cached file's mtime is unchanged since last read, deny the read with a `§unchanged§` token — Claude never spends tokens re-ingesting it.
2. **PostToolUse (first read):** cache content, return it in full.
3. **PostToolUse (seen, changed):** compute a Myers unified diff, return `[§delta:HASH§ path]` + the diff only.
4. **After compaction:** treat as "not seen" — serve full content on the next read.

### Bash (PostToolUse)
1. Try detectors in order: JSON array → columnar summary; `go test -v` → PASS/FAIL counts + failures; `git log` → compact commit table; `go test -bench` → aligned bench table.
2. Unstructured output that no detector compressed, but which is byte-identical to something already emitted (this or an earlier session) → a compact `§ref:HASH§` token. Resolve with `qdf-hook expand HASH`.
3. Anything else → pass through unchanged (and registered so a later identical repeat becomes a `§ref`).

Dedup is content-addressed and runs *after* execution, so it can never suppress a side effect — a `§ref` is emitted only when the current output's hash already has a stored blob, so the blob always equals the current bytes (no staleness, works across sessions).

### Write (PostToolUse)
Replaces the echoed file content with `[WRITE §ref:HASH§ path — N lines written]` and caches the content so the next read of that file is served as a delta.

### Glob (PostToolUse)
Collapses a flat list of matched paths into an indented directory tree.

### Compact hooks
- `PreCompact`: marks `CompactedAt` in session state so next reads serve full content.
- `PostCompact`: injects a file-manifest into the fresh context ("you've read these files before").

## Expected savings

| Scenario | Before | After | Reduction |
|----------|--------|-------|-----------|
| Re-read unchanged 500-line file | 500 lines | 1 line | −99.8% |
| Re-read file with 5-line change | 500 lines | ~20 lines | −96% |
| 1000-row JSON array | ~60KB | ~300 chars | −99.5% |
| `go test -v` 50 tests all pass | 300 lines | 3 lines | −99% |
| Write a 500-line file | 500 lines | 1 line | −99.8% |
| Glob 50-file listing | 50 lines | ~15 lines | ~−70% |
| After compaction (3 files) | 1500 lines | 15 lines | −99% |

## Development

```bash
go test -race ./...
go test -bench=. -benchmem -count=6 ./...
gofmt -w .

# Profile a single hook invocation:
qdf-hook glob --cpuprofile=cpu.prof < input.json
go tool pprof -top cpu.prof
```
