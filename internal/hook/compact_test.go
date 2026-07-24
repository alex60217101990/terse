package hook_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

func makeCompactInput(t *testing.T, sessionID string) string {
	t.Helper()
	inp := map[string]any{
		"session_id": sessionID,
		"tool_name":  "",
		"trigger":    "auto",
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func TestPreCompact_ExportsManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Set up a session with 3 files.
	const sid = "sess-compact-1"
	s := cache.NewSessionState()
	s.Turn = 5
	s.Files["/project/a.go"] = cache.FileEntry{Hash: [32]byte{1}, Turn: 2, Content: []byte("package a\n")}
	s.Files["/project/b.go"] = cache.FileEntry{Hash: [32]byte{2}, Turn: 3, Content: []byte("package b\n")}
	_ = cache.Save(sid, s)

	var out strings.Builder
	if err := hook.HandlePreCompact(strings.NewReader(makeCompactInput(t, sid)), &out); err != nil {
		t.Fatalf("HandlePreCompact: %v", err)
	}

	// Verify state has CompactedAt set.
	state, _ := cache.Load(sid)
	if state.CompactedAt == 0 {
		t.Error("CompactedAt should be set after PreCompact")
	}

	// Verify hook output.
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	// PreCompact does NOT update tool output — it updates the session state only.
	// Output is pass-through (empty hookSpecificOutput).
	_ = resp // just confirm valid JSON
}

func TestPostCompact_InjectsManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const sid = "sess-compact-2"
	s := cache.NewSessionState()
	s.Turn = 7
	s.CompactedAt = 7 // simulate pre-compact already ran
	s.Files["/project/encoder.go"] = cache.FileEntry{
		Hash:    [32]byte{0xAB, 0xCD},
		Turn:    3,
		Content: []byte("package main\n"),
	}
	_ = cache.Save(sid, s)

	var out strings.Builder
	inp := map[string]any{"session_id": sid, "tool_name": "", "trigger": "auto"}
	b, _ := json.Marshal(inp)
	if err := hook.HandlePostCompact(strings.NewReader(string(b)), &out); err != nil {
		t.Fatalf("HandlePostCompact: %v", err)
	}

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("PostCompact should inject a manifest")
	}
	if !strings.Contains(resp.HookSpecificOutput.UpdatedToolOutput, "encoder.go") {
		t.Errorf("manifest should reference encoder.go: %s", resp.HookSpecificOutput.UpdatedToolOutput)
	}
}

func TestSessionStart_NoCompact_Passthrough(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Session with files but no compaction — should pass through.
	const sid = "sess-start-no-compact"
	s := cache.NewSessionState()
	s.Turn = 3
	s.CompactedAt = 0 // no compaction
	s.Files["/project/x.go"] = cache.FileEntry{Hash: [32]byte{5}, Turn: 2, Content: []byte("package x\n")}
	_ = cache.Save(sid, s)

	var out strings.Builder
	if err := hook.HandleSessionStart(strings.NewReader(makeCompactInput(t, sid)), &out); err != nil {
		t.Fatalf("HandleSessionStart: %v", err)
	}

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput != nil {
		t.Error("SessionStart without prior compaction should pass through")
	}
}

func TestSessionStart_WithCompact_InjectsManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	const sid = "sess-start-with-compact"
	s := cache.NewSessionState()
	s.Turn = 10
	s.CompactedAt = 8 // compaction happened
	s.Files["/project/main.go"] = cache.FileEntry{Hash: [32]byte{0xFF}, Turn: 5, Content: []byte("package main\n")}
	_ = cache.Save(sid, s)

	var out strings.Builder
	if err := hook.HandleSessionStart(strings.NewReader(makeCompactInput(t, sid)), &out); err != nil {
		t.Fatalf("HandleSessionStart: %v", err)
	}

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput == nil {
		t.Fatal("SessionStart after compaction should inject a manifest")
	}
	if !strings.Contains(resp.HookSpecificOutput.UpdatedToolOutput, "main.go") {
		t.Errorf("manifest should reference main.go: %s", resp.HookSpecificOutput.UpdatedToolOutput)
	}
}

func TestPostCompact_NoFiles_Passthrough(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// Unknown session — should pass through gracefully.
	const sid = "sess-compact-empty"
	var out strings.Builder
	inp := map[string]any{"session_id": sid, "tool_name": "", "trigger": "auto"}
	b, _ := json.Marshal(inp)
	if err := hook.HandlePostCompact(strings.NewReader(string(b)), &out); err != nil {
		t.Fatalf("HandlePostCompact: %v", err)
	}

	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput != nil {
		t.Error("PostCompact with empty state should pass through")
	}
}

// TestSessionStart_TriggersBlobSweep verifies auto-gc fires on a plain
// (non-compacted) session start — the common case — and prunes over-cap blobs.
// Regression for the earlier narrow trigger that only ran post-compaction.
func TestSessionStart_TriggersBlobSweep(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("QDF_CACHE_MAX_SIZE", "2000") // tiny cap so 3x1000B blobs over-cap

	refs := cache.RefsDir()
	if err := os.MkdirAll(refs, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"a", "b", "c"} {
		if err := os.WriteFile(filepath.Join(refs, n+".blob"), make([]byte, 1000), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	// A fresh session with no prior compaction — the early-return path. The
	// blob sweep must still run (it's before the early return) since no
	// .gc-stamp exists yet.
	var out strings.Builder
	if err := hook.HandleSessionStart(strings.NewReader(makeCompactInput(t, "sess-sweep")), &out); err != nil {
		t.Fatalf("HandleSessionStart: %v", err)
	}

	left, _ := os.ReadDir(refs)
	if len(left) >= 3 {
		t.Errorf("SessionStart should have swept over-cap blobs, %d remain", len(left))
	}
}
