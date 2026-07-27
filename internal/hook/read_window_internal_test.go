package hook

import (
	"strings"
	"testing"
)

func TestSliceLines(t *testing.T) {
	const s = "l1\nl2\nl3\nl4\nl5\n"
	cases := []struct {
		name     string
		s        string
		start, n int
		want     string
		wantOK   bool
	}{
		{"first-line", s, 1, 1, "l1\n", true},
		{"middle-window", s, 2, 3, "l2\nl3\nl4\n", true},
		{"last-line-trailing-nl", s, 5, 1, "l5\n", true},
		{"whole-file", s, 1, 5, s, true},
		{"past-eof-start", s, 6, 1, "", false},
		{"window-runs-past-eof", s, 4, 5, "", false},
		{"no-trailing-nl-last", "a\nb\nc", 3, 1, "c", true},
		{"no-trailing-nl-window", "a\nb\nc", 2, 2, "b\nc", true},
		{"no-trailing-nl-past", "a\nb\nc", 3, 2, "", false},
		{"single-line-no-nl", "only", 1, 1, "only", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := sliceLines(tc.s, tc.start, tc.n)
			if ok != tc.wantOK || got != tc.want {
				t.Errorf("sliceLines(%q,%d,%d) = (%q,%v), want (%q,%v)",
					tc.s, tc.start, tc.n, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

// BenchmarkSliceLines asserts the single-pass slice is allocation-free (it
// returns a subslice of the input, never materializing a []string).
func BenchmarkSliceLines(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("some line of moderately sized content here\n")
	}
	s := sb.String()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := sliceLines(s, 400, 200); !ok {
			b.Fatal("unexpected miss")
		}
	}
}
