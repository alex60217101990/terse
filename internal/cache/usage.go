package cache

import (
	"os"
	"path/filepath"

	qdf "github.com/alex60217101990/qdf"
)

// UsageStat is the advisory usage record for one cache entry (a ref hash or a
// last-output key). LastUsed is unix seconds. Counters are advisory: they only
// order eviction, never affect correctness, so cross-process last-writer-wins
// is acceptable.
type UsageStat struct {
	Hits     uint32
	LastUsed int64
}

// UsageIndex maps a ref hash / last key to its usage.
type UsageIndex map[string]UsageStat

// usageFile is the on-disk shape (qdf can't marshal a bare map to a stable
// shape as cleanly as a struct with a named field).
type usageFile struct {
	Entries map[string]UsageStat
}

// UsageRefsPath / UsageLastPath are the two sidecar files (separate so a ref
// hash and a last key that happen to share hex can't collide).
func UsageRefsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".qdf-hook", "usage-refs.qdf")
}

// UsageLastPath is the usage index's last-key sidecar file (see UsageRefsPath).
func UsageLastPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".qdf-hook", "usage-last.qdf")
}

// LoadUsage reads a usage index. A missing or corrupt file yields an empty
// (non-nil) index — the entries are advisory, so failing open is safe.
func LoadUsage(path string) UsageIndex {
	data, err := os.ReadFile(path)
	if err != nil {
		return UsageIndex{}
	}
	var uf usageFile
	if err := qdf.Unmarshal(data, &uf); err != nil || uf.Entries == nil {
		return UsageIndex{}
	}
	return uf.Entries
}

// SaveUsage writes the index (plain lazy write, rebuildable state).
func SaveUsage(path string, idx UsageIndex) error {
	data, err := qdf.Marshal(&usageFile{Entries: idx}, qdf.OptBalanced)
	if err != nil {
		return err
	}
	return writeFileLazy(path, data)
}

// Bump records one use of key at nowSec.
func (idx UsageIndex) Bump(key string, nowSec int64) {
	s := idx[key]
	s.Hits++
	s.LastUsed = nowSec
	idx[key] = s
}
