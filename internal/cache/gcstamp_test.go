package cache_test

import (
	"testing"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/cache"
)

func TestShouldRunGC_Throttle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	now := time.Now().Unix()
	if !cache.ShouldRunGC(now) {
		t.Fatal("first call (no stamp) should run")
	}
	if cache.ShouldRunGC(now + 3600) { // 1h later
		t.Error("within 24h should NOT run again")
	}
	if !cache.ShouldRunGC(now + 25*3600) { // 25h later
		t.Error("after 24h should run again")
	}
}
