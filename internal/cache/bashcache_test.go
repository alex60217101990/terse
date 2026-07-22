package cache_test

import (
	"testing"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/cache"
)

func TestBashCacheRoundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cache.BashCacheSet("git status", "/project", "nothing to commit")
	got, ok := cache.BashCacheGet("git status", "/project")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got != "nothing to commit" {
		t.Errorf("got %q", got)
	}
}

func TestBashCacheMiss_DifferentCwd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cache.BashCacheSet("git status", "/project-a", "clean")
	_, ok := cache.BashCacheGet("git status", "/project-b")
	if ok {
		t.Error("different cwd should be a miss")
	}
}

func TestBashCacheMiss_Expired(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("QDF_BASH_CACHE_TTL_SEC", "0") // expire immediately
	cache.BashCacheSet("git status", "/p", "old")
	time.Sleep(10 * time.Millisecond)
	_, ok := cache.BashCacheGet("git status", "/p")
	if ok {
		t.Error("expired entry should be a miss")
	}
}

func TestIsReadOnlyCommand(t *testing.T) {
	yes := []string{
		"git status",
		"git log --oneline",
		"ls -la",
		"find . -name *.go",
		"go env GOPATH",
	}
	no := []string{
		"git commit",
		"rm -rf",
		"go build",
		"npm install",
	}
	for _, cmd := range yes {
		if !cache.IsReadOnlyCommand(cmd) {
			t.Errorf("expected read-only: %q", cmd)
		}
	}
	for _, cmd := range no {
		if cache.IsReadOnlyCommand(cmd) {
			t.Errorf("expected NOT read-only: %q", cmd)
		}
	}
}
