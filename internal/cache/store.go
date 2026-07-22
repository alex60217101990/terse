package cache

import (
	"os"
	"path/filepath"

	qdf "github.com/alex60217101990/qdf"
)

// StateDir returns (and creates) the directory where session state files live.
// It is $HOME/.config/qdf-hook/state.
func StateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, ".config", "qdf-hook", "state")
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

// statePath returns the on-disk path for a session's state file.
func statePath(sessionID string) string {
	return filepath.Join(StateDir(), sessionID+".qdf")
}

// Load reads the session state from disk. Returns an empty NewSessionState if
// the file does not exist. Returns an error only on I/O or decode failures.
func Load(sessionID string) (*SessionState, error) {
	path := statePath(sessionID)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewSessionState(), nil
	}
	if err != nil {
		return nil, err
	}
	var s SessionState
	if err := qdf.Unmarshal(data, &s); err != nil {
		// Corrupt file — start fresh rather than crash.
		return NewSessionState(), nil
	}
	if s.Files == nil {
		s.Files = make(map[string]FileEntry)
	}
	return &s, nil
}

// Save persists the session state to disk using qdf OptCompression.
// It writes to a temp file and renames atomically to avoid partial writes.
func Save(sessionID string, s *SessionState) error {
	data, err := qdf.Marshal(s, qdf.OptCompression)
	if err != nil {
		return err
	}
	path := statePath(sessionID)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
