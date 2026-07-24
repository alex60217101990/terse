package hook

import (
	"strings"
	"testing"
)

// TestBuildGrepSummary_SingleBarePath is the B3 regression: a single
// files_with_matches result (one bare path, no "file:line:text" shape) must
// route to the glob-tree compressor, not fall into the grouped path and emit
// a bogus "[grep: 0 matches in 0 files]". Before the fix, lineCount 1 / parsed
// 0 made `parsed < lineCount/2` == `0 < 0` (false), skipping the tree.
func TestBuildGrepSummary_SingleBarePath(t *testing.T) {
	path := "/some/deep/nested/" + strings.Repeat("x", 300) + "/file.go"
	summary, action := buildGrepSummary(path)
	if action != "tree" {
		t.Fatalf("single bare path must route to tree, got action=%q summary=%q", action, summary)
	}
	if strings.Contains(summary, "0 matches") {
		t.Errorf("must not emit a bogus '0 matches' summary: %q", summary)
	}
}
