package detect_test

import (
	"os"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/detect"
)

func TestIsGoTestOutput(t *testing.T) {
	pass, _ := os.ReadFile("../../testdata/gotest_pass.txt")
	if !detect.IsGoTestOutput(string(pass)) {
		t.Error("gotest_pass.txt should be detected as go test output")
	}
	fail, _ := os.ReadFile("../../testdata/gotest_fail.txt")
	if !detect.IsGoTestOutput(string(fail)) {
		t.Error("gotest_fail.txt should be detected as go test output")
	}
	if detect.IsGoTestOutput("just some text") {
		t.Error("plain text should not be detected as go test output")
	}
}

func TestSummarizeGoTest_Pass(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/gotest_pass.txt")
	s := detect.SummarizeGoTest(string(data))
	if !strings.Contains(s, "PASS") {
		t.Errorf("summary should contain PASS: %s", s)
	}
	if len(s) > 200 {
		t.Errorf("pass summary too long: %d chars", len(s))
	}
}

func TestSummarizeGoTest_Fail(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/gotest_fail.txt")
	s := detect.SummarizeGoTest(string(data))
	if !strings.Contains(s, "FAIL") {
		t.Errorf("summary should contain FAIL: %s", s)
	}
	if !strings.Contains(s, "TestBar") {
		t.Errorf("summary should contain failing test name: %s", s)
	}
	// Failed output lines should be preserved
	if !strings.Contains(s, "bar_test.go:42") {
		t.Errorf("summary should contain failure location: %s", s)
	}
}

// TestSummarizeGoTest_MultiPackage guards the multi-package attribution fix:
// a run spanning several packages must be labeled by count, not misattributed
// to whichever package's summary line came last.
func TestSummarizeGoTest_MultiPackage(t *testing.T) {
	in := "ok  \tpkg/a\t0.10s\nok  \tpkg/b\t0.20s\nok  \tpkg/c\t0.30s\n"
	out := detect.SummarizeGoTest(in)
	if !strings.Contains(out, "3 packages") {
		t.Errorf("multi-package summary should say '3 packages', got: %q", out)
	}
	if strings.Contains(out, "pkg/c") {
		t.Errorf("must not misattribute the summary to the last package only: %q", out)
	}
}

// TestSummarizeGoTest_CountsIndentedSubtests is the B5 regression: indented
// subtest results (`    --- FAIL: T/sub`) must be tallied in the headline
// [N PASS, M FAIL], not silently folded into the parent's detail only.
func TestSummarizeGoTest_CountsIndentedSubtests(t *testing.T) {
	in := "=== RUN   TestX\n" +
		"=== RUN   TestX/sub\n" +
		"    --- FAIL: TestX/sub (0.00s)\n" +
		"        x_test.go:5: boom\n" +
		"--- FAIL: TestX (0.00s)\n" +
		"FAIL\n"
	s := detect.SummarizeGoTest(in)
	if !strings.Contains(s, "2 FAIL") {
		t.Errorf("expected the indented subtest failure to be counted (2 FAIL), got:\n%s", s)
	}
}
