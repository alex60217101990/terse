package cache

import (
	"os"
	"strconv"
	"time"
)

// DefaultCacheMaxSize caps the total bytes of refs/ + last/ before eviction.
const DefaultCacheMaxSize = 128 << 20 // 128 MiB

// DefaultCacheTTL drops cache entries older than this regardless of size.
const DefaultCacheTTL = 720 * time.Hour // 30 days

// CacheMaxSize is the refs/+last/ size cap in bytes: env QDF_CACHE_MAX_SIZE, or
// DefaultCacheMaxSize when unset/unparseable.
func CacheMaxSize() int64 {
	if v := os.Getenv("QDF_CACHE_MAX_SIZE"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	return DefaultCacheMaxSize
}

// CacheTTL is the cache entry TTL: env QDF_CACHE_TTL (a Go duration like
// "720h"), or DefaultCacheTTL when unset/unparseable.
func CacheTTL() time.Duration {
	if v := os.Getenv("QDF_CACHE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return DefaultCacheTTL
}
