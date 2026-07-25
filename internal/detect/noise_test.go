package detect_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/detect"
)

func TestStripNoise_DropsJunkKeepsSignal(t *testing.T) {
	in := "zsh: command not found: _encode foo\n" +
		"go: downloading github.com/x/y v1.2.3\n" +
		"real output line 1\n" +
		"npm notice New version available\n" +
		"real output line 2\n"
	out := detect.StripNoise(in)
	for _, junk := range []string{"command not found: _", "go: downloading", "npm notice"} {
		if strings.Contains(out, junk) {
			t.Errorf("junk %q survived:\n%s", junk, out)
		}
	}
	for _, sig := range []string{"real output line 1", "real output line 2"} {
		if !strings.Contains(out, sig) {
			t.Errorf("signal %q dropped:\n%s", sig, out)
		}
	}
}

func TestStripNoise_CleanUnchanged(t *testing.T) {
	in := "func main() {}\nfmt.Println(\"hi\")\nreturn nil\n"
	if out := detect.StripNoise(in); out != in {
		t.Errorf("clean input must be returned unchanged, got:\n%s", out)
	}
}

func BenchmarkStripNoise_Clean(b *testing.B) {
	in := strings.Repeat("some ordinary output line with content\n", 100)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = detect.StripNoise(in)
	}
}

func BenchmarkStripNoise_Dirty(b *testing.B) {
	in := strings.Repeat("go: downloading x\nreal line\nnpm notice y\nreal line 2\n", 50)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = detect.StripNoise(in)
	}
}

// TestStripNoise_KeepsWaitingFor: the over-broad "Waiting" prefix used to drop
// legitimate application lines like "Waiting for database connection". Once
// StripNoise runs (a discriminator is present), such a line must survive.
func TestStripNoise_KeepsWaitingFor(t *testing.T) {
	in := "go: downloading example.com/x v1.0.0\nWaiting for database connection...\ndone\n"
	out := detect.StripNoise(in)
	if !strings.Contains(out, "Waiting for database connection") {
		t.Fatalf("legit 'Waiting for...' line was dropped:\n%s", out)
	}
	if strings.Contains(out, "go: downloading") {
		t.Fatalf("noise line 'go: downloading' not stripped:\n%s", out)
	}
}
