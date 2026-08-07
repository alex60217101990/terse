// Command qdf-replay measures qdf-hook in tokens rather than bytes.
//
// It is a development and CI tool, deliberately excluded from .goreleaser.yaml:
// it links a multi-megabyte tokenizer vocabulary that must never reach a
// shipped binary.
//
// The reference tokenizer is o200k, a public proxy for Claude's tokenizer
// rather than the real thing. Every number here is a way to compare two builds
// of qdf-hook against each other, not a statement about production token cost.
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
	var err error
	switch os.Args[1] {
	case "calibrate":
		err = runCalibrate(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "qdf-replay:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `usage: qdf-replay <command> [flags]

  calibrate <corpus-dir>
        Fit internal/tokens weights against the exact BPE over the COMMITTED
        synthetic corpus, and print a Go const block to paste into
        internal/tokens/weights.go.

Tokens are counted with o200k, a proxy for Claude's tokenizer. Numbers compare
builds; they do not estimate production cost.
`)
}
