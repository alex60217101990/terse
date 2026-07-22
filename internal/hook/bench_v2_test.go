package hook_test

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/hook"
)

// BenchmarkPreToolUse_Allow — no cache entry, so the interceptor always allows.
func BenchmarkPreToolUse_Allow(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	f, _ := os.CreateTemp("", "bench-allow-*.go")
	f.WriteString("package main\n")
	f.Close()
	defer os.Remove(f.Name())
	inp := map[string]any{
		"session_id": "bench-allow",
		"tool_name":  "Read",
		"tool_input": map[string]any{"file_path": f.Name()},
	}
	bs, _ := json.Marshal(inp)
	raw := string(bs)
	b.ResetTimer()
	for b.Loop() {
		var out strings.Builder
		_ = hook.HandlePreToolUse(strings.NewReader(raw), &out)
	}
}

// BenchmarkGlob_Tree — compress a 50-file Glob listing to a tree.
func BenchmarkGlob_Tree(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	var files []string
	for i := range 50 {
		files = append(files, "internal/hook/file"+strconv.Itoa(i)+".go")
	}
	content := strings.Join(files, "\n")
	inp := map[string]any{
		"session_id":    "bench-glob",
		"tool_name":     "Glob",
		"tool_input":    map[string]any{"pattern": "**/*.go"},
		"tool_response": map[string]any{"content": content},
	}
	bs, _ := json.Marshal(inp)
	raw := string(bs)
	b.ResetTimer()
	for b.Loop() {
		var out strings.Builder
		_ = hook.HandleGlob(strings.NewReader(raw), &out)
	}
}

// BenchmarkWrite_Compress — suppress the content echo of a ~500-byte Write.
func BenchmarkWrite_Compress(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	content := strings.Repeat("func foo() { return }\n", 25)
	inp := map[string]any{
		"session_id":    "bench-write",
		"tool_name":     "Write",
		"tool_input":    map[string]any{}, // no file_path: measures compression path only, bypasses cache I/O
		"tool_response": map[string]any{"content": content},
	}
	bs, _ := json.Marshal(inp)
	raw := string(bs)
	b.ResetTimer()
	for b.Loop() {
		var out strings.Builder
		_ = hook.HandleWrite(strings.NewReader(raw), &out)
	}
}

// BenchmarkReadHook_PreToolIntercept — full PreToolUse cycle with a warm cache entry.
func BenchmarkReadHook_PreToolIntercept(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	f, _ := os.CreateTemp("", "bench-intercept-*.go")
	content := []byte(strings.Repeat("package main\n", 20))
	f.Write(content)
	f.Close()
	defer os.Remove(f.Name())

	info, _ := os.Stat(f.Name())
	hash := sha256.Sum256(content)
	s := cache.NewSessionState()
	s.Turn = 10
	s.Files[f.Name()] = cache.FileEntry{
		Hash: hash, Turn: 10, Content: content,
		ModTime: info.ModTime().UnixNano(), ReadCount: 5, LastReadAt: 1753182000,
	}
	_ = cache.Save("bench-intercept", s)

	inp := map[string]any{
		"session_id": "bench-intercept",
		"tool_name":  "Read",
		"tool_input": map[string]any{"file_path": f.Name()},
	}
	bs, _ := json.Marshal(inp)
	raw := string(bs)
	b.ResetTimer()
	for b.Loop() {
		var out strings.Builder
		_ = hook.HandlePreToolUse(strings.NewReader(raw), &out)
	}
}
