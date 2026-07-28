package cache_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
)

// TestRefGet_ExpandRoundTrip covers the register-then-expand path directly via
// RefHashOf/RefPut/RefGet (the store-backed dedup in package hook is what
// production uses; this exercises the underlying content-addressed blob store).
func TestRefGet_ExpandRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := strings.Repeat("expandable payload\n", 30)
	hash := cache.RefHashOf(content)
	if cache.RefSeen(hash) {
		t.Fatal("blob must not exist before RefPut")
	}
	cache.RefPut(hash, content)
	if !cache.RefSeen(hash) {
		t.Fatal("RefSeen must be true after RefPut")
	}
	got, ok := cache.RefGet(hash)
	if !ok {
		t.Fatalf("RefGet(%q) miss", hash)
	}
	if got != content {
		t.Errorf("expanded content mismatch: got %d bytes, want %d", len(got), len(content))
	}
}

func TestRefGet_Missing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if _, ok := cache.RefGet("deadbeefdeadbeefdeadbeefdeadbeef"); ok {
		t.Error("RefGet of unknown hash must miss")
	}
}

func BenchmarkRefPut(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	base := strings.Repeat("fresh output line\n", 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		// Unique content each iteration => a real encode + blob write.
		content := base + strconvItoa(i)
		cache.RefPut(cache.RefHashOf(content), content)
	}
}

func BenchmarkRefGet(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	content := strings.Repeat("payload line\n", 100)
	hash := cache.RefHashOf(content)
	cache.RefPut(hash, content)
	b.ResetTimer()
	for b.Loop() {
		_, _ = cache.RefGet(hash)
	}
}

// strconvItoa avoids importing strconv just for the benchmark label.
func strconvItoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
