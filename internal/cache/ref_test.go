package cache_test

import (
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
)

func TestDedup_RoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := strings.Repeat("line of some command output\n", 40) // >256B

	// First call: not deduped, registers the blob.
	tok, dedup := cache.Dedup(content, 256)
	if dedup || tok != "" {
		t.Fatalf("first call must not dedup, got (%q, %v)", tok, dedup)
	}
	// Second identical call: deduped to a §ref token.
	tok, dedup = cache.Dedup(content, 256)
	if !dedup || !strings.HasPrefix(tok, "§ref:") {
		t.Fatalf("second call must dedup to §ref, got (%q, %v)", tok, dedup)
	}
	// Never-worse: the token must be smaller than the content.
	if len(tok) >= len(content) {
		t.Errorf("§ref token (%d) must be smaller than content (%d)", len(tok), len(content))
	}
}

func TestDedup_BelowMinSize(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	tok, dedup := cache.Dedup("tiny", 256)
	if dedup || tok != "" {
		t.Errorf("sub-minSize content must not dedup, got (%q, %v)", tok, dedup)
	}
}

func TestRefGet_ExpandRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	content := strings.Repeat("expandable payload\n", 30)
	if _, dedup := cache.Dedup(content, 256); dedup {
		t.Fatal("first Dedup should register, not dedup")
	}
	// Recompute the hash the way Dedup did, via a second (deduped) call.
	tok, dedup := cache.Dedup(content, 256)
	if !dedup {
		t.Fatal("expected dedup on second call")
	}
	hash := strings.TrimSuffix(strings.TrimPrefix(strings.Fields(tok)[0], "§ref:"), "§")
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

func BenchmarkDedup_Hit(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	content := strings.Repeat("cached output line\n", 100)
	cache.Dedup(content, 256) // register once
	b.ResetTimer()
	for b.Loop() {
		_, _ = cache.Dedup(content, 256)
	}
}

func BenchmarkDedup_Miss(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	base := strings.Repeat("fresh output line\n", 100)
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		// Unique content each iteration => always a miss + blob write.
		_, _ = cache.Dedup(base+strconvItoa(i), 256)
	}
}

func BenchmarkRefGet(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	content := strings.Repeat("payload line\n", 100)
	cache.Dedup(content, 256)
	tok, _ := cache.Dedup(content, 256)
	hash := strings.TrimSuffix(strings.TrimPrefix(strings.Fields(tok)[0], "§ref:"), "§")
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
