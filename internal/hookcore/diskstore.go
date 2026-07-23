package hookcore

import "github.com/alex60217101990/qdf-hook/internal/cache"

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

func (diskStore) LastGet(k string) (string, bool) { return cache.LastOutputGet(k) }

func (diskStore) LastPut(k, c string) { cache.LastOutputPut(k, c) }
