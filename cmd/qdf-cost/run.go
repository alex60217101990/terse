package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Task is one prompt to drive. Keep prompts explicit about which tools to use:
// the comparison only means something when both runs take the same route.
type Task struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

// Usage is the subset of the result message's usage object that is billed.
type Usage struct {
	InputTokens         int `json:"input_tokens"`
	CacheCreationTokens int `json:"cache_creation_input_tokens"`
	CacheReadTokens     int `json:"cache_read_input_tokens"`
	OutputTokens        int `json:"output_tokens"`
}

// Billed is what the whole session cost, in tokens. Cache writes and cache reads
// are kept apart from plain input because they are priced differently — folding
// them into one "input tokens" figure is what makes a token count stop tracking
// the bill.
func (u Usage) Billed() int {
	return u.InputTokens + u.CacheCreationTokens + u.CacheReadTokens + u.OutputTokens
}

// Result is the final `{"type":"result"}` line of a stream-json session.
type Result struct {
	Type         string  `json:"type"`
	Subtype      string  `json:"subtype"`
	IsError      bool    `json:"is_error"`
	NumTurns     int     `json:"num_turns"`
	DurationMS   int     `json:"duration_ms"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	Usage        Usage   `json:"usage"`
}

// Run is one task driven once under one variant.
type Run struct {
	Task    string `json:"task"`
	Variant string `json:"variant"` // "off" or "on"
	Attempt int    `json:"attempt"`
	Result  Result `json:"result"`
}

// Report is every run of an invocation, in the order they were driven.
type Report struct {
	Model string `json:"model"`
	Dir   string `json:"dir"`
	Runs  []Run  `json:"runs"`
}

const (
	variantOff = "off"
	variantOn  = "on"
)

// opts is one invocation's configuration, threaded through rather than passed as
// a widening parameter list.
type opts struct {
	runs    int
	model   string
	dir     string
	allowed []string
	timeout time.Duration
	asJSON  bool
}

func runCmd(tasksPath string, o opts) error {
	if o.runs < 1 {
		return fmt.Errorf("--runs must be at least 1")
	}
	tasks, err := loadTasks(tasksPath)
	if err != nil {
		return err
	}

	rep := Report{Model: o.model, Dir: o.dir}
	for attempt := 1; attempt <= o.runs; attempt++ {
		for _, task := range tasks {
			// Baseline first, every time. The order matters for the same reason
			// it matters on a laptop benchmark: the second run of a pair reads a
			// warmer cache, so a fixed order keeps that bias on one side where
			// it can be reasoned about, instead of alternating it into noise.
			for _, variant := range []string{variantOff, variantOn} {
				res, err := drive(task, variant, o)
				if err != nil {
					return fmt.Errorf("%s/%s attempt %d: %w", task.Name, variant, attempt, err)
				}
				rep.Runs = append(rep.Runs, Run{
					Task: task.Name, Variant: variant, Attempt: attempt, Result: res,
				})
				if !o.asJSON {
					fmt.Fprintf(os.Stderr, "  ran %-24s %-3s attempt %d — %d turns, $%.4f\n",
						task.Name, variant, attempt, res.NumTurns, res.TotalCostUSD)
				}
			}
		}
	}

	if o.asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	}
	writeReport(os.Stdout, rep)
	return nil
}

func loadTasks(path string) ([]Task, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tasks: %w", err)
	}
	var tasks []Task
	if err := json.Unmarshal(raw, &tasks); err != nil {
		return nil, fmt.Errorf("decode tasks: %w", err)
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("%s holds no tasks", path)
	}
	for i, t := range tasks {
		if t.Name == "" || t.Prompt == "" {
			return nil, fmt.Errorf("task %d needs both a name and a prompt", i)
		}
	}
	return tasks, nil
}

// drive runs one task under one variant and returns the session's result line.
//
// o.allowed is the pre-approved tool list and o.timeout bounds one session. Both
// exist for the same reason: an unattended sweep is 2 runs per task per attempt,
// and a single session stopping to ask for permission would otherwise hang the
// whole thing with no output to show for the runs that already cost money.
func drive(task Task, variant string, o opts) (Result, error) {
	args := claudeArgs(o.model, o.allowed)

	ctx, cancel := context.WithTimeout(context.Background(), o.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "claude", args...)
	// The prompt goes in on stdin, not as a trailing argument: --allowedTools and
	// --model are variadic, so a positional prompt is read as one more value for
	// whichever flag came last, and the session dies asking for input.
	cmd.Stdin = strings.NewReader(task.Prompt)
	cmd.Dir = o.dir
	cmd.Env = os.Environ()
	if variant == variantOff {
		cmd.Env = append(cmd.Env, "QDF_OFF=1")
	}
	// Note what this does NOT control: which qdf-hook actually runs. See the
	// package comment — that is settled by the installed settings and the
	// resident daemon, not by anything this process can pass down.
	// stderr is left attached: a session that stalls on a permission prompt or
	// dies on an auth error should say so on the terminal rather than time out
	// silently behind a pipe.
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return Result{}, fmt.Errorf("claude: %w", err)
	}
	res, err := parseResult(out)
	if err != nil {
		return Result{}, err
	}
	if res.IsError {
		return res, fmt.Errorf("session ended in error (%s)", res.Subtype)
	}
	return res, nil
}

// claudeArgs builds the CLI invocation. The prompt is deliberately absent: it
// travels on stdin, because --allowedTools and --model are variadic and would
// read a trailing positional prompt as one more value.
func claudeArgs(model string, allowed []string) []string {
	args := []string{"-p", "--output-format", "stream-json", "--verbose"}
	if model != "" {
		args = append(args, "--model", model)
	}
	if len(allowed) > 0 {
		args = append(args, "--allowedTools")
		args = append(args, allowed...)
	}
	return args
}

// parseResult pulls the final result line out of a stream-json transcript.
//
// It is the last line whose type is "result" AND which carries usage: the
// stream also contains short `"type":"result"` records for hook invocations,
// and those have no usage object to read.
func parseResult(stream []byte) (Result, error) {
	sc := bufio.NewScanner(bytes.NewReader(stream))
	// Assistant messages carry whole tool results; a 64KB default would split
	// them and turn every long line into a decode error.
	sc.Buffer(make([]byte, 0, 1<<20), 16<<20)

	var found Result
	var ok bool
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		if !bytes.Contains(line, []byte(`"type":"result"`)) {
			continue
		}
		var res Result
		if err := json.Unmarshal(line, &res); err != nil {
			continue
		}
		if res.Usage.Billed() == 0 {
			continue
		}
		found, ok = res, true
	}
	if err := sc.Err(); err != nil {
		return Result{}, fmt.Errorf("scan stream: %w", err)
	}
	if !ok {
		return Result{}, fmt.Errorf("no result line with usage in the transcript")
	}
	return found, nil
}

func writeReport(w *os.File, rep Report) {
	bw := bufio.NewWriter(w)
	defer func() { _ = bw.Flush() }()

	fmt.Fprintf(bw, "\n%-24s %-4s %-4s %8s %10s %10s %8s %10s\n",
		"TASK", "RUN", "HOOK", "IN", "CACHE-W", "CACHE-R", "OUT", "USD")

	var offTotal, onTotal Usage
	var offUSD, onUSD float64
	for _, pair := range pairs(rep.Runs) {
		for _, r := range []Run{pair.off, pair.on} {
			u := r.Result.Usage
			fmt.Fprintf(bw, "%-24s %-4d %-4s %8d %10d %10d %8d %10.4f\n",
				r.Task, r.Attempt, r.Variant,
				u.InputTokens, u.CacheCreationTokens, u.CacheReadTokens, u.OutputTokens,
				r.Result.TotalCostUSD)
		}
		fmt.Fprintf(bw, "%-24s %-4s %-4s %8s %10s %10s %8s %10s\n",
			"", "", "Δ",
			pct(pair.off.Result.Usage.InputTokens, pair.on.Result.Usage.InputTokens),
			pct(pair.off.Result.Usage.CacheCreationTokens, pair.on.Result.Usage.CacheCreationTokens),
			pct(pair.off.Result.Usage.CacheReadTokens, pair.on.Result.Usage.CacheReadTokens),
			pct(pair.off.Result.Usage.OutputTokens, pair.on.Result.Usage.OutputTokens),
			pctF(pair.off.Result.TotalCostUSD, pair.on.Result.TotalCostUSD))

		offTotal = add(offTotal, pair.off.Result.Usage)
		onTotal = add(onTotal, pair.on.Result.Usage)
		offUSD += pair.off.Result.TotalCostUSD
		onUSD += pair.on.Result.TotalCostUSD
	}

	fmt.Fprintf(bw, "\n%-34s %8d %10d %10d %8d %10.4f\n", "TOTAL off",
		offTotal.InputTokens, offTotal.CacheCreationTokens, offTotal.CacheReadTokens,
		offTotal.OutputTokens, offUSD)
	fmt.Fprintf(bw, "%-34s %8d %10d %10d %8d %10.4f\n", "TOTAL on",
		onTotal.InputTokens, onTotal.CacheCreationTokens, onTotal.CacheReadTokens,
		onTotal.OutputTokens, onUSD)
	fmt.Fprintf(bw, "%-34s %8s %10s %10s %8s %10s\n", "TOTAL Δ",
		pct(offTotal.InputTokens, onTotal.InputTokens),
		pct(offTotal.CacheCreationTokens, onTotal.CacheCreationTokens),
		pct(offTotal.CacheReadTokens, onTotal.CacheReadTokens),
		pct(offTotal.OutputTokens, onTotal.OutputTokens),
		pctF(offUSD, onUSD))

	fmt.Fprint(bw, "\nbilled tokens and cost as reported by the API, not an o200k estimate.\n"+
		"a live session is not deterministic: read the per-run spread before the total.\n")
}

// pair is one task's baseline and treatment run for a single attempt.
type pair struct{ off, on Run }

func pairs(runs []Run) []pair {
	var out []pair
	byKey := map[string]Run{}
	for _, r := range runs {
		key := fmt.Sprintf("%s/%d", r.Task, r.Attempt)
		if r.Variant == variantOff {
			byKey[key] = r
			continue
		}
		off, ok := byKey[key]
		if !ok {
			continue
		}
		out = append(out, pair{off: off, on: r})
		delete(byKey, key)
	}
	return out
}

func add(a, b Usage) Usage {
	return Usage{
		InputTokens:         a.InputTokens + b.InputTokens,
		CacheCreationTokens: a.CacheCreationTokens + b.CacheCreationTokens,
		CacheReadTokens:     a.CacheReadTokens + b.CacheReadTokens,
		OutputTokens:        a.OutputTokens + b.OutputTokens,
	}
}

// pct reports the treatment as a signed percentage change from the baseline.
// Negative is the hook saving tokens.
func pct(base, got int) string { return pctF(float64(base), float64(got)) }

func pctF(base, got float64) string {
	if base == 0 {
		if got == 0 {
			return "-"
		}
		return "new"
	}
	d := (got - base) / base * 100
	return strings.TrimSuffix(fmt.Sprintf("%+.1f%%", d), "\n")
}
