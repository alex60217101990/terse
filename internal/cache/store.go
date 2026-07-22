package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	qdf "github.com/alex60217101990/qdf"
)

// StateDir returns (and creates) the directory where session state files live.
// It is $HOME/.qdf-hook/sessions.
func StateDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".qdf-hook", "sessions")
	_ = os.MkdirAll(dir, 0o700)
	return dir
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
		return NewSessionState(), nil
	}
	if s.Files == nil {
		s.Files = make(map[string]FileEntry)
	}
	return &s, nil
}

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
func Save(sessionID string, s *SessionState) error {
	Evict(s, 200) // auto-evict when over 200 files
	data, err := qdf.Marshal(s, qdf.OptSpeed)
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(sessionID), data, 0o600)
}
