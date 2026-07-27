package detect_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

func makeBigObject() string {
	var b strings.Builder
	b.WriteString(`{"service":"widget-api","version":"2.3.1","replicas":4,"debug":false,`)
	b.WriteString(`"description":"` + strings.Repeat("long descriptive text ", 40) + `",`)
	b.WriteString(`"items":[`)
	for i := range 30 {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"id":%d,"state":"ready","zone":"eu-%d"}`, i, i%3)
	}
	b.WriteString(`],"limits":{"cpu":"2","mem":"4Gi"}}`)
	return b.String()
}

func TestSummarizeJSONObject_Basic(t *testing.T) {
	in := makeBigObject()
	out := detect.SummarizeJSONObject(in)
	if out == "" || len(out) >= len(in) {
		t.Fatalf("expected shrink, got %d vs %d", len(out), len(in))
	}
	for _, must := range []string{`"widget-api"`, `"2.3.1"`, "replicas", "items", "30", "…"} {
		if !strings.Contains(out, must) {
			t.Errorf("missing %q:\n%s", must, out)
		}
	}
}

func TestSummarizeJSONObject_Negatives(t *testing.T) {
	for i, s := range []string{
		"", "plain text", `["array","not","object"]`,
		`{"small":"object"}`, // < 1KB
		`{"broken": tru`,     // invalid JSON (>1KB via padding below)
	} {
		if i == 4 {
			s += strings.Repeat(" ", 1100)
		}
		if out := detect.SummarizeJSONObject(s); out != "" {
			t.Errorf("case %d false positive:\n%s", i, out)
		}
	}
}

// TestSummarizeJSONObject_TrailingGarbage_Rejected is the fix-report
// regression for the json.Decoder.Decode-stops-after-first-value bug:
// {"a":1} followed by non-JSON trailing bytes must NOT be summarized (the
// garbage would otherwise be silently invisible in the output).
func TestSummarizeJSONObject_TrailingGarbage_Rejected(t *testing.T) {
	content := `{"a":1}` + strings.Repeat("x", 2000)
	if out := detect.SummarizeJSONObject(content); out != "" {
		t.Fatalf("trailing garbage after a valid object must be rejected, got:\n%s", out)
	}
}

// TestSummarizeJSONObject_TrailingJSONValue_Rejected covers the same
// under-consumption bug for trailing bytes that also happen to be valid
// JSON: {"a":1}{"b":2} must not be treated as a single object either.
func TestSummarizeJSONObject_TrailingJSONValue_Rejected(t *testing.T) {
	content := strings.Repeat(`{"a":1}`, 300)
	if out := detect.SummarizeJSONObject(content); out != "" {
		t.Fatalf("a second top-level JSON value after the first must be rejected, got:\n%s", out)
	}
}

// TestSummarizeJSONObject_Branches directly exercises three shipped
// rendering branches that had no dedicated coverage: nested objects over the
// inline-key cap, short arrays of objects, and arrays of non-objects — all
// of which must fall back to the plain "array[N]" / "object{K keys}" forms.
func TestSummarizeJSONObject_Branches(t *testing.T) {
	// pad injects a "_filler" key so body clears the 1024-byte guard without
	// affecting the assertions below (they check for other keys' rendering).
	pad := func(body string) string {
		if len(body) >= 1200 {
			return body
		}
		filler := `,"_filler":"` + strings.Repeat("f", 1200-len(body)) + `"`
		return body[:len(body)-1] + filler + "}"
	}

	tests := []struct {
		name string
		in   string
		want []string
		not  []string
	}{
		{
			name: "nested object over inline cap collapses to object{K keys}",
			in:   pad(`{"config":{"a":1,"b":2,"c":3,"d":4,"e":5}}`),
			want: []string{"config: object{5 keys}"},
			not:  []string{"config: {a:1"},
		},
		{
			name: "short array of objects falls back to array[N]",
			in:   pad(`{"items":[{"a":1},{"a":2},{"a":3}]}`),
			want: []string{"items: array[3]"},
			not:  []string{"array[3] — {"},
		},
		{
			name: "array of non-objects falls back to array[N] regardless of length",
			in:   pad(`{"nums":[1,2,3,4,5,6,7]}`),
			want: []string{"nums: array[7]"},
			not:  []string{"array[7] — {"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := detect.SummarizeJSONObject(tt.in)
			if out == "" {
				t.Fatalf("expected a summary, got no match/no win for:\n%s", tt.in)
			}
			for _, w := range tt.want {
				if !strings.Contains(out, w) {
					t.Errorf("missing %q:\n%s", w, out)
				}
			}
			for _, n := range tt.not {
				if strings.Contains(out, n) {
					t.Errorf("unexpected %q:\n%s", n, out)
				}
			}
		})
	}
}

// TestSummarizeJSONObject_ManyTopLevelKeys covers the >200-key bound: only
// the first jsonObjectMaxKeys keys are listed, with a "… +K more keys"
// trailer for the rest, while the header still reports the true total.
func TestSummarizeJSONObject_ManyTopLevelKeys(t *testing.T) {
	const total = 250
	var b strings.Builder
	b.WriteByte('{')
	for i := range total {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `"k%03d":%d`, i, i)
	}
	b.WriteByte('}')
	content := b.String()

	out := detect.SummarizeJSONObject(content)
	if out == "" {
		t.Fatalf("expected a summary for %d top-level keys", total)
	}
	if !strings.Contains(out, fmt.Sprintf("[JSON OBJECT — %d top-level keys]", total)) {
		t.Errorf("expected header with true key count, got:\n%s", out)
	}
	if !strings.Contains(out, "k199: 199") {
		t.Errorf("expected the 200th key (k199) to be shown, got:\n%s", out)
	}
	if strings.Contains(out, "k200: 200") {
		t.Errorf("expected keys beyond the 200 cap to be elided, got:\n%s", out)
	}
	if !strings.Contains(out, "… +50 more keys") {
		t.Errorf("expected the elided-keys trailer, got:\n%s", out)
	}
}

func TestSummarizeJSONObject_NonMatchZeroAlloc(t *testing.T) {
	in := strings.Repeat("prose line\n", 200)
	if n := testing.AllocsPerRun(20, func() { detect.SummarizeJSONObject(in) }); n != 0 {
		t.Fatalf("non-match allocated %.0f", n)
	}
}

func BenchmarkSummarizeJSONObject_Match(b *testing.B) {
	in := makeBigObject()
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeJSONObject(in)
	}
}

func BenchmarkSummarizeJSONObject_NonMatch(b *testing.B) {
	in := strings.Repeat("prose line\n", 200)
	b.ReportAllocs()
	b.SetBytes(int64(len(in)))
	for b.Loop() {
		_ = detect.SummarizeJSONObject(in)
	}
}
