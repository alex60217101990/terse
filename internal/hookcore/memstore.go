package hookcore

import (
	"hash/fnv"
	"os"
	"strings"
	"sync"

	"github.com/alex60217101990/qdf-hook/internal/cache"
)

// shardCount is the number of independently-locked session shards. 16 keeps
// lock contention low for the handful of concurrent sessions a daemon
// realistically juggles, without the memory overhead of one mutex per id.
const shardCount = 16

// sessionShard is one shard of the session map, independently locked so
// operations on sessions in different shards never contend.
type sessionShard struct {
	mu       sync.RWMutex
	sessions map[string]*cache.SessionState
}

// MemStore is a fully in-RAM implementation of StateStore for the qdf-hookd
// daemon: session state, dedup refs, and last-tool-output blobs all live in
// sharded/mutex-guarded maps instead of hitting disk on every hook call.
// Writes only mark entries dirty; FlushDirty persists them via the same
// internal/cache writers the CLI's diskStore uses (qdf OptSpeed/OptBalanced,
// never encoding/json).
//
// Safe for concurrent use by multiple goroutines. LoadSession returns the
// live *cache.SessionState pointer under the shard's read lock (not a deep
// copy) — cheap, but callers must not mutate a session concurrently from two
// goroutines at once. In practice a given session id is only ever driven by
// one in-flight hook invocation at a time, so cross-goroutine mutation of the
// same session's Files map does not happen; concurrency across *different*
// session ids is what MemStore actually parallelizes, and that is race-free.
type MemStore struct {
	shards [shardCount]*sessionShard

	refsMu sync.RWMutex
	refs   map[string]string

	lastMu sync.RWMutex
	last   map[string]string

	dirtyMu       sync.Mutex
	dirtySessions map[string]struct{}
	dirtyRefs     map[string]struct{}
	dirtyLast     map[string]struct{}
}

// NewMemStore builds a MemStore and loads any existing on-disk state
// (sessions under cache.StateDir, refs under cache.RefsDir, last-output blobs
// under cache.LastOutDir) into RAM. Loading is best-effort: missing
// directories and individual decode failures are skipped rather than failing
// construction, matching cache.Load's own "corrupt file -> fresh state"
// behavior.
func NewMemStore() *MemStore {
	m := &MemStore{
		refs:          make(map[string]string),
		last:          make(map[string]string),
		dirtySessions: make(map[string]struct{}),
		dirtyRefs:     make(map[string]struct{}),
		dirtyLast:     make(map[string]struct{}),
	}
	for i := range m.shards {
		m.shards[i] = &sessionShard{sessions: make(map[string]*cache.SessionState)}
	}
	m.loadFromDisk()
	return m
}

// StateStore returns m as the StateStore interface. MemStore implements the
// interface directly, so this is an identity conversion (no adapter needed).
func (m *MemStore) StateStore() StateStore { return m }

// shardIndex picks a session's shard by FNV-1a hash of its id.
func shardIndex(id string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return int(h.Sum32() % shardCount)
}

func (m *MemStore) shardFor(id string) *sessionShard {
	return m.shards[shardIndex(id)]
}

// loadFromDisk walks the three on-disk stores and decodes each entry into the
// corresponding in-RAM map. Entries loaded here are, by definition, already
// in sync with disk, so none of them are marked dirty.
func (m *MemStore) loadFromDisk() {
	if entries, err := os.ReadDir(cache.StateDir()); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			id, ok := strings.CutSuffix(e.Name(), ".qdf")
			if !ok {
				continue
			}
			st, err := cache.Load(id)
			if err != nil {
				continue
			}
			sh := m.shardFor(id)
			sh.mu.Lock()
			sh.sessions[id] = st
			sh.mu.Unlock()
		}
	}

	if entries, err := os.ReadDir(cache.RefsDir()); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			hash, ok := strings.CutSuffix(e.Name(), ".blob")
			if !ok {
				continue
			}
			if content, ok := cache.RefGet(hash); ok {
				m.refsMu.Lock()
				m.refs[hash] = content
				m.refsMu.Unlock()
			}
		}
	}

	if entries, err := os.ReadDir(cache.LastOutDir()); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			key, ok := strings.CutSuffix(e.Name(), ".blob")
			if !ok {
				continue
			}
			if content, ok := cache.LastOutputGet(key); ok {
				m.lastMu.Lock()
				m.last[key] = content
				m.lastMu.Unlock()
			}
		}
	}
}

