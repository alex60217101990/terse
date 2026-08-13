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
// Three things to know before trusting a run:
//
//   - It spends real tokens and real money, twice per task.
//   - A live session is not deterministic. The model may take a different route
//     through the same prompt on the two runs, and that difference can be larger
//     than the effect being measured. Write tasks that pin the tool sequence,
//     use --runs to see the spread, and treat a single pair as an anecdote.
//   - It measures the INSTALLED hook, not this working tree, and there is no
//     flag that changes that. `qdf-hook init` writes absolute binary paths into
//     settings.json, and PostToolUse hooks from --settings are ADDED to those
//     rather than replacing them, so a second hook does not displace the first.
//     Worse, any qdf-hook invoked while the resident daemon holds
//     ~/.qdf-hook/d.sock proxies the work to that daemon's binary — so even
//     running a freshly built binary by absolute path measures the daemon.
//     To measure a working tree: install it and restart the daemon. Skipping
//     that once cost a 24-session sweep that silently graded the last release.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const usage = `usage: qdf-cost run [flags] <tasks.json>

  run <tasks.json>
        Drive every task in the file through Claude Code twice — once with the
        hook disabled (QDF_OFF=1), once with it live — and report the billed
        token counts and cost of each run side by side.

        A task file is a JSON array of objects:

          [{"name": "read-daemon",
            "prompt": "Read internal/daemon/daemon.go and reply with its line count."}]

        --runs N            repeat every task N times (default 1); the report
                            keeps each run separate, because the spread is the
                            point
        --model M           model to drive (default: Claude Code's own setting)
        --dir PATH          working directory for the session (default: the cwd)
        --allowed-tools L   comma-separated tools to pre-approve, so an
                            unattended sweep never stalls on a permission prompt
WARNING: this measures the INSTALLED hook, never the working tree. See the
package comment for why there is no flag for that.
        --timeout D         per-session timeout (default 5m)
        --json              emit the report as JSON on stdout

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
		o := opts{}
		fs.IntVar(&o.runs, "runs", 1, "repeat every task N times")
		fs.StringVar(&o.model, "model", "", "model to drive")
		fs.StringVar(&o.dir, "dir", "", "working directory for the session")
		fs.DurationVar(&o.timeout, "timeout", 5*time.Minute, "per-session timeout")
		fs.BoolVar(&o.asJSON, "json", false, "emit the report as JSON")
		allowed := fs.String("allowed-tools", "",
			"comma-separated tools to pre-approve, so a sweep never stalls on a prompt")
		if err := fs.Parse(os.Args[2:]); err != nil {
			os.Exit(2)
		}
		if fs.NArg() != 1 {
			fmt.Fprint(os.Stderr, usage)
			os.Exit(2)
		}
		for t := range strings.SplitSeq(*allowed, ",") {
			if t = strings.TrimSpace(t); t != "" {
				o.allowed = append(o.allowed, t)
			}
		}
		if err := runCmd(fs.Arg(0), o); err != nil {
			fmt.Fprintf(os.Stderr, "qdf-cost: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
}
