# qdf-hook — Universal tool pipeline + catch-all dispatch (design)

Date: 2026-07-23
Status: approved (brainstorm)

## Problem

Handlers are hardcoded per tool (Bash, Glob, Grep, Read, Write). Anything not
listed — MCP tools (`mcp__*`), future tools — gets **no compression**. But the
core savers (noise-strip, §ref dedup, re-run delta, squeeze, and the
content-sniffed structural summaries) are **tool-agnostic**: they work on any
textual output. We're leaving those savings on the table for every non-hardcoded
tool.

## Goal

Route **every** tool through one generic pipeline so new tools (incl. MCP) get
dedup + delta + noise-strip + squeeze + auto structural summaries for free — no
per-tool hardcoding. Keep the genuinely file-specific logic (Read, Write) as-is.

## Design

### Dispatch

New subcommand `qdf-hook post` reads the PostToolUse payload and routes by
`tool_name`:

- `Read` → existing read handler (mtime/size, file-content delta, compaction).
- `Write|Edit|MultiEdit` → existing write handler (caches real file bytes).
- **everything else** → `handleGeneric`.

`init` installs a single catch-all `PostToolUse` matcher `.*` → `qdf-hook post`
(plus the unchanged `PreToolUse Read → pretooluse` and Pre/PostCompact). Four
hooks instead of eight; MCP and future tools covered automatically. The old
per-tool subcommands (`bash`/`glob`/`grep`/`read`/`write`) stay as thin
delegates for backward compatibility.

### handleGeneric(toolName, inp, w)

1. **Skip list** — if `toolName` is in the verbatim set, passthrough immediately.
   Default: `TodoWrite`, `ExitPlanMode` (structured/verbatim). Extendable via
   `QDF_SKIP_TOOLS` (comma-separated).
2. `content = StripNoise(inp.ToolResponse.Text())`.
3. `len(content) < 256` → passthrough (record `Hook=toolName`).
4. Compression chain, first win takes it (all never-worse):
   1. format-by-toolname: `Glob` → dir tree; `Grep` → grouped matches.
   2. content-sniffed structural: JSON array → columnar; `go test -v`; `git log`;
      `go test -bench`. (Fire for ANY tool whose output matches — e.g. an MCP
      tool returning a JSON array.)
   3. §ref dedup (byte-identical repeat).
   4. re-run delta vs previous output for this tool+input.
   5. squeeze (ANSI + line RLE, ≥10% win).
   6. passthrough.
5. Store the (stripped) output for next-run delta when the action is
   passthrough/squeezed/rerun-delta.
6. Record analytics with `Hook=toolName` — stats naturally show per-tool,
   including `mcp__*`.

### Generalized delta key

Replace the bash-specific `command+cwd` key with a tool-agnostic one:
`key = sha256(toolName + 0x00 + tool_input_raw)[:16]`. Re-running the same tool
with the same arguments (any tool) diffs against its previous output. Hashing
uses zero-copy `unsafe.String` views. The `bashlast` store becomes a generic
`lastout` store (`~/.qdf-hook/lastout/<key>.blob`) taking a precomputed key;
`BashLastGet/Put` become `LastOutputGet/Put(key, ...)`.

## Performance (measure-first, hard requirement)

Reuse the proven techniques: qdf `OptBalanced`, plain write, `WithNoCopy`
decode, zero-copy `unsafe.String` key hashing, single-pass string building, no
regexp on the hot path.

**String-alloc cleanup (this pass):** several read-only `[]byte(content)`
conversions copy the entire tool output — `tryJSON` (`AnalyzeJSONArray([]byte(
content),…)`) and the re-run delta (`UnifiedDiff([]byte(prev), []byte(content))`).
The callees only read, so replace with a zero-copy view via a new
`internal/bytesconv` package: `S2B(s) []byte` / `B2S(b) string` (documented
read-only, never-mutated, lifetime-bound). This drops a full-output copy per
structured/delta bash call. The dispatch adds one `tool_name` string compare +
(for generic tools) the same chain Bash already runs — no new per-call cost for
Bash/Glob/Grep beyond a switch. Parity: Bash/Glob/Grep output must be
byte-identical to the current handlers (pin with existing tests).

Benchmarks (interleaved benchstat, n≥12): dispatch overhead vs direct call
(target < 5 µs), generic JSON/gotest paths unchanged vs current.

## Testing

- Parity: existing bash/glob/grep/read/write tests pass unchanged through
  dispatch.
- Generic coverage: an `mcp__x` tool returning a JSON array → columnar summary;
  a repeated `mcp__x` identical output → §ref; a near-identical re-run → delta.
- Skip list: a `TodoWrite` payload → always passthrough.
- `init` installs the catch-all and is idempotent; old per-tool subcommands
  still work.

## Out of scope

base64/binary → §ref; giant-output truncate+ref; block-level cross-tool dedup.
Follow-up specs.
