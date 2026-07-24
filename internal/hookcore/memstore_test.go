package hookcore_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
)

// TestMemStore_FlushEvictConcurrentLoad reproduces the daemon crash where
// FlushDirty handed the live stored *SessionState to cache.Save, whose Evict
// deletes from Files (>200-file sessions), racing a concurrent LoadSession's
// read of that same map — a fatal "concurrent map read and map write" that
// recover cannot catch. With the copy-before-Save fix this runs clean under
// -race. The session carries >200 files so Evict actually fires on flush.
func TestMemStore_FlushEvictConcurrentLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()

	newBig := func() *cache.SessionState {
		st := cache.NewSessionState()
		for i := range 260 { // > 200 so cache.Save's Evict deletes entries
			st.Files[fmt.Sprintf("/f/%d", i)] = cache.FileEntry{Turn: i}
		}
		return st
	}
	s.SaveSession("sess", newBig())

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				_ = m.LoadSession("sess") // reads Files under RLock
			}
		}()
	}
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				s.SaveSession("sess", newBig()) // re-mark dirty
				m.FlushDirty()                  // cache.Save -> Evict deletes
			}
		}()
	}
	wg.Wait()
}

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
	for g := range goroutines {
		go func(g int) {
			defer wg.Done()
			id := sessions[g%len(sessions)]
			for i := range perGoroutine {
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

// TestMemStore_ConcurrentLoadMutateSaveSameSession reproduces the real
// load->mutate-in-place->save pattern used by internal/hook/read.go and
// internal/hook/write.go: N goroutines all drive the SAME session id, each
// loading the session, mutating the returned state's Files map in place, and
// saving it back — while a concurrent goroutine repeatedly calls
// FlushDirty(), which reads/ranges that same Files map to marshal it.
//
// Before the fix, LoadSession returned the live pointer, so two goroutines
// racing on the same id shared one Files map: concurrent writes to it panic
// with "concurrent map writes", and a concurrent FlushDirty ranging the map
// while a handler writes it panics with "concurrent map iteration and map
// write". Under -race those show up as data races even when the panic
// doesn't fire on a given run. After the fix (LoadSession returns a copy),
// this must be race-free.
func TestMemStore_ConcurrentLoadMutateSaveSameSession(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()
	const sessionID = "shared-session"

	const goroutines = 50
	const perGoroutine = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(g int) {
			defer wg.Done()
			for i := range perGoroutine {
				state := s.LoadSession(sessionID)
				state.Turn++
				path := fmt.Sprintf("/file-%d-%d.go", g, i%5)
				state.Files[path] = cache.FileEntry{
					Content: []byte("content"),
					Turn:    state.Turn,
				}
				s.SaveSession(sessionID, state)
			}
		}(g)
	}

	// Concurrent flusher: FlushDirty ranges the stored session's Files map to
	// marshal it (via cache.Save -> qdf.Marshal), exercising the read side of
	// the same race.
	stop := make(chan struct{})
	var flushWg sync.WaitGroup
	flushWg.Go(func() {
		for {
			select {
			case <-stop:
				m.FlushDirty()
				return
			default:
				m.FlushDirty()
			}
		}
	})

	wg.Wait()
	close(stop)
	flushWg.Wait()

	if got := s.LoadSession(sessionID); got == nil {
		t.Fatal("session missing after concurrent load/mutate/save")
	}
}
