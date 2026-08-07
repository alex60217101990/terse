package tokens_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/tokens"
)

func TestCountKnownStrings(t *testing.T) {
	// Hand-verified against the o200k_base encoding.
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"    ", 1},
	}
	for _, c := range cases {
		if got := tokens.Count(c.in); got != c.want {
			t.Errorf("Count(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// Any non-empty input costs at least one token. A zero would let a gate treat a
// non-empty summary as free.
func TestCountNonEmptyAtLeastOne(t *testing.T) {
	for _, s := range []string{"a", " ", "\n", "§", " "} {
		if got := tokens.Count(s); got < 1 {
			t.Errorf("Count(%q) = %d, want >= 1", s, got)
		}
	}
}

func TestCountMonotonic(t *testing.T) {
	short := "func foo() error { return nil }"
	long := short + short
	if tokens.Count(long) <= tokens.Count(short) {
		t.Fatal("doubling the input must increase the token count")
	}
}

// Machine-generated identifiers cost more per byte than prose. This ordering is
// the whole reason the counter exists: bytes saved are not tokens saved, and a
// gate that used bytes would happily trade 100 bytes of hex for 100 bytes of
// prose while quadrupling the bill.
func TestCountDenseCostsMorePerByte(t *testing.T) {
	prose := strings.Repeat("the quick brown fox jumps over it ", 20)
	for name, dense := range map[string]string{
		"hex":        strings.Repeat("a1b2c3d4e5f60718", 40),
		"timestamps": strings.Repeat("2026-08-07T17:48:21.123Z ", 30),
		"uuids":      strings.Repeat("3f2504e0-4f89-11d3-9a0c-0305e82c3301 ", 20),
		"base64":     strings.Repeat("SGVsbG8gd29ybGQgdGhpcyBpcyBiYXNlNjQ=", 20),
	} {
		if perByte(dense) <= perByte(prose) {
			t.Errorf("%s must cost more per byte than prose: %s=%.4f prose=%.4f",
				name, name, perByte(dense), perByte(prose))
		}
	}
}

// Non-ASCII text is CHEAPER per byte than English prose, not dearer. o200k has
// whole Cyrillic and CJK words in its vocabulary, and each of their characters
// is 2-3 UTF-8 bytes, so a token covers more bytes than it does in English.
//
// This is pinned because it is counter-intuitive and because the deleted
// character-class approximator got it backwards: it charged non-ASCII as dense,
// which is one of the places its decisions diverged from the encoder's. Anyone
// reintroducing a byte-class heuristic will trip over this test, which is the
// point.
func TestCountNonASCIIIsCheaperPerByteThanProse(t *testing.T) {
	prose := strings.Repeat("the quick brown fox jumps over it ", 20)
	for name, s := range map[string]string{
		"cyrillic": strings.Repeat("привет мир ", 60),
		"cjk":      strings.Repeat("这是一个测试字符串 ", 40),
	} {
		if perByte(s) >= perByte(prose) {
			t.Errorf("%s should cost less per byte than prose: %s=%.4f prose=%.4f",
				name, name, perByte(s), perByte(prose))
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

// BPE folds a single leading space into the following word (" world" is one
// token), so counting words in isolation systematically overcounts prose.
func TestCountSingleSpaceIsFree(t *testing.T) {
	joined := tokens.Count("alpha bravo charlie delta echo")
	separate := tokens.Count("alpha") + tokens.Count("bravo") +
		tokens.Count("charlie") + tokens.Count("delta") + tokens.Count("echo")
	if joined > separate {
		t.Errorf("spaces were charged: joined=%d > sum of words=%d", joined, separate)
	}
}

// BenchmarkCount puts the steady-state cost of the exact counter on record over
// the committed corpus, which is a deliberate mix of the shapes qdf-hook sees:
// logs, diffs, stack traces, JSON, paths, hex, CJK and emoji. The one-time
// vocabulary load is warmed out of the measurement — TestVocabularyLoadsLazily
// is where the first-call cost is pinned.
func BenchmarkCount(b *testing.B) {
	_, contents := loadCorpus(b)
	payload := string(bytes.Join(contents, []byte("\n")))

	tokens.Count("warm")
	b.SetBytes(int64(len(payload)))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = tokens.Count(payload)
	}
}
