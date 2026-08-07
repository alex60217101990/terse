// Command qdf-replay measures qdf-hook against an archive of Claude Code
// transcripts: it replays recorded tool results through the hook and reports
// the token cost before and after, per hook and per action.
//
// It is a development and CI tool, not something a user installs. It is
// deliberately excluded from .goreleaser.yaml — that file builds only
// ./cmd/qdf-hook, ./cmd/qdf-hookd and ./cmd/qdf-hookc — because nothing here
// belongs on an end user's machine.
//
// Tokens are counted with o200k (internal/tokens), which is a public proxy for
// Claude's tokenizer rather than the real thing. Every number this tool prints
// is a way to compare two BUILDS of qdf-hook against a fixed corpus. It is not
// an estimate of production token cost, and must not be quoted as one.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "replay":
		if err := runReplay(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "qdf-replay:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: qdf-replay replay [--json] [--baseline FILE] <transcripts-dir>

  replay <transcripts-dir>
        Walk the directory for *.jsonl Claude Code transcripts, pair every
        tool_result with its tool_use, and replay each recorded result through
        the hook pipeline. Reports tokens in/out per hook and action.

        Each session replays against its own store rooted in a temporary
        directory, so a run never reads or writes the real ~/.qdf-hook and
        never appends to the real analytics.jsonl.

        --json            emit the report as JSON on stdout
        --baseline FILE   diff against a previously recorded JSON report.
                          Errors if the corpus fingerprints differ, and exits 1
                          if any category regresses. Tolerance is zero: the
                          replay is deterministic for a fixed corpus and binary.

Results that already carry qdf-hook markers are excluded from the corpus:
replaying them would double-compress and manufacture a fake win.

Tokens are counted with o200k, a proxy for Claude's tokenizer. Numbers compare
builds against a fixed corpus; they do not estimate production cost.
`)
}
