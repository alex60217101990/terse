package hookcore

import (
	"hash/fnv"
	"maps"
	"os"
	"strings"
	"sync"
	"time"
	"weak"

	"github.com/alex60217101990/terse/internal/cache"
)

// shardCount is the number of independently-locked session shards. 16 keeps
// lock contention low for the handful of concurrent sessions a daemon
// realistically juggles, without the memory overhead of one mutex per id.
const shardCount = 16

// sessionShard is one shard of the session map, independently locked so
// operations on sessions in different shards never contend.
// The session map holds weak pointers, not strong ones: a cold session (one
// that no dirty strong ref and no in-flight caller keeps alive) becomes
// GC-reclaimable, so the daemon's RSS no longer grows without bound with the
// number of sessions it has ever seen. Disk is the source of truth; a
// weak-collected clean session is transparently reloaded from disk on the
// next LoadSession. Dead (collected) entries are pruned lazily on access and
// swept in FlushDirty so the map itself stays bounded.
type sessionShard struct {
	sessions map[string]weak.Pointer[cache.SessionState]
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
// lock, so a stored session is never mutated in place after being stored —
// safe for FlushDirty to read/marshal it (whether inside or outside the
// shard lock; see FlushDirty's doc comment for the fast/slow path split).
// There is no enforced
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

	// dirtySessions holds a STRONG reference to every session that has been
	// saved but not yet persisted by FlushDirty. This is load-bearing: the
	// shard map only holds weak pointers, so without this strong ref a dirty
	// session could be GC-collected before its (only) copy reaches disk —
	// silent data loss, since disk is the source of truth and would have
	// nothing to reload. FlushDirty drops the strong ref only after a
	// successful persist (an encode/write failure keeps it, to retry).
	dirtySessions map[string]*cache.SessionState

	refsMu sync.RWMutex

	lastMu sync.RWMutex

	usageMu sync.Mutex

	dirtyMu sync.Mutex
}

