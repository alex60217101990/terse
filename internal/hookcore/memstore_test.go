package hookcore_test

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/hookcore"
)

// TestMemStore_RefSeenWithoutHoldingContent verifies that a dedup hit reads
// its content from disk (not from RAM) and that RefHit's usage bump reaches
// the on-disk sidecar via FlushDirty.
func TestMemStore_RefSeenWithoutHoldingContent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()

	big := strings.Repeat("x", 300_000) // 300 KB
	hash := "deadbeefdeadbeefdeadbeefdeadbeef"
	s.RefPut(hash, big)

	if !s.RefSeen(hash) {
		t.Fatal("RefSeen must be true after RefPut")
	}
	if got, ok := s.RefGet(hash); !ok || got != big {
		t.Fatalf("RefGet must return the stored content from disk (ok=%v len=%d)", ok, len(got))
	}
	// A dedup hit records usage; after flush the sidecar has it.
	s.RefHit(hash)
	m.FlushDirty()
	idx := cache.LoadUsage(cache.UsageRefsPath())
	if idx[hash].Hits < 1 {
		t.Errorf("RefHit should have bumped usage, got %+v", idx[hash])
	}
}

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
		wg.Go(func() {
			for range 50 {
				_ = m.LoadSession("sess") // reads Files under RLock
			}
		})
	}
	for range 4 {
		wg.Go(func() {
			for range 50 {
				s.SaveSession("sess", newBig()) // re-mark dirty
				m.FlushDirty()                  // cache.Save -> Evict deletes
			}
		})
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

// TestMemStore_RefSeenSelfHealsAfterBlobEvicted is the final-review regression:
// the periodic blob sweep deletes a blob but not the in-RAM seen-set, so
// RefSeen must confirm against disk on a set hit — otherwise it would report a
// pruned hash as seen, yielding a §ref token whose blob is gone. After the
// blob is removed, RefSeen must return false and forget the hash so the caller
// re-caches.
func TestMemStore_RefSeenSelfHealsAfterBlobEvicted(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()

	hash := "cafebabecafebabecafebabecafebabe"
	s.RefPut(hash, "some cached content that is long enough")
	if !s.RefSeen(hash) {
		t.Fatal("RefSeen should be true right after RefPut")
	}

	// Simulate the sweep deleting the blob from disk (seen-set untouched).
	if err := os.Remove(cache.RefPath(hash)); err != nil {
		t.Fatalf("remove blob: %v", err)
	}

	if s.RefSeen(hash) {
		t.Fatal("RefSeen must return false once the blob is gone (dangling-ref guard)")
	}
	// Re-caching works and is seen again.
	s.RefPut(hash, "some cached content that is long enough")
	if !s.RefSeen(hash) {
		t.Fatal("RefSeen should be true again after re-caching")
	}
}

// TestMemStore_FlushDirty_NoDeadlockUnderLoadSaveStorm exercises the safety
// case FlushDirty's zero-copy fast path depends on: encoding the LIVE stored
// *SessionState directly under the shard's *read* lock (see FlushDirty's doc
// comment) is only safe because nothing mutates that pointer's Files map
// in place while the encode runs. 16 goroutines hammer LoadSession/SaveSession
// against one session id (well under cache.MaxSessionFiles, so every flush
// takes the fast path) concurrently with a tight FlushDirty loop. This must
// neither deadlock nor pathologically serialize, and every on-disk snapshot
// FlushDirty produces along the way must qdf-decode cleanly — a torn Files
// map (a partial/interleaved encode) would show up as a decode error or a
// corrupt entry.
func TestMemStore_FlushDirty_NoDeadlockUnderLoadSaveStorm(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()
	const sessionID = "storm-session"

	const goroutines = 16
	const filesPerSession = 40 // well under cache.MaxSessionFiles: fast path

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(g int) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				st := s.LoadSession(sessionID)
				st.Turn++
				path := fmt.Sprintf("/storm/%d/%d.go", g, st.Turn%filesPerSession)
				st.Files[path] = cache.FileEntry{
					Content: []byte("payload"),
					Turn:    st.Turn,
				}
				s.SaveSession(sessionID, st)
			}
		}(g)
	}

	flushDone := make(chan struct{})
	go func() {
		defer close(flushDone)
		for range 2000 {
			m.FlushDirty()
			// Every snapshot FlushDirty just wrote must decode cleanly: a
			// torn/interleaved encode of the live map would surface here as
			// a decode error or a corrupt entry (negative Turn can only come
			// from decoding garbage bytes as a length-prefixed field).
			st, err := cache.Load(sessionID)
			if err != nil {
				panic(fmt.Sprintf("flushed snapshot failed to decode: %v", err))
			}
			for path, fe := range st.Files {
				if fe.Turn < 0 {
					panic(fmt.Sprintf("corrupt entry %q after flush: %+v", path, fe))
				}
			}
		}
	}()

	select {
	case <-flushDone:
	case <-time.After(20 * time.Second):
		t.Fatal("FlushDirty storm deadlocked or stalled")
	}
	close(done)
	wg.Wait()

	// Final flush + decode after the storm settles.
	m.FlushDirty()
	st, err := cache.Load(sessionID)
	if err != nil {
		t.Fatalf("decode final flushed snapshot: %v", err)
	}
	if len(st.Files) == 0 {
		t.Fatal("expected a non-empty session after the storm")
	}
}

