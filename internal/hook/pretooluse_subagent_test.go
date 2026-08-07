package hook_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/hook"
)

func makePreToolInputForAgent(t *testing.T, sessionID, agentID, path string) string {
	t.Helper()
	inp := map[string]any{
		"session_id": sessionID,
		"agent_id":   agentID,
		"tool_name":  "Read",
		"tool_input": map[string]any{"file_path": path},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

func decidePreToolUse(t *testing.T, payload string) (decision, reason string) {
	t.Helper()
	var out strings.Builder
	if err := hook.HandlePreToolUse(strings.NewReader(payload), &out); err != nil {
		t.Fatalf("HandlePreToolUse: %v", err)
	}
	var resp map[string]any
	if err := json.Unmarshal([]byte(out.String()), &resp); err != nil {
		t.Fatalf("decode response %q: %v", out.String(), err)
	}
	hso, ok := resp["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("no hookSpecificOutput in %q", out.String())
	}
	d, _ := hso["permissionDecision"].(string)
	r, _ := hso["permissionDecisionReason"].(string)
	return d, r
}

// The defect this guards: Claude Code reports a subagent's tool calls under the
// PARENT's session_id, but a subagent's context is separate and discarded, so
// nothing the parent read ever reached it. Serving the parent's cache to a
// subagent denies it a file it has never seen — and a deny cannot be downgraded
// to a passthrough, because there is no content in hand.
func TestPreToolUse_SubagentDoesNotInheritParentCache(t *testing.T) {
	_, sid, path := seedCachedFile(t, "package main\n\nfunc main() {}\n")

	// Sanity: the parent, which really did read it, is still denied. Without
	// this the test could pass simply because caching broke entirely.
	if d, _ := decidePreToolUse(t, makePreToolInput(t, sid, path)); d != "deny" {
		t.Fatalf("parent should still be denied for its own cached read, got %q", d)
	}

	d, r := decidePreToolUse(t, makePreToolInputForAgent(t, sid, "agent-9f2c1a", path))
	if d != "allow" {
		t.Fatalf("a subagent was denied a file only the PARENT had read; the "+
			"subagent never received that content. decision=%q reason=%q", d, r)
	}
}

// Two subagents of the same parent are separate contexts too.
func TestPreToolUse_SubagentsDoNotShareCache(t *testing.T) {
	_, sid, path := seedCachedFile(t, "package main\n\nfunc main() {}\n")
	if d, _ := decidePreToolUse(t, makePreToolInputForAgent(t, sid, "agent-aaa", path)); d != "allow" {
		t.Fatalf("agent-aaa should be allowed, got %q", d)
	}
	if d, _ := decidePreToolUse(t, makePreToolInputForAgent(t, sid, "agent-bbb", path)); d != "allow" {
		t.Fatalf("agent-bbb should be allowed, got %q", d)
	}
}

// A deny is the only lossy path that cannot fall back to a passthrough, so it
// must tell the reader how to obtain the content. It previously ended with
// "No re-read needed." and offered nothing.
func TestPreToolUse_DenyIsRecoverable(t *testing.T) {
	const content = "package main\n\nfunc main() { println(\"hello\") }\n"
	_, sid, path := seedCachedFile(t, content)

	d, reason := decidePreToolUse(t, makePreToolInput(t, sid, path))
	if d != "deny" {
		t.Fatalf("expected deny, got %q", d)
	}
	const marker = "qdf-hook expand "
	i := strings.Index(reason, marker)
	if i < 0 {
		t.Fatalf("deny reason offers no way to recover the content: %q", reason)
	}
	hash := strings.TrimSpace(reason[i+len(marker):])

	got, ok := cache.RefGet(hash)
	if !ok {
		t.Fatalf("hash %q from the deny reason does not resolve in the ref store", hash)
	}
	if got != content {
		t.Errorf("recovered content mismatch:\n got %q\nwant %q", got, content)
	}
}
