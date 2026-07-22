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
