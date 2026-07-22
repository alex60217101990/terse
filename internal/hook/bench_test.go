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

func BenchmarkReadHook_Unchanged(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	content, _ := os.ReadFile("../cache/testdata/encoder_go_v1.txt")
	raw := makeReadInput(b, "bench-sess", "/project/encoder.go", string(content))

	// Warm up the cache with one read so subsequent iterations hit the unchanged path.
	_ = hook.HandleRead(strings.NewReader(raw), &strings.Builder{})

	b.ResetTimer()
	for b.Loop() {
		var out strings.Builder
		_ = hook.HandleRead(strings.NewReader(raw), &out)
	}
}

func BenchmarkReadHook_FirstRead(b *testing.B) {
	b.Setenv("HOME", b.TempDir()) // once, before loop
	content, _ := os.ReadFile("../cache/testdata/encoder_go_v1.txt")
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		// Unique session ID per iteration gives a fresh (empty) cache each time.
		raw := makeReadInput(b, "bench-fresh-"+strconv.Itoa(i), "/project/encoder.go", string(content))
		var out strings.Builder
		_ = hook.HandleRead(strings.NewReader(raw), &out)
	}
}

func BenchmarkPreToolUse_Unchanged(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	// Create a real file and populate cache.
	f, _ := os.CreateTemp("", "bench-pre-*.go")
	content := []byte("package main\n")
	f.Write(content)
	f.Close()
	defer os.Remove(f.Name())

	info, _ := os.Stat(f.Name())
	hash := sha256.Sum256(content)
	s := cache.NewSessionState()
	s.Turn = 5
	s.Files[f.Name()] = cache.FileEntry{
		Hash: hash, Turn: 5, Content: content,
		ModTime: info.ModTime().UnixNano(),
	}
	_ = cache.Save("bench-pre", s)

	inp := map[string]any{
		"session_id": "bench-pre",
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

func BenchmarkBashHook_JSONArray(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	// HandleBash is stateless (no session cache), so a fixed session_id is safe.
	// encoder_go_v1.txt lives in internal/cache/testdata, not root testdata.
	jsonData, _ := os.ReadFile("../../testdata/json_array_1k.json")
	inp := map[string]any{
		"session_id":    "bench-bash",
		"tool_name":     "Bash",
		"tool_input":    map[string]any{"command": "cat data.json"},
		"tool_response": map[string]any{"content": string(jsonData)},
	}
	bs, _ := json.Marshal(inp)
	raw := string(bs)
	b.ResetTimer()
	for b.Loop() {
		var out strings.Builder
		_ = hook.HandleBash(strings.NewReader(raw), &out)
	}
}
