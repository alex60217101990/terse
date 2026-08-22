package cache_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/cache"
)

func TestCapturePath_IsUnderCaptureDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got := cache.CapturePath("abc123")
	if filepath.Dir(got) != cache.CaptureDir() {
		t.Errorf("CapturePath = %q, want a child of %q", got, cache.CaptureDir())
	}
	if !strings.HasSuffix(got, "abc123.out") {
		t.Errorf("CapturePath = %q, want the id plus .out", got)
	}
}

func TestCapturePath_RejectsTraversal(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	got := cache.CapturePath("../../etc/passwd")
	if filepath.Dir(got) != cache.CaptureDir() {
		t.Errorf("CapturePath escaped its directory: %q", got)
	}
}

func TestGCCaptures_RemovesOldFiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := os.MkdirAll(cache.CaptureDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	old := cache.CapturePath("old")
	if err := os.WriteFile(old, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(old, stale, stale); err != nil {
		t.Fatal(err)
	}
	fresh := cache.CapturePath("fresh")
	if err := os.WriteFile(fresh, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	cache.GCCaptures(24 * time.Hour)
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("a stale capture must be removed")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a fresh capture must survive: it is still recoverable")
	}
}
