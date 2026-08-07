package bpe_test

import (
	"testing"

	"github.com/alex60217101990/terse/internal/tokens/bpe"
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
		if got := bpe.Count(c.in); got != c.want {
			t.Errorf("Count(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestCountMonotonic(t *testing.T) {
	short := "func foo() error { return nil }"
	long := short + short
	if bpe.Count(long) <= bpe.Count(short) {
		t.Fatal("doubling the input must increase the token count")
	}
}
