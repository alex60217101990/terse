package hookcore

import (
	"time"

	"github.com/alex60217101990/terse/internal/cache"
)

// diskStore is the on-disk StateStore implementation used by the CLI. It
// delegates every method to the existing internal/cache package, preserving
// today's behavior exactly.
type diskStore struct{}

// NewDiskStore returns the on-disk StateStore backed by internal/cache.
func NewDiskStore() StateStore { return diskStore{} }

func (diskStore) LoadSession(id string) *cache.SessionState {
	s, _ := cache.Load(id)
	return s
}

func (diskStore) SaveSession(id string, s *cache.SessionState) { _ = cache.Save(id, s) }

func (diskStore) RefSeen(h string) bool { return cache.RefSeen(h) }

func (diskStore) RefPut(h, c string) { cache.RefPut(h, c) }

func (diskStore) RefGet(h string) (string, bool) { return cache.RefGet(h) }

// RefHit records a dedup hit against the ref hash. The CLI is one-shot per
// invocation, so this reads-modifies-writes the usage sidecar directly: a
// dedup hit is occasional per invocation, not a hot loop.
func (diskStore) RefHit(h string) {
	idx := cache.LoadUsage(cache.UsageRefsPath())
	idx.Bump(h, time.Now().Unix())
	_ = cache.SaveUsage(cache.UsageRefsPath(), idx)
}

func (diskStore) LastGet(k string) (string, bool) { return cache.LastOutputGet(k) }

func (diskStore) LastPut(k, c string) { cache.LastOutputPut(k, c) }
