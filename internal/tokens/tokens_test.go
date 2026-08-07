package tokens_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/tokens"
)

func TestCountEmpty(t *testing.T) {
	if got := tokens.Count(""); got != 0 {
		t.Errorf("Count(\"\") = %d, want 0", got)
	}
}

// Any non-empty input costs at least one token. A zero would let a gate treat a
// non-empty summary as free.
func TestCountNonEmptyAtLeastOne(t *testing.T) {
	for _, s := range []string{"a", " ", "\n", "§", " "} {
		if got := tokens.Count(s); got < 1 {
			t.Errorf("Count(%q) = %d, want >= 1", s, got)
		}
	}
}

// Dense classes must cost more per byte than prose. This ordering is the whole
// reason the counter exists: if it inverts, every picker decision built on it is
// wrong, and wrong in the direction of preferring output that costs more.
func TestCountDenseCostsMorePerByte(t *testing.T) {
	prose := strings.Repeat("the quick brown fox jumps over it ", 20)
	for name, dense := range map[string]string{
		"hex":        strings.Repeat("a1b2c3d4e5f60718", 40),
		"timestamps": strings.Repeat("2026-08-07T17:48:21.123Z ", 30),
		"non-ascii":  strings.Repeat("привет мир ", 60),
	} {
		if perByte(dense) <= perByte(prose) {
			t.Errorf("%s must cost more per byte than prose: %s=%.4f prose=%.4f",
				name, name, perByte(dense), perByte(prose))
		}
	}
}

func perByte(s string) float64 { return float64(tokens.Count(s)) / float64(len(s)) }

// Concatenation must not lose tokens: a gate comparing a whole against its
// parts would otherwise see a free lunch.
func TestCountMonotonicUnderConcatenation(t *testing.T) {
	a := "func handler(w http.ResponseWriter, r *http.Request) {\n"
	b := "\treturn fmt.Errorf(\"parse %q: %w\", name, err)\n"
	if tokens.Count(a+b) < tokens.Count(a) {
		t.Error("appending content reduced the count")
	}
	if tokens.Count(a+b) < tokens.Count(b) {
		t.Error("prepending content reduced the count")
	}
}

// A single space before a word is folded into the following token by BPE, so
// charging it separately would systematically overcount ordinary prose.
func TestCountSingleSpaceIsFree(t *testing.T) {
	joined := tokens.Count("alpha bravo charlie delta echo")
	separate := tokens.Count("alpha") + tokens.Count("bravo") +
		tokens.Count("charlie") + tokens.Count("delta") + tokens.Count("echo")
	if joined > separate {
		t.Errorf("spaces were charged: joined=%d > sum of words=%d", joined, separate)
	}
}

func TestCountZeroAlloc(t *testing.T) {
	s := strings.Repeat("func foo() error { return nil }\n", 100)
	if n := testing.AllocsPerRun(100, func() { _ = tokens.Count(s) }); n != 0 {
		t.Errorf("Count allocated %v times per run, want 0", n)
	}
}
