package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/cache"
)

func TestSweepCache_EvictsOverCap(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("QDF_CACHE_MAX_SIZE", "2000")
	refs := cache.RefsDir()
	_ = os.MkdirAll(refs, 0o755)
	for _, n := range []string{"a", "b", "c"} {
		_ = os.WriteFile(filepath.Join(refs, n+".blob"), make([]byte, 1000), 0o600)
	}
	sweepCache(time.Now().Unix())
	left, _ := os.ReadDir(refs)
	if len(left) >= 3 {
		t.Errorf("sweep should have evicted at least one blob, %d remain", len(left))
	}
}
