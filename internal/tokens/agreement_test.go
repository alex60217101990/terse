package tokens_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/alex60217101990/terse/internal/tokens"
	"github.com/alex60217101990/terse/internal/tokens/bpe"
)

// loadCorpus returns every committed sample, plus each sample sliced into
// smaller windows. The windows matter: the picker compares summaries that are
// often only a few hundred bytes, so the fit has to be honest at that scale and
// not merely on whole files.
func loadCorpus(t *testing.T) []string {
	t.Helper()
	root := filepath.Join("testdata", "corpus")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read corpus dir: %v", err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(root, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		s := string(b)
		out = append(out, s)
		for start := 0; start+256 <= len(s); start += 512 {
			out = append(out, s[start:min(start+512, len(s))])
		}
	}
	if len(out) < 100 {
		t.Fatalf("corpus yielded %d samples, want >= 100", len(out))
	}
	return out
}

// The gate that actually matters. The picker only ever needs the SIGN of a
// comparison, so an approximator that is 20% off but never inverts an ordering
// is perfectly usable, while one that is 2% off but flips decisions is not.
func TestDecisionAgreement(t *testing.T) {
	corpus := loadCorpus(t)
	pairs, agree := 0, 0
	for i := 0; i+1 < len(corpus); i++ {
		a, b := corpus[i], corpus[i+1]
		ea, eb := bpe.Count(a), bpe.Count(b)
		if ea == eb {
			continue // no decision to make
		}
		pairs++
		if (tokens.Count(a) < tokens.Count(b)) == (ea < eb) {
			agree++
		}
	}
	if pairs == 0 {
		t.Fatal("no comparable pairs in corpus")
	}
	rate := float64(agree) / float64(pairs)
	t.Logf("decision agreement: %.4f over %d pairs", rate, pairs)
	if rate < 0.99 {
		t.Errorf("decision agreement %.4f < 0.99; widen the class model rather "+
			"than lowering this bar", rate)
	}
}

func TestErrorBounds(t *testing.T) {
	corpus := loadCorpus(t)
	errs := make([]float64, 0, len(corpus))
	for _, s := range corpus {
		exact := bpe.Count(s)
		if exact == 0 {
			continue
		}
		e := float64(tokens.Count(s)-exact) / float64(exact)
		if e < 0 {
			e = -e
		}
		errs = append(errs, e)
	}
	if len(errs) == 0 {
		t.Fatal("no countable samples in corpus")
	}
	sort.Float64s(errs)
	median := errs[len(errs)/2]
	p95 := errs[len(errs)*95/100]
	t.Logf("relative error: median=%.4f p95=%.4f over %d samples", median, p95, len(errs))
	if median > 0.05 {
		t.Errorf("median relative error %.4f > 0.05", median)
	}
	if p95 > 0.12 {
		t.Errorf("p95 relative error %.4f > 0.12", p95)
	}
}
