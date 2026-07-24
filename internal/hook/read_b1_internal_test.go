package hook

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/hookcore"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// TestHandleRead_PartialReadNotCached is the B1 regression: a windowed read
// (offset/limit) must NOT be cached as the file's entry, so a later full read
// isn't diffed against the window and emitted as a bogus "§delta — changes
// since last read". Before the fix, the window's content+hash poisoned the
// cache slot and the next full read produced a spurious delta.
func TestHandleRead_PartialReadNotCached(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	store := hookcore.NewDiskStore()
	const path = "/tmp/qdf-b1-does-not-exist.go"

	mk := func(limit int, content string) *protocol.HookInput {
		ti, _ := json.Marshal(protocol.ReadInput{FilePath: path, Limit: limit})
		return &protocol.HookInput{
			SessionID:    "sess-b1",
			ToolInput:    ti,
			ToolResponse: &protocol.ToolResponse{Content: content},
		}
	}

	// 1) Windowed read (limit set) → must pass through and not cache.
	var o1 strings.Builder
	if err := handleRead(store, mk(100, strings.Repeat("window line\n", 50)), &o1); err != nil {
		t.Fatalf("handleRead (windowed): %v", err)
	}
	var r1 protocol.HookOutput
	_ = json.Unmarshal([]byte(o1.String()), &r1)
	if r1.HookSpecificOutput != nil {
		t.Fatalf("windowed read must pass through, got replacement: %q", r1.HookSpecificOutput.UpdatedToolOutput)
	}

	// 2) Full read (no limit) of the same file+session → first-touch full
	//    passthrough, NOT a delta against the (uncached) window.
	var o2 strings.Builder
	if err := handleRead(store, mk(0, strings.Repeat("full different content\n", 50)), &o2); err != nil {
		t.Fatalf("handleRead (full): %v", err)
	}
	var r2 protocol.HookOutput
	_ = json.Unmarshal([]byte(o2.String()), &r2)
	if r2.HookSpecificOutput != nil && strings.Contains(r2.HookSpecificOutput.UpdatedToolOutput, "§delta") {
		t.Fatalf("full read after a windowed read must not be a bogus delta: %q",
			r2.HookSpecificOutput.UpdatedToolOutput)
	}
}

// TestHandleRead_CachesFileContent proves handleRead reads the Read tool's
// nested file.content (not the empty top-level Content). Before the protocol
// fix it cached "" and Read compression was a no-op on live Claude output.
func TestHandleRead_CachesFileContent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	store := hookcore.NewDiskStore()
	const path = "/tmp/qdf-filecontent-x.go"
	fileContent := strings.Repeat("package main\n", 60)
	ti, _ := json.Marshal(protocol.ReadInput{FilePath: path})
	inp := &protocol.HookInput{
		SessionID: "sess-file",
		ToolInput: ti,
		ToolResponse: &protocol.ToolResponse{
			File: &protocol.FileResponse{
				Content: fileContent, FilePath: path,
				StartLine: 1, NumLines: 60, TotalLines: 60,
			},
		},
	}
	var out strings.Builder
	if err := handleRead(store, inp, &out); err != nil {
		t.Fatalf("handleRead: %v", err)
	}
	st := store.LoadSession("sess-file")
	if got := string(st.Files[path].Content); got != fileContent {
		t.Fatalf("Read must cache file.content (len %d), cached len %d", len(fileContent), len(got))
	}
}
