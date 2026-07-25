package cache_test

import (
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/cache"
)

func TestCacheMaxSize(t *testing.T) {
	if got := cache.CacheMaxSize(); got != 128<<20 {
		t.Errorf("default = %d, want %d", got, 128<<20)
	}
	t.Setenv("QDF_CACHE_MAX_SIZE", "1048576")
	if got := cache.CacheMaxSize(); got != 1048576 {
		t.Errorf("env override = %d, want 1048576", got)
	}
	t.Setenv("QDF_CACHE_MAX_SIZE", "garbage")
	if got := cache.CacheMaxSize(); got != 128<<20 {
		t.Errorf("bad env should fall back to default, got %d", got)
	}
}

func TestCacheTTL(t *testing.T) {
	if got := cache.CacheTTL(); got != 720*time.Hour {
		t.Errorf("default = %v, want 720h", got)
	}
	t.Setenv("QDF_CACHE_TTL", "1h30m")
	if got := cache.CacheTTL(); got != 90*time.Minute {
		t.Errorf("env override = %v, want 1h30m", got)
	}
	t.Setenv("QDF_CACHE_TTL", "nonsense")
	if got := cache.CacheTTL(); got != 720*time.Hour {
		t.Errorf("bad env should fall back to default, got %v", got)
	}
}
