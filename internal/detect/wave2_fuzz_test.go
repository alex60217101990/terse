package detect_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

// Every wave-2 transform must (a) never panic, (b) never grow the input,
// (c) keep lossless transforms reversible in the properties they promise.
func fuzzNeverWorse(f *testing.F, fn func(string) string, emptyMeansNoWin bool) {
	f.Add("diff --git a/x b/x\n@@ -1 +1 @@\n-a\n+b\n")
	f.Add("CONTAINER ID   IMAGE\nabc   img:1\n")
	f.Add("Traceback (most recent call last):\n  File \"x\", line 1, in f\n    x\nE: boom\n")
	f.Add(strings.Repeat("/a/b/c/file.go\n", 9))
	f.Add("### hdr\nid=1 ok\n\n\n### hdr\nid=2 ok\n")
	f.Add("{\"k\":1}")
	f.Fuzz(func(t *testing.T, s string) {
		out := fn(s)
		if emptyMeansNoWin {
			if out != "" && len(out) >= len(s) {
				t.Fatalf("grew: %d -> %d", len(s), len(out))
			}
		} else if len(out) > len(s) {
			t.Fatalf("grew: %d -> %d", len(s), len(out))
		}
	})
}

func FuzzSummarizeGitDiff(f *testing.F)    { fuzzNeverWorse(f, detect.SummarizeGitDiff, true) }
func FuzzSummarizeTable(f *testing.F)      { fuzzNeverWorse(f, detect.SummarizeTable, true) }
func FuzzSummarizeStackTrace(f *testing.F) { fuzzNeverWorse(f, detect.SummarizeStackTrace, true) }
func FuzzSummarizeJSONObject(f *testing.F) { fuzzNeverWorse(f, detect.SummarizeJSONObject, true) }
func FuzzFoldRepeatedBlocks(f *testing.F)  { fuzzNeverWorse(f, detect.FoldRepeatedBlocks, false) }
func FuzzFoldPathPrefix(f *testing.F)      { fuzzNeverWorse(f, detect.FoldPathPrefix, false) }
func FuzzFoldLinePrefixes(f *testing.F)    { fuzzNeverWorse(f, detect.FoldLinePrefixes, true) }
func FuzzThinLineNumbers(f *testing.F)     { fuzzNeverWorse(f, detect.ThinLineNumbers, true) }
func FuzzThinLineNumberRuns(f *testing.F)  { fuzzNeverWorse(f, detect.ThinLineNumberRuns, true) }
