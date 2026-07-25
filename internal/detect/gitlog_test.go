package detect_test

import (
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/alex60217101990/terse/internal/detect"
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

// TestSummarizeGitLog_BlankLineCount pins the fix for the header counting split
// lines (including interior blanks) instead of actual commits.
func TestSummarizeGitLog_BlankLineCount(t *testing.T) {
	in := "abc1234 first commit\n\ndef5678 second commit\n"
	s := detect.SummarizeGitLog(in)
	if !strings.Contains(s, "2 commits") {
		t.Errorf("expected '2 commits' despite interior blank line, got: %q", s)
	}
}

// TestSummarizeGitLog_UTF8Truncation ensures a long multi-byte message is not
// cut mid-rune (which would emit invalid UTF-8).
func TestSummarizeGitLog_UTF8Truncation(t *testing.T) {
	long := "abc1234 " + strings.Repeat("я", 100) // 100 Cyrillic runes (2 bytes each)
	s := detect.SummarizeGitLog(long)
	if !utf8.ValidString(s) {
		t.Errorf("truncated summary must stay valid UTF-8: %q", s)
	}
}
