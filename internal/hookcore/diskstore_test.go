package hookcore_test

import (
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/hookcore"
)

func TestDiskStore_RoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := hookcore.NewDiskStore()
	st := cache.NewSessionState()
	st.Turn = 3
	s.SaveSession("x", st)
	if got := s.LoadSession("x"); got.Turn != 3 {
		t.Fatalf("turn=%d", got.Turn)
	}
	s.RefPut("abc", "hello")
	if v, ok := s.RefGet("abc"); !ok || v != "hello" {
		t.Fatalf("ref %q %v", v, ok)
	}
}
