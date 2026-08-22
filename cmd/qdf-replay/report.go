package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
)

// errRegression is returned when a --baseline comparison finds a category that
// got worse. main turns it into exit status 1.
var errRegression = errors.New("replay regressed against the baseline")

func runReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "emit the report as JSON")
	baseline := fs.String("baseline", "", "diff against a previously recorded JSON report")
	fs.Usage = usage
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		usage()
		return errors.New("replay needs exactly one transcripts directory")
	}

	corpus, err := LoadSessions(fs.Arg(0))
	if err != nil {
		return err
	}
	rep, err := Replay(corpus.Sessions)
	if err != nil {
		return err
	}
	rep.Fingerprint = corpus.Fingerprint
	rep.Skipped = corpus.Skipped
	rep.Filtered = corpus.Filtered
	rep.Unpaired = corpus.Unpaired

	if *jsonOut {
		if err := writeJSON(os.Stdout, rep); err != nil {
			return err
		}
	} else {
		printReport(os.Stdout, rep)
	}

	if *baseline == "" {
		return nil
	}
	base, err := readReport(*baseline)
	if err != nil {
		return err
	}
	return compareBaseline(os.Stderr, base, rep)
}

func writeJSON(w io.Writer, rep Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(rep)
}

func readReport(path string) (Report, error) {
	var rep Report
	data, err := os.ReadFile(path)
	if err != nil {
		return rep, fmt.Errorf("baseline: %w", err)
	}
	if err := json.Unmarshal(data, &rep); err != nil {
		return rep, fmt.Errorf("baseline %s: %w", path, err)
	}
	return rep, nil
}

// printReport renders the human table.
//
// Skipped/filtered/unpaired are printed next to the totals rather than tucked
// away, because they are how a run over a damaged archive announces itself. A
// report that only showed savings would make a corpus that half-failed to load
// look like a corpus that compressed unusually well.
func printReport(w io.Writer, rep Report) {
	fmt.Fprintf(w, "corpus %s  %d files, %s, sessions=%d triples=%d\n",
		shortHash(rep.Fingerprint.Hash), rep.Fingerprint.Files,
		humanBytes(rep.Fingerprint.TotalBytes), rep.Sessions, rep.Triples)
	fmt.Fprintf(w, "dropped: %d malformed lines, %d already-compressed results, %d unpaired results\n\n",
		rep.Skipped, rep.Filtered, rep.Unpaired)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "HOOK/ACTION\tN\tTOKENS IN\tTOKENS OUT\tSAVED\tSAVED %")
	for _, k := range sortedKeys(rep.ByHookAction) {
		c := rep.ByHookAction[k]
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\t%.1f%%\n", k, c.N, c.TokensIn, c.TokensOut, c.Saved(), c.SavedPct())
	}
	fmt.Fprintf(tw, "TOTAL\t%d\t%d\t%d\t%d\t%.1f%%\n",
		rep.Total.N, rep.Total.TokensIn, rep.Total.TokensOut, rep.Total.Saved(), rep.Total.SavedPct())
	_ = tw.Flush()

	fmt.Fprintln(w, "\ntokens counted with o200k, a proxy for Claude's tokenizer:")
	fmt.Fprintln(w, "these numbers compare builds against a fixed corpus, they do not estimate production cost.")
}

// compareBaseline diffs cur against base and reports every category that got
// worse.
//
// Tolerance is zero. The replay is deterministic for a fixed corpus and a fixed
// binary — identical inputs, no clock, no concurrency in the ledger — so any
// movement at all is a real behavior change, and a tolerance band would only
// hide the small regressions that accumulate.
func compareBaseline(w io.Writer, base, cur Report) error {
	if base.Fingerprint.Hash != cur.Fingerprint.Hash {
		return fmt.Errorf(
			"corpus fingerprints differ: baseline %s (%d files, %s) vs current %s (%d files, %s) — "+
				"the two runs are not comparable; re-record the baseline",
			shortHash(base.Fingerprint.Hash), base.Fingerprint.Files, humanBytes(base.Fingerprint.TotalBytes),
			shortHash(cur.Fingerprint.Hash), cur.Fingerprint.Files, humanBytes(cur.Fingerprint.TotalBytes))
	}

	type row struct {
		key       string
		why       string
		base, cur Cell
	}
	var bad []row

	check := func(key string, b, c Cell) {
		switch {
		case c.TokensOut > b.TokensOut:
			bad = append(bad, row{key, "tokens out rose", b, c})
		case c.SavedPct() < b.SavedPct():
			bad = append(bad, row{key, "saved % fell", b, c})
		}
	}

	for _, k := range sortedKeys(base.ByHookAction) {
		b := base.ByHookAction[k]
		c, ok := cur.ByHookAction[k]
		if !ok {
			// Same corpus, so the triples still exist — they were routed
			// somewhere else. That is a behavior change the gate must not
			// swallow just because the total happened to improve.
			bad = append(bad, row{k, "category disappeared", b, Cell{}})
			continue
		}
		check(k, b, c)
	}
	check("TOTAL", base.Total, cur.Total)

	if len(bad) == 0 {
		fmt.Fprintf(w, "no regression against baseline (%d categories, total %d -> %d tokens out)\n",
			len(base.ByHookAction), base.Total.TokensOut, cur.Total.TokensOut)
		return nil
	}

	fmt.Fprintln(w, "\nREGRESSIONS:")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "HOOK/ACTION\tWHY\tBASE OUT\tCUR OUT\tBASE SAVED %\tCUR SAVED %")
	for _, r := range bad {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%.1f%%\t%.1f%%\n",
			r.key, r.why, r.base.TokensOut, r.cur.TokensOut, r.base.SavedPct(), r.cur.SavedPct())
	}
	_ = tw.Flush()
	return errRegression
}

func sortedKeys(m map[string]Cell) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func shortHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	if h == "" {
		return "(none)"
	}
	return h
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n/div >= unit && exp < 3 {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}
