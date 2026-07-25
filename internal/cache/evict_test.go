package cache_test

import (
	"maps"
	"math"
	"strconv"
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
)

func TestUtilityScore_RecentHighFreq(t *testing.T) {
	entry := cache.FileEntry{ReadCount: 50, LastReadAt: timeNow() - 3600} // 1h ago
	score := cache.UtilityScore(entry, timeNow())
	// 50 * exp(-0.1 * 1) ≈ 45.2
	if score < 40 || score > 50 {
		t.Errorf("expected ~45, got %f", score)
	}
}

func TestUtilityScore_OldLowFreq(t *testing.T) {
	entry := cache.FileEntry{ReadCount: 1, LastReadAt: timeNow() - 48*3600} // 48h ago
	score := cache.UtilityScore(entry, timeNow())
	// 1 * exp(-0.1 * 48) ≈ 0.0082
	if score > 0.02 {
		t.Errorf("expected near zero, got %f", score)
	}
}

func TestUtilityScore_FreqBeatsAge(t *testing.T) {
	recent := cache.FileEntry{ReadCount: 1, LastReadAt: timeNow() - 3600}
	frequent := cache.FileEntry{ReadCount: 50, LastReadAt: timeNow() - 24*3600}
	if cache.UtilityScore(recent, timeNow()) >= cache.UtilityScore(frequent, timeNow()) {
		t.Error("high-frequency old file should beat low-frequency recent file")
	}
}

func TestEvict_DropsLowScore(t *testing.T) {
	state := cache.NewSessionState()
	now := timeNow()
	// Fill 4 files with good utility
	for i := range 4 {
		path := "/project/used-" + string(rune('a'+i)) + ".go"
		state.Files[path] = cache.FileEntry{ReadCount: 10, LastReadAt: now - 3600}
	}
	// Add one stale file with near-zero utility
	state.Files["/project/stale.go"] = cache.FileEntry{ReadCount: 0, LastReadAt: now - 7*24*3600}
	cache.Evict(state, 4) // maxFiles=4, we have 5
	if _, ok := state.Files["/project/stale.go"]; ok {
		t.Error("stale.go should have been evicted")
	}
}

func TestEvict_NoopUnderLimit(t *testing.T) {
	state := cache.NewSessionState()
	state.Files["/a.go"] = cache.FileEntry{ReadCount: 1}
	cache.Evict(state, 10)
	if len(state.Files) != 1 {
		t.Error("should not evict when under limit")
	}
}

func TestDecayLambda_Default(t *testing.T) {
	t.Setenv("QDF_DECAY_LAMBDA", "")
	if cache.DecayLambda() != 0.1 {
		t.Error("default lambda should be 0.1")
	}
}

func TestDecayLambda_EnvOverride(t *testing.T) {
	t.Setenv("QDF_DECAY_LAMBDA", "0.5")
	if math.Abs(cache.DecayLambda()-0.5) > 1e-9 {
		t.Errorf("expected 0.5, got %f", cache.DecayLambda())
	}
}

func timeNow() int64 {
	return 1753182000 // fixed unix timestamp for deterministic tests
}

func BenchmarkUtilityScore(b *testing.B) {
	entry := cache.FileEntry{ReadCount: 10, LastReadAt: 1753182000 - 3600}
	b.ResetTimer()
	for b.Loop() {
		_ = cache.UtilityScore(entry, 1753182000)
	}
}

func BenchmarkEvict_200Files(b *testing.B) {
	base := cache.NewSessionState()
	now := int64(1753182000)
	for i := range 250 {
		path := "/project/file-" + strconv.Itoa(i) + ".go"
		base.Files[path] = cache.FileEntry{
			ReadCount:  i % 20,
			LastReadAt: now - int64(i*3600),
		}
	}
	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		s := cache.NewSessionState()
		maps.Copy(s.Files, base.Files)
		b.StartTimer()
		cache.Evict(s, 200)
	}
}
