package cache_test

import (
	"fmt"
	"os"
	"path/filepath"
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
	t.Setenv("HOME", t.TempDir())
	// The directory is created lazily on the first Save, not by StateDir().
	if err := cache.Save("s", cache.NewSessionState()); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(cache.StateDir()); err != nil {
		t.Errorf("StateDir not created after Save: %v", err)
	}
}

func TestLoadCorruptFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Write garbage to the state file location (dir is no longer auto-created
	// by StateDir(), so create it here).
	if err := os.MkdirAll(cache.StateDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cache.StateDir(), "corrupt-session.qdf")
	if err := os.WriteFile(path, []byte("not valid qdf data \x00\xff"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := cache.Load("corrupt-session")
	if err != nil {
		t.Fatalf("Load of corrupt file should return empty state, not error: %v", err)
	}
	if len(s.Files) != 0 {
		t.Errorf("expected empty SessionState from corrupt file, got %+v", s)
	}
}

func BenchmarkSaveLoad(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	s := cache.NewSessionState()
	for i := range 50 {
		s.Files[fmt.Sprintf("file%d.go", i)] = cache.FileEntry{
			Hash:    [32]byte{byte(i)},
			Turn:    i,
			Content: []byte("package main\nfunc foo() {}\n"),
		}
	}

	for b.Loop() {
		_ = cache.Save("bench-session", s)
		_, _ = cache.Load("bench-session")
	}
}
