package detect_test

import (
	"os"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/detect"
)

func TestIsGitLogOutput(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/gitlog.txt")
	if !detect.IsGitLogOutput(string(data)) {
		t.Error("gitlog.txt should be detected as git log output")
	}
	if detect.IsGitLogOutput("package main\nfunc foo() {}") {
		t.Error("Go source should not be detected as git log")
	}
}

func TestSummarizeGitLog(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/gitlog.txt")
	s := detect.SummarizeGitLog(string(data))
	if !strings.Contains(s, "feat") && !strings.Contains(s, "perf") {
		t.Errorf("git log summary should contain commit types: %s", s)
	}
	// Should be shorter than source
	if len(s) >= len(data) {
		t.Errorf("summary should be shorter than raw log: %d >= %d", len(s), len(data))
	}
}