// LoadSession returns the session state for id, or a fresh empty state if
// none is in RAM yet. See the MemStore doc comment for the aliasing contract.
func (m *MemStore) LoadSession(id string) *cache.SessionState {
	sh := m.shardFor(id)
	sh.mu.RLock()
	st, ok := sh.sessions[id]
	sh.mu.RUnlock()
	if ok {
		return st
	}
	return cache.NewSessionState()
}

// SaveSession stores s as the current state for id and marks it dirty.
func (m *MemStore) SaveSession(id string, s *cache.SessionState) {
	sh := m.shardFor(id)
	sh.mu.Lock()
	sh.sessions[id] = s
	sh.mu.Unlock()

	m.dirtyMu.Lock()
	m.dirtySessions[id] = struct{}{}
	m.dirtyMu.Unlock()
}

// RefSeen reports whether content addressed by hash is already in RAM.
func (m *MemStore) RefSeen(hash string) bool {
	m.refsMu.RLock()
	_, ok := m.refs[hash]
	m.refsMu.RUnlock()
	return ok
}

// RefPut stores content under hash and marks it dirty.
func (m *MemStore) RefPut(hash, content string) {
	m.refsMu.Lock()
	m.refs[hash] = content
	m.refsMu.Unlock()

	m.dirtyMu.Lock()
	m.dirtyRefs[hash] = struct{}{}
	m.dirtyMu.Unlock()
}

// RefGet returns the content stored under hash.
func (m *MemStore) RefGet(hash string) (string, bool) {
	m.refsMu.RLock()
	v, ok := m.refs[hash]
	m.refsMu.RUnlock()
	return v, ok
}

// LastGet returns the previous tool output stored under key.
func (m *MemStore) LastGet(key string) (string, bool) {
	m.lastMu.RLock()
	v, ok := m.last[key]
	m.lastMu.RUnlock()
	return v, ok
}

// LastPut stores the current tool output under key and marks it dirty.
func (m *MemStore) LastPut(key, content string) {
	m.lastMu.Lock()
	m.last[key] = content
	m.lastMu.Unlock()

	m.dirtyMu.Lock()
	m.dirtyLast[key] = struct{}{}
	m.dirtyMu.Unlock()
}

// FlushDirty persists every dirty session/ref/last-output entry to disk via
// the internal/cache writers (cache.Save uses qdf OptSpeed; RefPut/
// LastOutputPut use qdf OptBalanced), then clears the dirty set.
//
// The dirty sets are swapped out under dirtyMu and iterated afterwards
// without holding it, so writes that arrive concurrently with a flush land in
// a fresh dirty set rather than being lost or blocking on disk I/O.
func (m *MemStore) FlushDirty() {
	m.dirtyMu.Lock()
	sessions := m.dirtySessions
	refs := m.dirtyRefs
	last := m.dirtyLast
	m.dirtySessions = make(map[string]struct{})
	m.dirtyRefs = make(map[string]struct{})
	m.dirtyLast = make(map[string]struct{})
	m.dirtyMu.Unlock()

	for id := range sessions {
		sh := m.shardFor(id)
		sh.mu.RLock()
		st, ok := sh.sessions[id]
		sh.mu.RUnlock()
		if ok {
			_ = cache.Save(id, st)
		}
	}

	for hash := range refs {
		m.refsMu.RLock()
		content, ok := m.refs[hash]
		m.refsMu.RUnlock()
		if ok {
			cache.RefPut(hash, content)
		}
	}

	for key := range last {
		m.lastMu.RLock()
		content, ok := m.last[key]
		m.lastMu.RUnlock()
		if ok {
			cache.LastOutputPut(key, content)
		}
	}
}
