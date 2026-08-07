package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/alex60217101990/terse/internal/tokens"
	"github.com/alex60217101990/terse/internal/tokens/bpe"
)

// score is how good one weight set is. Agreement comes first because it is the
// only property the picker actually depends on: it needs the SIGN of a
// comparison, so a set that is 20% off but never inverts an ordering beats one
// that is 2% off but flips decisions.
type score struct {
	agreement float64 // fraction of ordered pairs whose sign matches the BPE
	median    float64 // relative error
	p95       float64
}

// better reports whether a beats b: agreement first, then p95, then median.
// The agreement comparison uses a small epsilon so that a change of one pair in
// several thousand does not outrank a real error improvement.
func (a score) better(b score) bool {
	const eps = 1e-4
	switch {
	case a.agreement > b.agreement+eps:
		return true
	case b.agreement > a.agreement+eps:
		return false
	case a.p95 != b.p95:
		return a.p95 < b.p95
	default:
		return a.median < b.median
	}
}

func (s score) String() string {
	return fmt.Sprintf("agreement=%.4f median=%.4f p95=%.4f", s.agreement, s.median, s.p95)
}

func runCalibrate(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: qdf-replay calibrate <corpus-dir>")
	}
	samples, err := loadCorpusDir(args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "corpus: %d samples from %s\n", len(samples), args[0])

	// Exact counts are the expensive part and never change, so pay for them
	// once instead of on every candidate weight set.
	exact := make([]int, len(samples))
	for i, s := range samples {
		exact[i] = bpe.Count(s)
	}

	w := tokens.Default()
	best := evaluate(samples, exact, w)
	fmt.Fprintf(os.Stderr, "start:  %v\n", best)

	// Coordinate descent over the real Count function. A closed-form least
	// squares fit would optimise a linear stand-in, but Count rounds up once per
	// run, and that rounding is exactly what a linear model cannot express.
	fields := []struct {
		get    func(*tokens.Weights) *int
		lo, hi int
	}{
		{func(w *tokens.Weights) *int { return &w.MbPerTokLower }, 500, 12000},
		{func(w *tokens.Weights) *int { return &w.MbPerTokUpper }, 500, 12000},
		{func(w *tokens.Weights) *int { return &w.MbPerTokDigit }, 500, 12000},
		{func(w *tokens.Weights) *int { return &w.MbPerTokSpace }, 500, 32000},
		{func(w *tokens.Weights) *int { return &w.MbPerTokPunct }, 300, 8000},
		{func(w *tokens.Weights) *int { return &w.MbPerTokHigh2 }, 300, 12000},
		{func(w *tokens.Weights) *int { return &w.MbPerTokHigh3 }, 300, 12000},
		{func(w *tokens.Weights) *int { return &w.MbPerTokHigh4 }, 300, 12000},
		{func(w *tokens.Weights) *int { return &w.MtPerNewline }, 100, 2000},
	}

	const passes = 6
	for pass := range passes {
		improved := false
		// Step sizes shrink each pass: a coarse sweep first, then refinement.
		steps := []float64{0.5, 0.75, 0.9, 1.1, 1.35, 2.0}
		if pass >= 2 {
			steps = []float64{0.9, 0.95, 0.98, 1.02, 1.05, 1.1}
		}
		if pass >= 4 {
			steps = []float64{0.98, 0.99, 1.01, 1.02}
		}
		for _, f := range fields {
			cur := *f.get(&w)
			for _, mul := range steps {
				cand := w
				v := clamp(int(float64(cur)*mul), f.lo, f.hi)
				if v == cur {
					continue
				}
				*f.get(&cand) = v
				if sc := evaluate(samples, exact, cand); sc.better(best) {
					best, w, improved = sc, cand, true
				}
			}
		}
		fmt.Fprintf(os.Stderr, "pass %d: %v\n", pass, best)
		if !improved {
			break
		}
	}

	fmt.Fprintf(os.Stderr, "\nfinal:  %v\n", best)
	if best.agreement < 0.99 {
		fmt.Fprintf(os.Stderr,
			"\nWARNING: decision agreement %.4f is below the 0.99 gate.\n"+
				"The class model is too coarse for this corpus — add a class it\n"+
				"conflates (for example split '_'/'-'/'.' out of punctuation, or\n"+
				"split 2-byte from 3-byte UTF-8 runs). Do NOT lower the gate.\n",
			best.agreement)
	}

	printWeights(w, args[0], len(samples), best)
	return nil
}

func evaluate(samples []string, exact []int, w tokens.Weights) score {
	got := make([]int, len(samples))
	for i, s := range samples {
		got[i] = tokens.CountWith(s, w)
	}

	pairs, agree := 0, 0
	for i := 0; i+1 < len(samples); i++ {
		if exact[i] == exact[i+1] {
			continue // no decision to make
		}
		pairs++
		if (got[i] < got[i+1]) == (exact[i] < exact[i+1]) {
			agree++
		}
	}

	errs := make([]float64, 0, len(samples))
	for i := range samples {
		if exact[i] == 0 {
			continue
		}
		e := float64(got[i]-exact[i]) / float64(exact[i])
		if e < 0 {
			e = -e
		}
		errs = append(errs, e)
	}
	sort.Float64s(errs)

	var s score
	if pairs > 0 {
		s.agreement = float64(agree) / float64(pairs)
	}
	if len(errs) > 0 {
		s.median = errs[len(errs)/2]
		s.p95 = errs[min(len(errs)*95/100, len(errs)-1)]
	}
	return s
}

// loadCorpusDir mirrors the agreement test's loader: whole files plus sliced
// windows. The windows matter because the picker compares summaries only a few
// hundred bytes long, so a fit that is honest only on whole files is not honest
// where it is used.
func loadCorpusDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read corpus dir: %w", err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		s := string(b)
		out = append(out, s)
		for start := 0; start+256 <= len(s); start += 512 {
			out = append(out, s[start:min(start+512, len(s))])
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("corpus dir %s yielded no samples", dir)
	}
	return out, nil
}

func printWeights(w tokens.Weights, corpusDir string, n int, s score) {
	fmt.Printf(`package tokens

// Class weights used by Count.
//
// Fitted by `+"`qdf-replay calibrate "+corpusDir+"`"+` over %d samples
// (whole files plus 512-byte windows) against the exact o200k BPE.
// Achieved: %v.
//
// Do not hand-tune. Refit and paste, so every number stays traceable to a
// corpus a contributor can regenerate.
const (
	// Bytes per token within a run of the given class, in milli-bytes
	// (4400 == 4.4 bytes per token).
	mbPerTokLower = %d
	mbPerTokUpper = %d
	mbPerTokDigit = %d
	mbPerTokSpace = %d
	mbPerTokPunct = %d
	mbPerTokHigh2 = %d
	mbPerTokHigh3 = %d
	mbPerTokHigh4 = %d

	// Flat per-occurrence cost, in milli-tokens.
	mtPerNewline = %d
)
`, n, s,
		w.MbPerTokLower, w.MbPerTokUpper, w.MbPerTokDigit, w.MbPerTokSpace,
		w.MbPerTokPunct, w.MbPerTokHigh2, w.MbPerTokHigh3, w.MbPerTokHigh4,
		w.MtPerNewline)
}

func clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}
