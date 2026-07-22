package cache_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/cache"
)

func TestLoadSaveRoundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	s := cache.NewSessionState()
	s.Turn = 7
	s.Files["encoder.go"] = cache.FileEntry{
		Hash:    [32]byte{1, 2, 3},
		Turn:    3,
		Content: []byte("package main\n"),
	}

	const sid = "test-session-42"
	if err := cache.Save(sid, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := cache.Load(sid)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Turn != 7 {
		t.Errorf("Turn = %d, want 7", got.Turn)
	}
	entry, ok := got.Files["encoder.go"]
	if !ok {
		t.Fatal("Files[encoder.go] missing after round-trip")
	}
	if entry.Hash != [32]byte{1, 2, 3} {
		t.Errorf("Hash mismatch")
	}
	if string(entry.Content) != "package main\n" {
		t.Errorf("Content = %q, want %q", entry.Content, "package main\n")
	}
}

func TestLoadMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s, err := cache.Load("no-such-session")
	if err != nil {
		t.Fatalf("Load of missing session should return empty state, got err: %v", err)
	}
	if s == nil || len(s.Files) != 0 {
		t.Errorf("expected empty SessionState, got %+v", s)
	}
}

func TestStateDirCreated(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	_ = cache.StateDir()
	if _, err := os.Stat(cache.StateDir()); err != nil {
		t.Errorf("StateDir not created: %v", err)
	}
}

func BenchmarkSaveLoad(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	s := cache.NewSessionState()
	for i := 0; i < 50; i++ {
		s.Files[fmt.Sprintf("file%d.go", i)] = cache.FileEntry{
			Hash:    [32]byte{byte(i)},
			Turn:    i,
			Content: []byte("package main\nfunc foo() {}\n"),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Save("bench-session", s)
		_, _ = cache.Load("bench-session")
	}
}
