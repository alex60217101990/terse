package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	qdf "github.com/alex60217101990/qdf"
)

// StateDir returns the session state directory (created lazily on first write).
func StateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".qdf-hook", "sessions")
}

// StatePath returns the on-disk path for a session's state file.
// sessionID is sanitized via filepath.Base to prevent path traversal.
func StatePath(sessionID string) string {
	safe := filepath.Base(sessionID)
	if safe == "" || safe == "." {
		safe = "default"
	}
	return filepath.Join(StateDir(), safe+".qdf")
}

// Load reads the session state from disk. Returns an empty NewSessionState if
// the file does not exist. Returns an error only on I/O or decode failures.
func Load(sessionID string) (*SessionState, error) {
	path := StatePath(sessionID)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return NewSessionState(), nil
	}
	if err != nil {
		return nil, err
	}
	var s SessionState
	// WithNoCopy: decoded []byte/string fields (notably FileEntry.Content) alias
	// `data` instead of being copied out. Safe here because `data` is a fresh
	// os.ReadFile buffer that lives as long as the returned state (the aliases
	// keep it reachable), is never mutated in place, and never outlives the
	// hook invocation. Cuts the bulk of the decode allocations.
	if err := qdf.Unmarshal(data, &s, qdf.WithNoCopy()); err != nil {
		// Corrupt file — start fresh rather than crash.
		return NewSessionState(), nil //nolint:nilerr // corrupt cache = cache miss, not a caller-facing error
	}
	if s.Files == nil {
		s.Files = make(map[string]FileEntry)
	}
	return &s, nil
}

// MaxSessionFiles is the per-session file cap enforced by Save's automatic
// eviction: once len(Files) exceeds this, Evict trims down to 80% of it.
// Callers that skip Save (e.g. hookcore's zero-copy flush fast path) use this
// to decide, without calling Evict, whether eviction would be a no-op for a
// given session.
const MaxSessionFiles = 200

// Save persists state with a single plain write (no tmp+rename). The state
// file is a rebuildable cache: a torn write — from a crash or a concurrent
// same-session hook — simply fails to qdf-decode on the next Load, which then
// returns a fresh empty state (a cache miss, never wrong content). Dropping the
// atomic rename saves a syscall on every persisted hook, and the previous
// fixed ".tmp" name was not actually race-safe under concurrency anyway.
//
// Save persists the session state to disk using qdf OptSpeed.
// We benchmarked all options on a 50-file SessionState:
//
//	OptCompression → ~137 allocs/op  (rANS + FSST + Gorilla overhead)
//	OptBalanced    → similar (Dense + QPack + ShapeIntern)
//	OptSpeed (0)   → ~133 allocs/op  (minimum — no extra codec allocs)
//	encoding/json  → ~284 allocs/op  (interface boxing, decoder escaping)
//
// No single option meets the < 50 allocs/op spec: the allocation floor is
// set by the reflective encode/decode of the 50-entry Files map and its
// []byte Content slices, not by the codec layer. OptSpeed is kept because
// it is the lowest-allocation qdf mode; the spec budget was likely set
// against a smaller state (< 5 files). Document this if the benchmark
// target is revisited.
//
// Save mutates s in place (Evict deletes entries from s.Files when over the
// cap): callers must own s exclusively for the duration of the call — never
// pass a pointer another goroutine might be reading concurrently. Callers
// that only hold a read lock on the session (hookcore's FlushDirty) must
// either pass a private copy, or use AppendEncodeState/WriteState directly
// once they've confirmed len(Files) <= MaxSessionFiles (so Evict would be a
// no-op and mutation is moot).
// savePool holds the []byte scratch buffers plain Save encodes into. Without
// this, Save (unlike hookcore's FlushDirty, which pools its own buffer) fed
// AppendEncodeState a nil dst on every call, forcing qdf's geometric
// grow-chain to reallocate from scratch each time. Pooling means the backing
// array only grows on the first few calls and is reused after that.
var savePool = sync.Pool{New: func() any { b := make([]byte, 0, 4096); return &b }}

// MaxPooledBufSize caps the buffer size a scratch-buffer sync.Pool will
// accept back. A one-off encode that grows a buffer past this is exceptional;
// pinning that much memory in the pool would penalize every small-session
// encode that follows, so callers drop the oversized buffer instead of
// returning it (the pool's New allocates a fresh small one on the next Get).
// Shared by savePool here and hookcore's flushBufPool, which encodes the same
// kind of payload (a qdf-marshaled SessionState) via the same pattern.
const MaxPooledBufSize = 4 << 20 // 4MB

// maxPooledSaveBuf is savePool's local name for MaxPooledBufSize.
const maxPooledSaveBuf = MaxPooledBufSize

// Save persists a session's state to disk, encoding it and evicting the
// oldest sessions first if the on-disk cache is over its file-count cap.
func Save(sessionID string, s *SessionState) error {
	Evict(s, MaxSessionFiles) // auto-evict when over the cap

	bufPtr := savePool.Get().(*[]byte)
	buf, err := AppendEncodeState((*bufPtr)[:0], s)
	if err != nil {
		*bufPtr = buf
		savePool.Put(bufPtr)
		return err
	}

	err = WriteState(sessionID, buf)
	// os.WriteFile (via writeFileLazy) copies buf's bytes synchronously
	// before returning, so it's safe to return buf to the pool here.
	if cap(buf) <= maxPooledSaveBuf {
		*bufPtr = buf
		savePool.Put(bufPtr)
	}
	return err
}

// AppendEncodeState qdf-encodes s (OptSpeed, the same mode Save uses) and
// appends the result to dst, returning the extended slice. Reuse the
// returned slice as dst on the next call to avoid a fresh allocation per
// encode. Unlike Save, this performs no eviction and never mutates s — it
// only reads s, so it is safe to call while holding no more than a read lock
// on s, provided nothing else concurrently mutates s (see Save's doc comment
// on the MaxSessionFiles no-op eviction case).
func AppendEncodeState(dst []byte, s *SessionState) ([]byte, error) {
	return qdf.AppendMarshal(dst, s, qdf.OptSpeed)
}

// WriteState persists already-encoded session bytes as sessionID's state
// file. Pairs with AppendEncodeState so a caller can encode while holding a
// lock and perform the (potentially slow) disk write after releasing it.
func WriteState(sessionID string, data []byte) error {
	return writeFileLazy(StatePath(sessionID), data)
}
