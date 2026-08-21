package cache_test

import (
	"path/filepath"
	"strings"
	"testing"

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
