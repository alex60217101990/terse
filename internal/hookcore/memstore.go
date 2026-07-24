package hookcore

import (
	"hash/fnv"
	"maps"
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
	sessions map[string]*cache.SessionState
	mu       sync.RWMutex
}

// MemStore is a fully in-RAM implementation of StateStore for the qdf-hookd
// daemon: session state, dedup refs, and last-tool-output blobs all live in
// sharded/mutex-guarded maps instead of hitting disk on every hook call.
// Writes only mark entries dirty; FlushDirty persists them via the same
// internal/cache writers the CLI's diskStore uses (qdf OptSpeed/OptBalanced,
// never encoding/json).
//
// Safe for concurrent use by multiple goroutines. LoadSession returns a copy
// of the stored session (made under the shard's lock, with its own Files
// map), never the live pointer, so callers are always free to mutate the
// returned state in place. SaveSession is the only way a stored session
// changes: it atomically swaps in the given pointer under the shard's write
// lock, so a stored session is never mutated after being stored — safe for
// FlushDirty to read/marshal it outside the lock. There is no enforced
// single-writer-per-session invariant: two concurrent LoadSession/SaveSession
// round trips for the *same* id race at the application level and the later
// SaveSession simply overwrites the earlier one (last-writer-wins); neither
// call panics or corrupts memory, but one caller's edits can be lost.
type MemStore struct {
	shards [shardCount]*sessionShard

	refs map[string]string

	last map[string]string

	dirtySessions map[string]struct{}
	dirtyRefs     map[string]struct{}
	dirtyLast     map[string]struct{}

	refsMu sync.RWMutex

	lastMu sync.RWMutex

	dirtyMu sync.Mutex
}

// NewMemStore builds a MemStore and loads any existing on-disk state
// (sessions under cache.StateDir, refs under cache.RefsDir, last-output blobs
// under cache.LastOutDir) into RAM. Loading is best-effort: missing
// directories fail open (nothing is loaded for that store), and construction
// never fails because of them. Session files are handled differently: a
// corrupt session file is not skipped — cache.Load returns err == nil for a
// decode failure and hands back a fresh, empty SessionState, so a corrupt
// on-disk session is loaded into RAM as that empty state (matching
// cache.Load's own "corrupt file -> fresh state" behavior) rather than being
// left out of the map entirely.
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

// LoadSession returns a copy of the session state for id, or a fresh empty
// state if none is in RAM yet. The copy has its own Files map, populated
// under the shard lock, so the caller can freely mutate the returned state
// (including its Files map) without racing a concurrent FlushDirty or another
// LoadSession/SaveSession round trip on the same id.
//
// This is last-writer-wins, not single-writer: if two goroutines both load
// the same session, mutate their copies, and save, the second SaveSession
// wins and the first goroutine's edits are silently lost. That is a callers'
// concern (real callers currently only drive one hook invocation per session
// id at a time); LoadSession/SaveSession themselves never panic or corrupt
// state no matter how many goroutines race on the same id.
func (m *MemStore) LoadSession(id string) *cache.SessionState {
	sh := m.shardFor(id)
	sh.mu.RLock()
	st, ok := sh.sessions[id]
	if !ok {
		sh.mu.RUnlock()
		return cache.NewSessionState()
	}
	cp := copySession(st)
	sh.mu.RUnlock()
	return cp
}

// copySession returns a copy of st with its own Files map. The stored session
// must never be handed out or written to a mutating consumer directly: both
// LoadSession's callers (which edit in place) and FlushDirty's cache.Save
// (which mutates Files via Evict) operate on a copy, so the map a concurrent
// LoadSession reads under sh.mu is never structurally modified without a lock.
// FileEntry values (including Content []byte) are shared by the copy, which is
// safe because nothing mutates a FileEntry in place — entries are only added,
// replaced, or deleted at the map level.
func copySession(st *cache.SessionState) *cache.SessionState {
	cp := &cache.SessionState{
		Turn:        st.Turn,
		CompactedAt: st.CompactedAt,
		Files:       make(map[string]cache.FileEntry, len(st.Files)),
	}
	maps.Copy(cp.Files, st.Files)
	return cp
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
		var cp *cache.SessionState
		if ok {
			// Copy under the lock: cache.Save calls Evict, which deletes from
			// Files. Mutating the stored map here (even though we only hold
			// RLock) would race a concurrent LoadSession's read of the same
			// map — a fatal "concurrent map read and map write" that recover
			// cannot catch. Save the copy instead.
			cp = copySession(st)
		}
		sh.mu.RUnlock()
		if ok {
			_ = cache.Save(id, cp)
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
