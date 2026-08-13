package detect_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

// unfold reverses FoldLinePrefixes: every "^suffix" line is the nearest
// preceding "[^=prefix]" declaration plus its suffix. If this cannot rebuild
// the input byte-for-byte, the transform is losing content.
func unfold(t *testing.T, folded string) string {
	t.Helper()
	trailingNL := strings.HasSuffix(folded, "\n")
	var b strings.Builder
	prefix := ""
	for i, l := range strings.Split(strings.TrimRight(folded, "\n"), "\n") {
		switch {
		case strings.HasPrefix(l, "[^=") && strings.HasSuffix(l, "]"):
			prefix = l[3 : len(l)-1]
			continue
		case strings.HasPrefix(l, "^"):
			if prefix == "" {
				t.Fatalf("line %d references a prefix before any declaration: %q", i+1, l)
			}
			b.WriteString(prefix)
			b.WriteString(l[1:])
		default:
			b.WriteString(l)
		}
		b.WriteByte('\n')
	}
	out := b.String()
	if !trailingNL {
		out = strings.TrimSuffix(out, "\n")
	}
	return out
}

func TestFoldLinePrefixes_FoldsAndRoundTrips(t *testing.T) {
	in := "" +
		"/srv/app/internal/billing/charger.go:42:\tif err != nil {\n" +
		"/srv/app/internal/billing/receipt.go:17:\tif err != nil {\n" +
		"/srv/app/internal/billing/provider.go:88:\tif err != nil {\n" +
		"/srv/app/internal/router/mux.go:12:\treturn nil\n" +
		"/srv/app/internal/router/mux_test.go:31:\treturn nil\n"

	out := detect.FoldLinePrefixes(in)
	if out == "" {
		t.Fatal("a run of paths under one directory must fold")
	}
	if len(out) >= len(in) {
		t.Fatalf("folded form is not smaller: %d >= %d", len(out), len(in))
	}
	// The run extends greedily: it widens to admit the router lines, which
	// shortens the shared prefix to the directory both have in common. Refusing
	// to widen when the prefix shrinks — splitting into a billing run and a
	// router run — measures 0.2 points better over the real corpus, which does
	// not pay for a second rule in a transform that must stay easy to reason
	// about.
	if !strings.Contains(out, "[^=/srv/app/internal/]") {
		t.Errorf("the run's shared prefix was not hoisted:\n%s", out)
	}
	if strings.Count(out, "[^=") != 1 {
		t.Errorf("one run, one declaration; got:\n%s", out)
	}
	if got := unfold(t, out); got != in {
		t.Fatalf("round trip lost content:\n got:\n%s\nwant:\n%s", got, in)
	}
}

// Lines that share nothing with their neighbours must survive untouched, and
// the payload must still round-trip around them.
func TestFoldLinePrefixes_InterleavedLinesSurvive(t *testing.T) {
	in := "" +
		"running 3 checks\n" +
		"/srv/app/internal/billing/charger.go: ok\n" +
		"/srv/app/internal/billing/receipt.go: ok\n" +
		"/srv/app/internal/billing/provider.go: ok\n" +
		"all done\n"

	out := detect.FoldLinePrefixes(in)
	if out == "" {
		t.Fatal("expected a fold")
	}
	for _, must := range []string{"running 3 checks", "all done"} {
		if !strings.Contains(out, must) {
			t.Errorf("non-matching line %q was not preserved:\n%s", must, out)
		}
	}
	if got := unfold(t, out); got != in {
		t.Fatalf("round trip lost content:\n got:\n%s\nwant:\n%s", got, in)
	}
}

// A hoisted prefix must end on a separator. Cutting mid-filename produces a
// declaration that reads as a typo and suffixes that read as different files.
func TestFoldLinePrefixes_PrefixEndsOnABoundary(t *testing.T) {
	in := "" +
		"/srv/app/handlers/checkout_create.go:1:package handlers\n" +
		"/srv/app/handlers/checkout_cancel.go:1:package handlers\n" +
		"/srv/app/handlers/checkout_refund.go:1:package handlers\n"
	out := detect.FoldLinePrefixes(in)
	if out == "" {
		t.Fatal("expected a fold")
	}
	decl := out[strings.Index(out, "[^=")+3 : strings.IndexByte(out, ']')]
	if !strings.ContainsAny(decl[len(decl)-1:], "/ :\t.-_") {
		t.Errorf("prefix %q does not end on a boundary", decl)
	}
	if got := unfold(t, out); got != in {
		t.Fatalf("round trip lost content")
	}
}

// If the payload already contains something that reads as this transform's own
// output, folding on top of it makes original text indistinguishable from a
// substitution and reconstruction corrupts a line nobody touched.
func TestFoldLinePrefixes_RefusesItsOwnOutput(t *testing.T) {
	base := "" +
		"/srv/app/internal/billing/charger.go: ok\n" +
		"/srv/app/internal/billing/receipt.go: ok\n" +
		"/srv/app/internal/billing/provider.go: ok\n"
	for name, in := range map[string]string{
		"declaration present": base + "note: a [^=/tmp] marker appears verbatim\n",
		"caret line":          base + "^suffix looking line\n",
		"caret first line":    "^leading caret\n" + base,
	} {
		if out := detect.FoldLinePrefixes(in); out != "" {
			t.Errorf("%s must be refused, got:\n%s", name, out)
		}
	}
}

func TestFoldLinePrefixes_RefusesWhatItCannotHelp(t *testing.T) {
	for name, in := range map[string]string{
		"unrelated lines": "alpha\nbravo\ncharlie\ndelta\necho\nfoxtrot\n",
		"too few lines":   "/srv/app/internal/billing/charger.go\n",
		"empty":           "",
		"short prefixes":  "ab1\nab2\nab3\nab4\nab5\nab6\n",
	} {
		if out := detect.FoldLinePrefixes(in); out != "" {
			t.Errorf("%s must be refused, got:\n%s", name, out)
		}
	}
}

// The trailing newline is content: a payload that ended without one must not
// gain one, or a diff against the next run reports a change that never happened.
func TestFoldLinePrefixes_PreservesTrailingNewline(t *testing.T) {
	body := "" +
		"/srv/app/internal/billing/charger.go: ok\n" +
		"/srv/app/internal/billing/receipt.go: ok\n" +
		"/srv/app/internal/billing/provider.go: ok"
	out := detect.FoldLinePrefixes(body)
	if out == "" {
		t.Fatal("expected a fold")
	}
	if strings.HasSuffix(out, "\n") {
		t.Errorf("input had no trailing newline, output grew one:\n%q", out)
	}
	if got := unfold(t, out); got != body {
		t.Fatalf("round trip lost content:\n got %q\nwant %q", got, body)
	}
}