// NewMemStore builds a MemStore and scans existing on-disk state (refs under
// cache.RefsDir, last-output blobs under cache.LastOutDir) into the in-RAM
// seen-sets. Loading is best-effort: missing directories fail open (nothing
// is loaded for that store), and construction never fails because of them.
//
// Sessions are deliberately NOT preloaded: the whole point of the weak
// session cache is that cold sessions do not occupy RAM, so eagerly decoding
// every session file at startup would defeat it (and spike RSS proportional
// to the number of sessions ever seen). Instead sessions are loaded lazily
// from disk on the first LoadSession that needs one, and cached weakly.
func NewMemStore() *MemStore {
	m := &MemStore{
		refs:          make(map[string]struct{}),
		last:          make(map[string]struct{}),
		usageRefs:     cache.LoadUsage(cache.UsageRefsPath()),
		usageLast:     cache.LoadUsage(cache.UsageLastPath()),
		dirtySessions: make(map[string]*cache.SessionState),
	}
	for i := range m.shards {
		m.shards[i] = &sessionShard{sessions: make(map[string]weak.Pointer[cache.SessionState])}
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

// loadFromDisk scans the ref and last-output stores for their *names* — each
// on-disk blob's hash/key is added to the seen-set, never its content, so a
// daemon with a large on-disk cache does not balloon its RAM footprint at
// startup. Sessions are not scanned here at all: they are loaded lazily and
// cached weakly by LoadSession (see NewMemStore's doc comment).
func (m *MemStore) loadFromDisk() {
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

// LoadSession returns a copy of the session state for id. The copy has its
// own Files map, populated under the shard lock, so the caller can freely
// mutate the returned state (including its Files map) without racing a
// concurrent FlushDirty or another LoadSession/SaveSession round trip on the
// same id.
//
// The shard map holds only weak pointers. There are three outcomes:
//
//   - Live hit: the weak pointer still resolves (the session is dirty, or
//     another caller holds it) — copy and return it.
//   - Collected / absent: the session was never cached, or it was clean and
//     the GC reclaimed its weak-only reference. Disk is the source of truth,
//     so reload it with cache.Load (which itself returns a fresh empty state
//     when no file exists), re-cache it weakly, and return a copy. This is
//     transparent to callers — a reclaimed clean session is indistinguishable
//     from a resident one except for the disk read.
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
	if wp, ok := sh.sessions[id]; ok {
		if st := wp.Value(); st != nil {
			cp := copySession(st)
			sh.mu.RUnlock()
			return cp
		}
	}
	sh.mu.RUnlock()

	// Miss: never cached, or the weak pointer was collected. Reload from disk
	// (the source of truth) without holding the shard lock — the read can be
	// slow and must not block concurrent shards' work.
	st, err := cache.Load(id)
	if err != nil || st == nil {
		// I/O error reading an existing file: preserve the non-nil contract
		// and don't cache a guess. cache.Load already maps "no file" and
		// "corrupt file" to a fresh empty state (err == nil), so reaching
		// here means a real read failure.
		return cache.NewSessionState()
	}

	sh.mu.Lock()
	// Re-check under the write lock: a concurrent LoadSession/SaveSession may
	// have installed a live pointer while we were reading disk. Prefer it so
	// concurrent callers converge on one cached instance and a just-saved
	// (possibly dirty) state is never clobbered by our disk snapshot.
	if wp, ok := sh.sessions[id]; ok {
		if cur := wp.Value(); cur != nil {
			cp := copySession(cur)
			sh.mu.Unlock()
			return cp
		}
	}
	sh.sessions[id] = weak.Make(st)
	cp := copySession(st)
	sh.mu.Unlock()
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

// SaveSession stores s as the current state for id and marks it dirty. The
// shard map gets a weak pointer (so the session becomes GC-reclaimable once
// clean and otherwise unreferenced), while dirtySessions keeps a strong
// pointer to the same s until FlushDirty persists it — the strong ref is what
// prevents a not-yet-flushed session from being collected out from under the
// weak pointer (see the dirtySessions field comment).
func (m *MemStore) SaveSession(id string, s *cache.SessionState) {
	sh := m.shardFor(id)
	sh.mu.Lock()
	sh.sessions[id] = weak.Make(s)
	sh.mu.Unlock()

	m.dirtyMu.Lock()
	m.dirtySessions[id] = s
	m.dirtyMu.Unlock()
}

// RefSeen reports whether content addressed by hash was already stored (a
// seen-set lookup — never touches disk).
func (m *MemStore) RefSeen(hash string) bool {
	m.refsMu.RLock()
	_, ok := m.refs[hash]
	m.refsMu.RUnlock()
	if !ok {
		return false // a miss is authoritative: the set is a superset of disk
	}
	// A set hit can lag disk: the periodic blob sweep (cache.SweepBlobs) deletes
	// evicted blobs but does not touch this in-RAM set, so a hash can remain
	// here after its blob is gone. Confirm the blob still exists — otherwise we
	// would emit a §ref token for pruned content and a later `expand` would
	// fail. If the blob is gone, forget the hash so dedupWithStore re-caches it
	// (restoring the design's "an evicted entry only costs a re-cache" invariant,
	// exactly as the CLI's stat-based RefSeen already behaves). The stat is paid
	// only on a set hit — the common miss path stays a pure RAM lookup.
	if cache.RefSeen(hash) {
		return true
	}
	m.refsMu.Lock()
	delete(m.refs, hash)
	m.refsMu.Unlock()
	return false
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

// PruneUsage drops entries from the in-RAM usageRefs/usageLast indices whose
// LastUsed is older than maxAge — the same TTL floor cache.PruneDir applies
// to the on-disk usage sidecars (nowSec-used > ttl ⇒ drop; see evict.go).
// Without this, Bump only ever grows usageRefs/usageLast (nothing in RefHit
// or LastPut ever deletes from them), so FlushDirty's periodic SaveUsage
// would keep rewriting every entry back into the sidecar — silently
// resurrecting whatever PruneDir had just pruned on disk, and growing the
// RAM maps without bound over the life of a long-running daemon.
//
// Call this from the same tick that invokes cache.SweepBlobs (the daemon's
// sweep ticker) so the RAM map's retention exactly tracks the sidecar's TTL
// policy: after a sweep+flush cycle, both sidecar and RAM agree on what
// survived. Safe to call with an empty or nil-backed index.
func (m *MemStore) PruneUsage(maxAge time.Duration) {
	cutoff := time.Now().Unix() - int64(maxAge/time.Second)
	m.usageMu.Lock()
	for k, v := range m.usageRefs {
		if v.LastUsed < cutoff {
			delete(m.usageRefs, k)
		}
	}
	for k, v := range m.usageLast {
		if v.LastUsed < cutoff {
			delete(m.usageLast, k)
		}
	}
	m.usageMu.Unlock()
}

// flushBufPool holds the []byte scratch buffers FlushDirty encodes sessions
// into. Pooling (rather than qdf.Marshal's per-call allocation) means the
// buffer's backing array only grows on the first few flushes and is reused
// for every session/flush after that, instead of being allocated fresh per
// session per flush.
var flushBufPool = sync.Pool{New: func() any { b := make([]byte, 0, 4096); return &b }}

// FlushDirty persists every dirty session to disk, then flushes the in-RAM
// usage indices (ref/last dedup-hit and access stats) to their sidecars.
// Ref/last-output *blobs* are never dirty here — RefPut/LastPut already wrote
// them straight to disk — so there is nothing to flush for them beyond usage.
//
// Each dirty session is flushed from the STRONG pointer held in
// dirtySessions, not by looking it back up in the (weak) shard map. That
// strong ref both guarantees the session is still alive to flush and pins the
// exact bytes that were saved. No shard lock is needed to read it: a saved
// *SessionState is never mutated in place (LoadSession hands callers a copy,
// so every mutation lands on a fresh pointer that a later SaveSession swaps
// in), so the only concurrent access to this pointer's Files map is other
// read-only reads — a concurrent LoadSession's copySession — and concurrent
// map reads are safe. Persistence has two paths:
//
//   - Fast path (the common case): the session has at most
//     cache.MaxSessionFiles entries, so cache.Save's automatic eviction would
//     be a no-op. qdf-encode the strong-ref *SessionState directly (read-only)
//     into a pooled buffer, then write to disk. This skips the copy entirely.
//   - Slow path (rare: a session actually over the cap): cache.Save's Evict
//     step deletes entries from the passed state's Files map. Mutating the
//     shared pointer's map would race a concurrent LoadSession's read of it —
//     a fatal "concurrent map read and map write" (this crash is what
//     copySession was introduced to fix; see TestMemStore_FlushEvictConcurrentLoad).
//     So this path copies the session before handing it to cache.Save.
//
// A session whose encode OR disk write fails keeps its strong ref (it is
// re-added to dirtySessions unless a newer save already superseded it), so a
// transient failure never drops the only copy of unpersisted state to the GC.
// After a successful persist the strong ref is dropped, which is exactly what
// makes the now-clean session GC-reclaimable via its remaining weak pointer.
// Finally the shards are swept of dead (collected) weak pointers so the maps
// stay bounded regardless of how many sessions have come and gone.
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
	m.dirtySessions = make(map[string]*cache.SessionState)
	m.dirtyMu.Unlock()

	bufPtr := flushBufPool.Get().(*[]byte)
	buf := *bufPtr

	var failed map[string]*cache.SessionState
	for id, st := range sessions {
		var err error
		if len(st.Files) > cache.MaxSessionFiles {
			// Slow path: eviction will actually delete entries, so it must
			// not run against the shared pointer's map (a concurrent
			// LoadSession may be reading it). Copy first.
			cp := copySession(st)
			err = cache.Save(id, cp)
		} else {
			// Fast path: encode the strong-ref pointer directly (read-only,
			// no copy) into the pooled buffer, then write to disk.
			buf, err = cache.AppendEncodeState(buf[:0], st)
			if err == nil {
				err = cache.WriteState(id, buf)
			}
		}
		if err != nil {
			// Persist failed: keep the strong ref so this unpersisted state
			// is neither lost nor weak-collected before a later retry.
			if failed == nil {
				failed = make(map[string]*cache.SessionState)
			}
			failed[id] = st
		}
	}

	// Cap what goes back to the pool: an outsized encode (a one-off session
	// with unusually large file content) would otherwise pin that much memory
	// in flushBufPool for every later small-session flush. Mirrors
	// cache.Save's savePool/cache.MaxPooledBufSize pattern exactly, since both
	// pools encode the same kind of payload (a qdf-marshaled SessionState).
	if cap(buf) <= cache.MaxPooledBufSize {
		*bufPtr = buf
		flushBufPool.Put(bufPtr)
	}

	// Re-arm strong refs for sessions that failed to persist, unless a newer
	// SaveSession already re-marked the id dirty (its pointer supersedes ours).
	if failed != nil {
		m.dirtyMu.Lock()
		for id, st := range failed {
			if _, superseded := m.dirtySessions[id]; !superseded {
				m.dirtySessions[id] = st
			}
		}
		m.dirtyMu.Unlock()
	}

	// Sweep collected weak pointers so the shard maps do not accumulate dead
	// entries for every session that has ever been seen and since reclaimed.
	for _, sh := range m.shards {
		sh.mu.Lock()
		for id, wp := range sh.sessions {
			if wp.Value() == nil {
				delete(sh.sessions, id)
			}
		}
		sh.mu.Unlock()
	}

	m.usageMu.Lock()
	ur := maps.Clone(m.usageRefs)
	ul := maps.Clone(m.usageLast)
	m.usageMu.Unlock()
	_ = cache.SaveUsage(cache.UsageRefsPath(), ur)
	_ = cache.SaveUsage(cache.UsageLastPath(), ul)
}
