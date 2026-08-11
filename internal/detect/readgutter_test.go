package detect_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

// catN renders count lines of cat -n output starting at start.
func catN(start, count int) string {
	var b strings.Builder
	for i := range count {
		fmt.Fprintf(&b, "%d\t\tif err := step%02d(ctx); err != nil {\n", start+i, i)
	}
	return b.String()
}

func TestThinLineNumbers_KeepsAnchorsDropsTheRest(t *testing.T) {
	out := detect.ThinLineNumbers(catN(1, 24))
	if out == "" {
		t.Fatal("consecutive cat -n content must be thinned")
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 24 {
		t.Fatalf("line count changed: %d, want 24", len(lines))
	}
	numbered := map[int]bool{}
	for i, l := range lines {
		if n, _, ok := gutter(l); ok {
			numbered[n] = true
			if n != i+1 {
				t.Errorf("line %d carries the wrong number %d", i+1, n)
			}
		}
	}
	// First, last, and every tenth.
	for _, want := range []int{1, 10, 20, 24} {
		if !numbered[want] {
			t.Errorf("line %d must keep its number as an anchor:\n%s", want, out)
		}
	}
	for _, unwanted := range []int{2, 7, 13, 21} {
		if numbered[unwanted] {
			t.Errorf("line %d must have lost its number:\n%s", unwanted, out)
		}
	}
}

// The whole point is that only the gutter goes. Every line's content must come
// back byte-for-byte, or this transform is corrupting files.
func TestThinLineNumbers_BodyIsUntouched(t *testing.T) {
	in := catN(1, 30)
	out := detect.ThinLineNumbers(in)
	if out == "" {
		t.Fatal("expected thinning")
	}
	strip := func(s string) []string {
		var bodies []string
		for _, l := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
			if _, body, ok := gutter(l); ok {
				bodies = append(bodies, body)
			} else {
				bodies = append(bodies, l)
			}
		}
		return bodies
	}
	before, after := strip(in), strip(out)
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("line %d body changed:\n before %q\n after  %q", i+1, before[i], after[i])
		}
	}
}

// A windowed read starts partway into the file. Anchors must be multiples of
// ten in the FILE's numbering, not the window's, or a cited line is off by the
// window offset.
func TestThinLineNumbers_WindowKeepsAbsoluteNumbers(t *testing.T) {
	out := detect.ThinLineNumbers(catN(137, 30))
	if out == "" {
		t.Fatal("expected thinning")
	}
	var anchors []int
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if n, _, ok := gutter(l); ok {
			anchors = append(anchors, n)
		}
	}
	if len(anchors) < 3 || anchors[0] != 137 || anchors[len(anchors)-1] != 166 {
		t.Fatalf("first and last line must stay numbered, got %v", anchors)
	}
	for _, n := range anchors[1 : len(anchors)-1] {
		if n%10 != 0 {
			t.Errorf("interior anchor %d is not a multiple of ten: %v", n, anchors)
		}
	}
}

// Content that merely looks numbered must be left alone. A TSV whose first
// column is an id passes a naive "digits then tab" test on every line, and
// mangling it would silently corrupt data the model is reading.
func TestThinLineNumbers_RefusesNonConsecutive(t *testing.T) {
	var b strings.Builder
	for _, id := range []int{100, 250, 251, 900, 901, 902, 1000, 1001, 1500, 2000, 2001, 2002} {
		fmt.Fprintf(&b, "%d\twidget-%d\tin stock\n", id, id)
	}
	if out := detect.ThinLineNumbers(b.String()); out != "" {
		t.Fatalf("a non-consecutive first column is not a gutter:\n%s", out)
	}
}

func TestThinLineNumbers_RefusesUngutteredAndShort(t *testing.T) {
	cases := map[string]string{
		"plain prose":     "the quick brown fox\njumped over\nthe lazy dog\n",
		"one missing":     "1\ta\n2\tb\n3\tc\n4\td\n5\te\n6\tf\n7\tg\n8\th\n9\ti\nno gutter here\n11\tk\n",
		"too few lines":   catN(1, 4),
		"empty":           "",
		"gutter-only tab": "\ta\n\tb\n\tc\n\td\n\te\n\tf\n\tg\n\th\n\ti\n\tj\n",
	}
	for name, in := range cases {
		if out := detect.ThinLineNumbers(in); out != "" {
			t.Errorf("%s must be refused, got:\n%s", name, out)
		}
	}
}

// gutter splits a "<number><tab>" prefix off a line.
func gutter(l string) (num int, body string, ok bool) {
	i := strings.IndexByte(l, '\t')
	if i <= 0 {
		return 0, l, false
	}
	for _, c := range l[:i] {
		if c < '0' || c > '9' {
			return 0, l, false
		}
	}
	n := 0
	for _, c := range l[:i] {
		n = n*10 + int(c-'0')
	}
	return n, l[i+1:], true
}

// The shape the run variant exists for: a header line no strict check tolerates,
// then a numbered dump, then another header and another dump.
func TestThinLineNumberRuns_ThinsEachRunAroundHeaders(t *testing.T) {
	in := "=== a.go ===\n" + catN(1, 24) + "=== b.go ===\n" + catN(1, 12)
	out := detect.ThinLineNumberRuns(in)
	if out == "" {
		t.Fatal("mixed header/listing content must be thinned")
	}
	if detect.ThinLineNumbers(in) != "" {
		t.Fatal("the strict variant is supposed to refuse this shape")
	}
	if strings.Count(out, "\n") != strings.Count(in, "\n") {
		t.Fatalf("line count changed:\n%s", out)
	}
	for _, want := range []string{"=== a.go ===", "=== b.go ===", "1\t", "10\t", "20\t", "24\t", "12\t"} {
		if !strings.Contains(out, want) {
			t.Errorf("%q must survive:\n%s", want, out)
		}
	}
	// Every body line comes back byte-for-byte, gutter or not.
	for _, l := range strings.Split(strings.TrimRight(in, "\n"), "\n") {
		_, body, _ := gutter(l)
		if !strings.Contains(out, body) {
			t.Errorf("body %q was corrupted:\n%s", body, out)
		}
	}
}

// A run shorter than the anchor interval is left alone: that is the guard that
// keeps a numeric TSV out of reach.
func TestThinLineNumberRuns_RefusesShortRunsAndPlainText(t *testing.T) {
	cases := map[string]string{
		"short run":  "header\n" + catN(1, 9) + "footer\n",
		"plain":      "the quick brown fox\njumped over\n",
		"empty":      "",
		"restarting": "1\ta\n1\tb\n1\tc\n1\td\n1\te\n1\tf\n1\tg\n1\th\n1\ti\n1\tj\n",
	}
	for name, in := range cases {
		if out := detect.ThinLineNumberRuns(in); out != "" {
			t.Errorf("%s must be refused, got:\n%s", name, out)
		}
	}
}

// Whole-payload listings are the strict variant's job, and the run variant must
// agree with it there — one shape, one answer.
func TestThinLineNumberRuns_MatchesStrictOnAWholeListing(t *testing.T) {
	in := catN(1, 40)
	if runs, strict := detect.ThinLineNumberRuns(in), detect.ThinLineNumbers(in); runs != strict {
		t.Errorf("run variant disagrees with strict:\n%s\n---\n%s", runs, strict)
	}
}
