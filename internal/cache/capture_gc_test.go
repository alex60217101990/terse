package cache_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/cache"
)

// writeCapture drops a capture of n bytes with the given age.
func writeCapture(t *testing.T, name string, n int, age time.Duration) string {
	t.Helper()
	path := filepath.Join(cache.CaptureDir(), name+".out")
	if err := os.WriteFile(path, make([]byte, n), 0o600); err != nil {
		t.Fatal(err)
	}
	when := time.Now().Add(-age)
	if err := os.Chtimes(path, when, when); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGCCaptures_DropsOldKeepsFresh(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := os.MkdirAll(cache.CaptureDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	old := writeCapture(t, "old", 100, 48*time.Hour)
	fresh := writeCapture(t, "fresh", 100, time.Minute)

	removed, freed := cache.GCCaptures(24 * time.Hour)
	if removed != 1 || freed != 100 {
		t.Errorf("removed %d captures freeing %d bytes, want 1 and 100", removed, freed)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("the aged-out capture is still there")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Errorf("the fresh capture was removed: %v", err)
	}
}

// TestGCCaptures_EnforcesSizeCap covers what age alone cannot: a heavy session
// writing thousands of captures inside one TTL window.
func TestGCCaptures_EnforcesSizeCap(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := os.MkdirAll(cache.CaptureDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	// Twice the cap, all fresh, each one distinctly older than the next.
	const each = 1 << 20
	count := 2 * cache.CaptureMaxSize / each // int, both operands are untyped constants
	for i := range count {
		writeCapture(t, fmt.Sprintf("c%03d", i), each, time.Duration(count-i)*time.Minute)
	}

	removed, freed := cache.GCCaptures(24 * time.Hour)
	if removed == 0 {
		t.Fatal("nothing was removed although the directory is twice the cap")
	}

	entries, err := os.ReadDir(cache.CaptureDir())
	if err != nil {
		t.Fatal(err)
	}
	var total int64
	for _, e := range entries {
		info, ierr := e.Info()
		if ierr != nil {
			continue
		}
		total += info.Size()
	}
	if total > cache.CaptureMaxSize {
		t.Errorf("captures still hold %d bytes, over the %d cap", total, cache.CaptureMaxSize)
	}
	if freed < int64(each) {
		t.Errorf("freed only %d bytes", freed)
	}
	// The newest capture is the one a running turn might still ask for.
	newest := filepath.Join(cache.CaptureDir(), fmt.Sprintf("c%03d.out", count-1))
	if _, err := os.Stat(newest); err != nil {
		t.Errorf("the newest capture was evicted first: %v", err)
	}
}
