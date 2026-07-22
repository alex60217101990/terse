package detect_test

import (
	"os"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/detect"
)

func TestIsGoBenchOutput(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/bench_output.txt")
	if !detect.IsGoBenchOutput(string(data)) {
		t.Error("bench_output.txt should be detected as benchmark output")
	}
}

func TestSummarizeGoBench(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/bench_output.txt")
	s := detect.SummarizeGoBench(string(data))
	if !strings.Contains(s, "ns/op") {
		t.Errorf("bench summary should contain ns/op: %s", s)
	}
	if !strings.Contains(s, "Benchmark") {
		t.Errorf("bench summary should contain benchmark names: %s", s)
	}
}
