package cache

// SessionState is the per-session persistent state of qdf-hook.
// Stored as a qdf-compressed file between hook invocations.
type SessionState struct {
	Turn        int                  `json:"turn"`
	CompactedAt int                  `json:"compacted_at"`
	Files       map[string]FileEntry `json:"files"`
}

// FileEntry records the last-seen state of a file Claude read.
type FileEntry struct {
	Hash    [32]byte `json:"hash"`
	Turn    int      `json:"turn"`
	Content []byte   `json:"content"`
}

// NewSessionState returns an empty, ready-to-use SessionState.
func NewSessionState() *SessionState {
	return &SessionState{
		Files: make(map[string]FileEntry),
	}
}

// SeenAfterCompact reports whether the file was read in the current
// (post-compaction) epoch. Files seen before the last compact need
// full content on re-read because the compaction erased the context.
func (s *SessionState) SeenAfterCompact(path string) bool {
	e, ok := s.Files[path]
	return ok && e.Turn > s.CompactedAt
}
