package hook_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/hook"
)

func makePreToolInput(t *testing.T, sessionID, path string) string {
	t.Helper()
	inp := map[string]any{
		"session_id": sessionID,
		"tool_name":  "Read",
		"tool_input": map[string]any{"file_path": path},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func TestPreToolUse_AllowsOnFirstRead(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create a real temp file.
	f, _ := os.CreateTemp("", "qdf-pre-*.go")
	f.WriteString("package main\n")
	f.Close()
	defer os.Remove(f.Name())

	var out strings.Builder
	err := hook.HandlePreToolUse(strings.NewReader(makePreToolInput(t, "sess-pre-1", f.Name())), &out)
	if err != nil {
		t.Fatalf("HandlePreToolUse: %v", err)
	}
	var resp map[string]any
	_ = json.Unmarshal([]byte(out.String()), &resp)
	hso := resp["hookSpecificOutput"].(map[string]any)
	if hso["permissionDecision"] != "allow" {
		t.Errorf("first read should be allow, got: %v", hso["permissionDecision"])
	}
}

func TestPreToolUse_DeniesUnchanged(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Create temp file.
	f, _ := os.CreateTemp("", "qdf-pre-*.go")
	content := []byte("package main\n")
	f.Write(content)
	f.Close()
	defer os.Remove(f.Name())

	// Pre-populate cache with current mtime.
	info, _ := os.Stat(f.Name())
	s := cache.NewSessionState()
	s.Turn = 2
	s.Files[f.Name()] = cache.FileEntry{
		Hash:    [32]byte{1},
		Turn:    2,
		Content: content,
		ModTime: info.ModTime().UnixNano(),
	}
	_ = cache.Save("sess-pre-2", s)

	var out strings.Builder
	err := hook.HandlePreToolUse(strings.NewReader(makePreToolInput(t, "sess-pre-2", f.Name())), &out)
	if err != nil {
		t.Fatalf("HandlePreToolUse: %v", err)
	}
	var resp map[string]any
	_ = json.Unmarshal([]byte(out.String()), &resp)
	hso := resp["hookSpecificOutput"].(map[string]any)
	if hso["permissionDecision"] != "deny" {
		t.Errorf("cached unchanged file should be deny, got: %v", hso["permissionDecision"])
	}
	reason, _ := hso["permissionDecisionReason"].(string)
	if !strings.Contains(reason, "§unchanged") {
		t.Errorf("reason should contain §unchanged, got: %s", reason)
	}
}

func TestPreToolUse_AllowsAfterModtime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	f, _ := os.CreateTemp("", "qdf-pre-*.go")
	f.WriteString("package main\n")
	f.Close()
	defer os.Remove(f.Name())

	// Cache with OLD mtime.
	s := cache.NewSessionState()
	s.Turn = 1
	s.Files[f.Name()] = cache.FileEntry{
		Turn:    1,
		ModTime: time.Now().Add(-time.Hour).UnixNano(), // old mtime
	}
	_ = cache.Save("sess-pre-3", s)

	var out strings.Builder
	_ = hook.HandlePreToolUse(strings.NewReader(makePreToolInput(t, "sess-pre-3", f.Name())), &out)
	var resp map[string]any
	_ = json.Unmarshal([]byte(out.String()), &resp)
	hso := resp["hookSpecificOutput"].(map[string]any)
	if hso["permissionDecision"] != "allow" {
		t.Errorf("changed mtime should be allow, got: %v", hso["permissionDecision"])
	}
}