// newFlushBenchState builds a SessionState with n FileEntries of ~2KB
// Content each, matching the daemon's alloc-profile scenario (many
// moderately-sized files read during a long session).
func newFlushBenchState(base, n int) *cache.SessionState {
	content := []byte(strings.Repeat("x", 2048))
	st := cache.NewSessionState()
	for i := range n {
		st.Files[fmt.Sprintf("/bench/%d/file-%d.go", base, i)] = cache.FileEntry{
			Content: content,
			Turn:    i,
			Hash:    [32]byte{byte(i)},
		}
	}
	return st
}

// BenchmarkFlushDirty is Task 2's baseline/gate benchmark: 50 sessions x 40
// FileEntries (~2KB Content each), half re-marked dirty every iteration so
// each FlushDirty call has real work to do. Compare B/op before/after the
// zero-copy fast path with benchstat; target is a 50%+ reduction.
func BenchmarkFlushDirty(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()

	const numSessions = 50
	const filesPerSession = 40
	ids := make([]string, numSessions)
	states := make([]*cache.SessionState, numSessions)
	for i := range numSessions {
		ids[i] = fmt.Sprintf("bench-session-%d", i)
		states[i] = newFlushBenchState(i, filesPerSession)
		s.SaveSession(ids[i], states[i])
	}

	for b.Loop() {
		// Re-mark half the sessions dirty each iteration so FlushDirty has a
		// realistic mixed dirty/clean set to work through. SaveSession's own
		// cost (a pointer swap under the shard lock) is unchanged by this
		// task and identical across before/after runs, so it doesn't affect
		// the relative B/op comparison FlushDirty is gated on.
		for i := range numSessions / 2 {
			s.SaveSession(ids[i], states[i])
		}
		m.FlushDirty()
	}
}

// BenchmarkLoadSession is the no-regression gate for LoadSession: copySession
// is untouched by this task (still used on every LoadSession call), so this
// should be flat before/after.
func BenchmarkLoadSession(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()
	s.SaveSession("bench-load", newFlushBenchState(0, 40))

	for b.Loop() {
		_ = s.LoadSession("bench-load")
	}
}

// BenchmarkSaveSession is the no-regression gate for SaveSession: it is
// untouched by this task (still a pointer swap + dirty-set insert under
// lock), so this should be flat before/after.
func BenchmarkSaveSession(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	m := hookcore.NewMemStore()
	s := m.StateStore()
	st := newFlushBenchState(0, 40)

	for b.Loop() {
		s.SaveSession("bench-save", st)
	}
}
