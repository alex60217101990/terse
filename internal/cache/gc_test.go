package cache_test

import (
	"os"
	"testing"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/cache"
)

func TestRunGC_RemovesLowScore(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("QDF_DECAY_LAMBDA", "10") // aggressive decay

	// Write a session with no reads (score == 0, below threshold).
	s := cache.NewSessionState()
	_ = cache.Save("gc-test-stale", s)

	// Write a session with many reads, read just now (age ≈ 0, score ≈ 100).
	s2 := cache.NewSessionState()
	s2.Files["/a.go"] = cache.FileEntry{ReadCount: 100, LastReadAt: time.Now().Unix()}
	_ = cache.Save("gc-test-active", s2)

	result, err := cache.RunGC(false, 0.01)
	if err != nil {
		t.Fatalf("RunGC: %v", err)
	}
	if result.Removed == 0 {
		t.Error("expected at least one session removed")
	}
	// Active session should survive.
	if _, err := os.Stat(cache.StatePath("gc-test-active")); os.IsNotExist(err) {
		t.Error("active session should not be removed")
	}
}

func TestRunGC_DryRun_NoDelete(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	s := cache.NewSessionState()
	_ = cache.Save("gc-dry", s)

	_, err := cache.RunGC(true, 0.01) // dry-run: nothing should be deleted
	if err != nil {
		t.Fatalf("RunGC: %v", err)
	}
	// File must still exist after dry-run.
	if _, err := os.Stat(cache.StatePath("gc-dry")); os.IsNotExist(err) {
		t.Error("dry run should not delete files")
	}
}
