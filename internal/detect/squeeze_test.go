package detect_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/detect"
)

func TestSqueezeOutput_CollapsesRepeatedLines(t *testing.T) {
	in := strings.Repeat("downloading chunk...\n", 100) + "done\n"
	out := detect.SqueezeOutput(in)
	if !strings.Contains(out, "⨯100") {
		t.Errorf("expected run-length marker ⨯100, got:\n%s", out)
	}
	if !strings.Contains(out, "done") {
		t.Errorf("must keep the trailing distinct line, got:\n%s", out)
	}
	if len(out) >= len(in) {
		t.Errorf("must shrink: %d >= %d", len(out), len(in))
	}
}

func TestSqueezeOutput_StripsANSI(t *testing.T) {
	in := "\x1b[31mERROR\x1b[0m: something failed\n\x1b[32mok\x1b[0m\n"
	out := detect.SqueezeOutput(in)
	if strings.ContainsRune(out, 0x1b) {
		t.Errorf("ANSI escapes must be stripped, got: %q", out)
	}
	if !strings.Contains(out, "ERROR: something failed") {
		t.Errorf("text content must survive, got: %q", out)
	}
}

func TestSqueezeOutput_NoChange(t *testing.T) {
	in := "line one\nline two\nline three\n"
	if out := detect.SqueezeOutput(in); out != in {
		t.Errorf("distinct plain lines must be returned unchanged, got: %q", out)
	}
}

func BenchmarkSqueezeOutput(b *testing.B) {
	in := strings.Repeat("\x1b[2K\rprogress: step\n", 500) + "final line\n"
	b.ResetTimer()
	for b.Loop() {
		_ = detect.SqueezeOutput(in)
	}
}
