package hook

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/protocol"
)

// TestTryRerunDelta_LargeInputSkipsDiff guards the O((N+M)²) Myers diff: a
// rerun whose combined line count exceeds the cap must return (,false) via the
// size guard, WITHOUT running UnifiedDiff (which would risk a hang/OOM that
// takes down every daemon connection). Returns fast — no diff computed.
func TestTryRerunDelta_LargeInputSkipsDiff(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	store := hookcore.NewMemStore().StateStore()

	var prev, cur strings.Builder
	for i := range 6000 { // 6000 + 6000 = 12000 > 10000 cap
		prev.WriteString("prev line ")
		prev.WriteString(strings.Repeat("x", i%7))
		prev.WriteByte('\n')
		cur.WriteString("cur line ")
		cur.WriteString(strings.Repeat("y", i%5))
		cur.WriteByte('\n')
	}
	key := "k"
	store.LastPut(key, prev.String())

	if out, ok := tryRerunDelta(store, "Bash", key, cur.String()); ok || out != "" {
		t.Fatalf("large rerun must skip the diff (guard), got ok=%v len(out)=%d", ok, len(out))
	}
}

// TestHandleWrite_NeverWorse checks the guard: when the compressed marker is
// not actually shorter than the response (a long file path on a small
// response), the hook passes the response through instead of growing it.
func TestHandleWrite_NeverWorse(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	store := hookcore.NewMemStore().StateStore()

	longPath := "/" + strings.Repeat("very/long/nested/dir/", 30) + "file_does_not_exist.go"
	ti, _ := json.Marshal(protocol.ReadInput{FilePath: longPath})
	inp := &protocol.HookInput{
		SessionID: "s",
		ToolInput: ti,
		ToolResponse: &protocol.ToolResponse{
			// Just over the 256-byte threshold so we enter the compress path,
			// but far shorter than the marker built from the long path.
			Content: strings.Repeat("x", 300),
		},
	}

	var out strings.Builder
	if err := handleWrite(store, inp, &out); err != nil {
		t.Fatalf("handleWrite: %v", err)
	}
	var resp protocol.HookOutput
	_ = json.Unmarshal([]byte(out.String()), &resp)
	if resp.HookSpecificOutput != nil {
		t.Fatalf("marker (%d+ bytes via long path) exceeds 300-byte content — must pass through, got replacement:\n%s",
			len(longPath), resp.HookSpecificOutput.UpdatedToolOutput)
	}
}
