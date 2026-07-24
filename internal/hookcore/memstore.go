package hookcore

import (
	"hash/fnv"
	"maps"
	"os"
	"strings"
	"sync"
	"time"

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
// daemon: session state lives in sharded/mutex-guarded maps instead of
// hitting disk on every hook call. Sessions are the only content held in RAM
// and flushed lazily; ref/last-output blob *content* is never held in RAM —
// refs/last are hash/key seen-sets only, and every RefPut/LastPut writes its
// blob straight to disk (via internal/cache) so a daemon restart or crash
// never loses a blob body. Usage (dedup-hit / access recency, for eviction
// scoring) is tracked in RAM and flushed lazily to the usage sidecars.
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

	// refs / last are seen-sets only (hash / key -> presence), never content.
	// Content lives on disk from the moment RefPut/LastPut is called.
	refs map[string]struct{}

	last map[string]struct{}

	// usageRefs / usageLast track dedup-hit / access usage in RAM, flushed
	// lazily to their sidecars by FlushDirty. Guarded by usageMu, independent
	// of refsMu/lastMu (which guard only the seen-sets).
	usageRefs cache.UsageIndex

	usageLast cache.UsageIndex

	dirtySessions map[string]struct{}

	refsMu sync.RWMutex

	lastMu sync.RWMutex

	usageMu sync.Mutex

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
		refs:          make(map[string]struct{}),
		last:          make(map[string]struct{}),
		usageRefs:     cache.LoadUsage(cache.UsageRefsPath()),
		usageLast:     cache.LoadUsage(cache.UsageLastPath()),
		dirtySessions: make(map[string]struct{}),
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

// loadFromDisk walks the three on-disk stores. Sessions are decoded fully
// into RAM (they're the one thing MemStore actually caches); refs/last are
// only scanned for their *names* — each on-disk blob's hash/key is added to
// the seen-set, never its content, so a daemon with a large on-disk cache
// does not balloon its RAM footprint at startup.
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
			m.refsMu.Lock()
			m.refs[hash] = struct{}{}
			m.refsMu.Unlock()
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
			m.lastMu.Lock()
			m.last[key] = struct{}{}
			m.lastMu.Unlock()
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

// RefSeen reports whether content addressed by hash was already stored (a
// seen-set lookup — never touches disk).
func (m *MemStore) RefSeen(hash string) bool {
	m.refsMu.RLock()
	_, ok := m.refs[hash]
	m.refsMu.RUnlock()
	return ok
}

// RefPut writes content to disk under hash (lazy qdf-encoded blob) and adds
// hash to the seen-set. Unlike sessions, ref blobs are never held dirty in
// RAM — they go straight to disk here, so a daemon crash between RefPut and
// the next FlushDirty never loses a blob body.
func (m *MemStore) RefPut(hash, content string) {
	cache.RefPut(hash, content)

	m.refsMu.Lock()
	m.refs[hash] = struct{}{}
	m.refsMu.Unlock()
}

// RefGet reads the content stored under hash from disk on demand. MemStore
// never holds ref content in RAM.
func (m *MemStore) RefGet(hash string) (string, bool) {
	return cache.RefGet(hash)
}

// RefHit records a dedup hit against hash: bumps its in-RAM usage stat
// (hits + last-access time), flushed lazily to the usage sidecar by
// FlushDirty. Guarded by usageMu, independent of refsMu (which guards only
// the seen-set).
func (m *MemStore) RefHit(hash string) {
	m.usageMu.Lock()
	m.usageRefs.Bump(hash, time.Now().Unix())
	m.usageMu.Unlock()
}

// LastGet reads the previous tool output stored under key from disk on
// demand. MemStore never holds last-output content in RAM.
func (m *MemStore) LastGet(key string) (string, bool) {
	return cache.LastOutputGet(key)
}

// LastPut writes the current tool output to disk under key and adds key to
// the seen-set. Like RefPut, this goes straight to disk — no dirty content
// held in RAM.
func (m *MemStore) LastPut(key, content string) {
	cache.LastOutputPut(key, content)

	m.lastMu.Lock()
	m.last[key] = struct{}{}
	m.lastMu.Unlock()
}

// FlushDirty persists every dirty session to disk via cache.Save (qdf
// OptSpeed), then flushes the in-RAM usage indices (ref/last dedup-hit and
// access stats) to their sidecars. Ref/last-output *blobs* are never dirty
// here — RefPut/LastPut already wrote them straight to disk — so there is
// nothing to flush for them beyond usage.
//
// The dirty session set is swapped out under dirtyMu and iterated afterwards
// without holding it, so writes that arrive concurrently with a flush land in
// a fresh dirty set rather than being lost or blocking on disk I/O. The usage
// maps are cloned under usageMu for the same reason: SaveUsage's JSON
// marshaling happens outside the lock so a concurrent RefHit/usage bump never
// blocks on disk I/O.
func (m *MemStore) FlushDirty() {
	m.dirtyMu.Lock()
	sessions := m.dirtySessions
	m.dirtySessions = make(map[string]struct{})
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

	m.usageMu.Lock()
	ur := maps.Clone(m.usageRefs)
	ul := maps.Clone(m.usageLast)
	m.usageMu.Unlock()
	_ = cache.SaveUsage(cache.UsageRefsPath(), ur)
	_ = cache.SaveUsage(cache.UsageLastPath(), ul)
}
