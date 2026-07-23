package hookcore_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
)

func TestMemStore_RoundTripAndFlush(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()
	st := cache.NewSessionState()
	st.Turn = 7
	s.SaveSession("a", st)
	s.RefPut("h1", "payload")
	m.FlushDirty()

	m2 := hookcore.NewMemStore().StateStore()
	if m2.LoadSession("a").Turn != 7 {
		t.Fatal("session not persisted")
	}
	if v, ok := m2.RefGet("h1"); !ok || v != "payload" {
		t.Fatalf("ref %q", v)
	}
}

func TestMemStore_LastOutputRoundTripAndFlush(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()
	s.LastPut("k1", "out")
	m.FlushDirty()

	m2 := hookcore.NewMemStore().StateStore()
	if v, ok := m2.LastGet("k1"); !ok || v != "out" {
		t.Fatalf("last %q %v", v, ok)
	}
}

func TestMemStore_LoadSessionMissingReturnsEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := hookcore.NewMemStore().StateStore()
	got := s.LoadSession("does-not-exist")
	if got == nil {
		t.Fatal("expected a non-nil empty session state")
	}
	if got.Turn != 0 || len(got.Files) != 0 {
		t.Fatalf("expected zero-value session, got %+v", got)
	}
}

// TestMemStore_ConcurrentSessionsAndRefs hammers two distinct sessions plus
// the shared ref/last-output maps from many goroutines under -race, to
// exercise the sharded session locks and the ref/last mutexes independently.
func TestMemStore_ConcurrentSessionsAndRefs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()

	const goroutines = 50
	const perGoroutine = 200
	sessions := []string{"session-a", "session-b"}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			id := sessions[g%len(sessions)]
			for i := 0; i < perGoroutine; i++ {
				st := cache.NewSessionState()
				st.Turn = i
				s.SaveSession(id, st)
				_ = s.LoadSession(id)

				hash := fmt.Sprintf("ref-%d-%d", g, i%7)
				s.RefPut(hash, "payload")
				_ = s.RefSeen(hash)
				_, _ = s.RefGet(hash)

				key := fmt.Sprintf("last-%d-%d", g, i%7)
				s.LastPut(key, "out")
				_, _ = s.LastGet(key)
			}
		}(g)
	}
	wg.Wait()

	m.FlushDirty()

	for _, id := range sessions {
		if got := s.LoadSession(id); got == nil {
			t.Fatalf("session %q missing after flush", id)
		}
	}
}
