# qdf-hook

Claude Code hook binary that reduces token consumption by 50–99% on:
- **Repeated file reads** — returns `§unchanged§` or unified diff instead of full content
- **Structured Bash output** — collapses JSON arrays, `go test -v`, `git log`, benchmarks into compact summaries
- **Post-compaction re-reads** — injects a file-read manifest after compaction so Claude doesn't start from scratch

## Install

```bash
go install github.com/alex60217101990/qdf-hook/cmd/qdf-hook@latest
```

Or from source:
```bash
git clone https://github.com/alex60217101990/qdf-hook
cd qdf-hook
go build -o $(go env GOPATH)/bin/qdf-hook ./cmd/qdf-hook/
```

## Configure Claude Code

Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [],
    "PostToolUse": [
      {
        "matcher": "Read|Glob|Grep",
        "hooks": [
          {
            "type": "command",
            "command": "qdf-hook read"
          }
        ]
      },
      {
        "matcher": "Bash|PowerShell",
        "hooks": [
          {
            "type": "command",
            "command": "qdf-hook bash"
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "qdf-hook precompact"
          }
        ]
      }
    ],
    "PostCompact": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "qdf-hook postcompact"
          }
        ]
      }
    ]
  }
}
```

> **Note:** If you already have sqz wired for Bash, put `qdf-hook bash` AFTER sqz in the Bash hooks array, or replace sqz with qdf-hook (qdf-hook's Bash handler is a superset of sqz for structured data detection).

## State files

State is stored per-session in `~/.qdf-hook/sessions/{session_id}.qdf` (qdf-compressed). Safe to delete; qdf-hook recreates on next session.

## How it works

### Read hook
1. SHA-256 the file content.
2. If not seen before: cache and return full content.
3. If seen and unchanged: return `[§unchanged:HASH§ path — read at turn N]` (1 line instead of N lines).
4. If seen and changed: compute Myers unified diff, return `[§delta:HASH§ path]` + diff only.
5. After compaction: treat as "not seen" — serve full content on next read.

### Bash hook
Tries detectors in order:
1. JSON array of objects → columnar summary (schema + per-column stats)
2. `go test -v` output → PASS/FAIL count + failure details only
3. `git log` output → compact commit table
4. `go test -bench` → aligned benchmark table
5. Anything else → pass through unchanged

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
| `git log` 20 commits | 20 lines | 22 lines | ~same |
| After compaction (3 files) | 1500 lines | 15 lines | −99% |

## Development

```bash
go test -race ./...
go test -bench=. -benchmem -count=10 ./...
gofmt -w .
```
