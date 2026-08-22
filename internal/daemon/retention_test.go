package daemon_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/protocol"
)

// TestDispatchBytes_RetainsNothing guards the pooling invariant in pool.go.
//
// The request decoder is zero-copy: tool_input and the tool_response bytes are
// slices of the caller's buffer, which the daemon returns to a pool the moment
// dispatch ends and overwrites with the next request. If any handler kept those
// bytes — in the session state, in a cache entry, in an analytics record — the
// next request would silently rewrite what the last one stored.
//
// So: dispatch a request, scribble over every byte of the buffer, and read the
// state back. Anything that changed was aliased.
func TestDispatchBytes_RetainsNothing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	file := filepath.Join(home, "target.go")
	if err := os.WriteFile(file, []byte("package main\n\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	requests := []map[string]any{
		{
			"session_id":      "retention-session",
			"hook_event_name": "PostToolUse",
			"tool_name":       "Read",
			"tool_input":      map[string]any{"file_path": file},
			"tool_response":   map[string]any{"file": map[string]any{"content": "package main\n\nfunc main() {}\n", "filePath": file}},
		},
		{
			"session_id":      "retention-session",
			"hook_event_name": "PostToolUse",
			"tool_name":       "Bash",
			"tool_input":      map[string]any{"command": "echo retention"},
			"tool_response":   map[string]any{"stdout": strings.Repeat("line of output\n", 40)},
		},
	}

	store := hookcore.NewDiskStore()
	for _, req := range requests {
		body, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}
		// The daemon hands dispatch a slice of a pooled buffer; mimic that with
		// a buffer of our own so it can be overwritten afterwards.
		buf := make([]byte, len(body))
		copy(buf, body)

		var out bytes.Buffer
		if err := hook.DispatchBytes(store, buf, &out); err != nil {
			t.Fatalf("dispatch %s: %v", req["tool_name"], err)
		}
		for i := range buf {
			buf[i] = 'Z' // the next request lands here
		}
	}

	state := store.LoadSession(hook.ContextKey(&protocol.HookInput{SessionID: "retention-session"}))
	if state == nil {
		t.Fatal("no session state was stored")
	}
	for path, entry := range state.Files {
		if strings.Contains(path, "ZZZ") {
			t.Errorf("a stored file path aliased the request buffer: %q", path)
		}
		if bytes.Contains(entry.Content, []byte("ZZZ")) {
			t.Errorf("stored content for %q aliased the request buffer", path)
		}
	}
	if _, ok := state.Files[file]; !ok {
		t.Errorf("the Read entry is missing; stored paths: %v", keysOf(state.Files))
	}
}

func keysOf[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
