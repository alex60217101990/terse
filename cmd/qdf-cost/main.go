// Command qdf-cost measures what the hook is worth in billed tokens.
//
// cmd/qdf-replay answers a different question: replaying an archive through the
// pipeline counts the tokens the hook removes from a tool result, with o200k as
// a stand-in for Claude's tokenizer. That number is deterministic and cheap, and
// it is not a price. Compressed tool output lands in the prompt prefix, where
// every later turn re-reads it at cache-read rates — so a payload the replay
// scores at 26% saved is worth 26% of a discounted line item, not of the bill.
//
// This command runs the real thing instead. Each task is driven twice through
// `claude -p --output-format stream-json`, once with QDF_OFF=1 and once without,
// and the two runs are compared on the numbers the API actually reports:
// input_tokens, cache_creation_input_tokens, cache_read_input_tokens,
// output_tokens, and total_cost_usd.
//
// Two things to know before trusting a run:
//
//   - It spends real tokens and real money, twice per task.
//   - A live session is not deterministic. The model may take a different route
//     through the same prompt on the two runs, and that difference can be larger
//     than the effect being measured. Write tasks that pin the tool sequence,
//     use --runs to see the spread, and treat a single pair as an anecdote.
package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `usage: qdf-cost run [flags] <tasks.json>

  run <tasks.json>
        Drive every task in the file through Claude Code twice — once with the
        hook disabled (QDF_OFF=1), once with it live — and report the billed
        token counts and cost of each run side by side.

        A task file is a JSON array of objects:

          [{"name": "read-daemon",
            "prompt": "Read internal/daemon/daemon.go and reply with its line count."}]

        --runs N      repeat every task N times (default 1); the report keeps
                      each run separate, because the spread is the point
        --model M     model to drive (default: whatever Claude Code is set to)
        --dir PATH    working directory for the session (default: the cwd)
        --json        emit the report as JSON on stdout

WARNING: every task costs real tokens twice over. Start with one small task.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "run":
		fs := flag.NewFlagSet("run", flag.ExitOnError)
		runs := fs.Int("runs", 1, "repeat every task N times")
		model := fs.String("model", "", "model to drive")
		dir := fs.String("dir", "", "working directory for the session")
		asJSON := fs.Bool("json", false, "emit the report as JSON")
		if err := fs.Parse(os.Args[2:]); err != nil {
			os.Exit(2)
		}
		if fs.NArg() != 1 {
			fmt.Fprint(os.Stderr, usage)
			os.Exit(2)
		}
		if err := runCmd(fs.Arg(0), *runs, *model, *dir, *asJSON); err != nil {
			fmt.Fprintf(os.Stderr, "qdf-cost: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
