package cache_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/cache"
)

func TestBashLast_RoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if _, ok := cache.BashLastGet("go test ./...", "/proj"); ok {
		t.Fatal("no prior entry expected")
	}
	out := strings.Repeat("ok pkg 0.1s\n", 50)
	cache.BashLastPut("go test ./...", "/proj", out)

	got, ok := cache.BashLastGet("go test ./...", "/proj")
	if !ok || got != out {
		t.Fatalf("round-trip mismatch: ok=%v len(got)=%d want=%d", ok, len(got), len(out))
	}
	// Different command+cwd must not collide.
	if _, ok := cache.BashLastGet("go test ./...", "/other"); ok {
		t.Error("cwd must be part of the key")
	}
}

func BenchmarkBashLastPut(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	out := strings.Repeat("some command output line\n", 100)
	b.ResetTimer()
	for b.Loop() {
		cache.BashLastPut("cmd", "/cwd", out)
	}
}

func BenchmarkBashLastGet(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	out := strings.Repeat("some command output line\n", 100)
	cache.BashLastPut("cmd", "/cwd", out)
	b.ResetTimer()
	for b.Loop() {
		_, _ = cache.BashLastGet("cmd", "/cwd")
	}
}
