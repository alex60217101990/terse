package hook_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/hookcore"
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
	f, _ := os.CreateTemp(t.TempDir(), "qdf-pre-*.go")
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
	// Seed through the real write path so the cached entry carries a genuine
	// CtimeNS, exactly as production does — a deny requires a real ctime match.
	_, sid, path := seedCachedFile(t, "package main\n")

	var out strings.Builder
	err := hook.HandlePreToolUse(strings.NewReader(makePreToolInput(t, sid, path)), &out)
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
	f, _ := os.CreateTemp(t.TempDir(), "qdf-pre-*.go")
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

// --- ctime deny-gate hardening (closes the mtime+size staleness window) ---

// fileTimes is a snapshot of a file's mtime (nanoseconds) and size, taken
// before a test forges the mtime back.
type fileTimes struct {
	mtimeNS int64
	size    int64
}

// statFileTimes stats path and returns its mtime and size.
func statFileTimes(t *testing.T, path string) fileTimes {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s): %v", path, err)
	}
	return fileTimes{mtimeNS: info.ModTime().UnixNano(), size: info.Size()}
}

// seedCachedFile creates a temp file with content and runs it through the
// real Write pipeline (hook.Dispatch) so the resulting cache.FileEntry —
// Hash, ModTime, and CtimeNS — is populated exactly as production code would
// populate it on a genuine Write/Edit. Returns the store, session id, and
// file path for the caller to drive further hook calls against.
func seedCachedFile(t *testing.T, content string) (store hookcore.StateStore, sid, path string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	store = hookcore.NewDiskStore()

	f, err := os.CreateTemp(t.TempDir(), "qdf-pre-ctime-*.go")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	path = f.Name()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })

	sid = "sess-ctime-" + t.Name()
	var out strings.Builder
	if err := hook.Dispatch(store, strings.NewReader(makeWriteInput(t, sid, path, content)), &out); err != nil {
		t.Fatalf("Dispatch (seed write): %v", err)
	}
	return store, sid, path
}

// runPreToolUse drives a PreToolUse Read event for path through hook.Dispatch
// against store/sid, and returns the resulting permissionDecision ("allow" or
// "deny").
func runPreToolUse(t *testing.T, store hookcore.StateStore, sid, path string) string {
	t.Helper()
	inp := map[string]any{
		"session_id":      sid,
		"tool_name":       "Read",
		"hook_event_name": "PreToolUse",
		"tool_input":      map[string]any{"file_path": path},
	}
	b, err := json.Marshal(inp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out strings.Builder
	if err := hook.Dispatch(store, strings.NewReader(string(b)), &out); err != nil {
		t.Fatalf("Dispatch (pretooluse): %v", err)
	}
	var resp map[string]any
	_ = json.Unmarshal([]byte(out.String()), &resp)
	hso, _ := resp["hookSpecificOutput"].(map[string]any)
	dec, _ := hso["permissionDecision"].(string)
	return dec
}

// overwriteSameSize rewrites path with content, which must be exactly the
// same byte length as the file's current content — this changes the file's
// ctime (and, on most filesystems, its mtime) while leaving size unchanged,
// so the caller can then forge mtime back to simulate cp -p / touch -r.
func overwriteSameSize(t *testing.T, path, content string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%s): %v", path, err)
	}
	if int64(len(content)) != info.Size() {
		t.Fatalf("overwriteSameSize: new content is %d bytes, want %d (same size as original)", len(content), info.Size())
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// zeroCachedCtime forces the cached FileEntry.CtimeNS for path to 0, so tests
// can verify a pre-upgrade cache entry (written before CtimeNS existed)
// degrades to mtime+size instead of forcing a re-read.
func zeroCachedCtime(t *testing.T, store hookcore.StateStore, sid, path string) {
	t.Helper()
	state := store.LoadSession(sid)
	entry, ok := state.Files[path]
	if !ok {
		t.Fatalf("zeroCachedCtime: no cached entry for %s in session %s", path, sid)
	}
	entry.CtimeNS = 0
	state.Files[path] = entry
	store.SaveSession(sid, state)
}

// A file re-touched to a preserved mtime+size but changed content must NOT be
// denied — the new ctime reveals the change even though mtime was rewound.
func TestPreToolUse_CtimePreservedMtimeChangedContent_Allows(t *testing.T) {
	store, sid, path := seedCachedFile(t, "package a\nfunc F(){}\n") // caches entry incl. CtimeNS
	old := statFileTimes(t, path)                                    // {mtimeNS, size}

	// Change content to the SAME size, then rewind mtime to the cached value.
	overwriteSameSize(t, path, "package a\nfunc G(){}\n")        // len unchanged
	_ = os.Chtimes(path, time.Time{}, time.Unix(0, old.mtimeNS)) // forge mtime back

	dec := runPreToolUse(t, store, sid, path)
	if dec == "deny" {
		t.Error("content changed under preserved mtime+size must be allowed (ctime moved), got deny")
	}
}

// A genuinely unchanged file is still denied (no regression).
func TestPreToolUse_TrulyUnchanged_Denies(t *testing.T) {
	store, sid, path := seedCachedFile(t, "package a\nfunc F(){}\n")
	if dec := runPreToolUse(t, store, sid, path); dec != "deny" {
		t.Errorf("unchanged file must be denied, got %q", dec)
	}
}

// A deny asserts "unchanged, don't re-read" with no content in hand, so it is
// the one path that cannot be downgraded to a safe passthrough if wrong. When
// ctime is unavailable (a pre-CtimeNS cache entry, or a platform without it) a
// deny would rest on mtime+size alone — which cp -p / touch -r can forge across
// a same-size content change, serving stale content as authoritative. So an
// unavailable ctime must NOT deny; it falls through to allow (the PostToolUse
// handler then makes a content-based, never-worse decision). One extra read of
// a stale-cache file, never stale content.
func TestPreToolUse_ZeroCachedCtime_Allows(t *testing.T) {
	store, sid, path := seedCachedFile(t, "package a\nfunc F(){}\n")
	zeroCachedCtime(t, store, sid, path) // set entry.CtimeNS = 0
	if dec := runPreToolUse(t, store, sid, path); dec == "deny" {
		t.Errorf("zero cached ctime must not deny (no content-safe recovery), got %q", dec)
	}
}

// The stale-content hazard the above guards against: zero cached ctime, then a
// same-size content change with mtime forged back. Must be allowed, not denied.
func TestPreToolUse_ZeroCtimeStaleContent_Allows(t *testing.T) {
	store, sid, path := seedCachedFile(t, "package a\nfunc F(){}\n")
	old := statFileTimes(t, path)
	zeroCachedCtime(t, store, sid, path)
	overwriteSameSize(t, path, "package a\nfunc G(){}\n")
	_ = os.Chtimes(path, time.Time{}, time.Unix(0, old.mtimeNS)) // forge mtime back
	if dec := runPreToolUse(t, store, sid, path); dec == "deny" {
		t.Errorf("stale content under zero ctime + forged mtime must be allowed, got %q", dec)
	}
}
