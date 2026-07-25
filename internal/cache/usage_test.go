package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
)

func TestUsageIndex_BumpSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage-refs.qdf")

	idx := cache.LoadUsage(path) // missing file -> empty, not nil
	if idx == nil {
		t.Fatal("LoadUsage of a missing file must return an empty index, not nil")
	}
	idx.Bump("abc", 1000)
	idx.Bump("abc", 2000)
	idx.Bump("def", 1500)

	if err := cache.SaveUsage(path, idx); err != nil {
		t.Fatalf("SaveUsage: %v", err)
	}

	got := cache.LoadUsage(path)
	if got["abc"].Hits != 2 || got["abc"].LastUsed != 2000 {
		t.Errorf("abc = %+v, want {Hits:2 LastUsed:2000}", got["abc"])
	}
	if got["def"].Hits != 1 || got["def"].LastUsed != 1500 {
		t.Errorf("def = %+v, want {Hits:1 LastUsed:1500}", got["def"])
	}
}

func TestLoadUsage_CorruptReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "u.qdf")
	if err := writeFileForTest(path, []byte("not qdf")); err != nil {
		t.Fatal(err)
	}
	if idx := cache.LoadUsage(path); len(idx) != 0 {
		t.Errorf("corrupt file must load as empty, got %d entries", len(idx))
	}
}

func writeFileForTest(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}
