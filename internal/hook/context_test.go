package hook

import (
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// agent_id is the documented discriminator: Claude Code sends it only on
// subagent tool calls. It must win over any path heuristic.
func TestContextKeyPrefersAgentID(t *testing.T) {
	const sid = "s1"
	parent := &protocol.HookInput{SessionID: sid, TranscriptPath: "/p/" + sid + ".jsonl"}
	sub := &protocol.HookInput{
		SessionID: sid,
		AgentID:   "9f2c1a",
		// Deliberately a main-thread-looking path: agent_id alone must be enough.
		TranscriptPath: "/p/" + sid + ".jsonl",
	}
	if ContextKey(sub) == ContextKey(parent) {
		t.Fatalf("agent_id was ignored; both contexts got key %q", ContextKey(parent))
	}
	if got, want := ContextKey(sub), sid+"/9f2c1a"; got != want {
		t.Errorf("ContextKey = %q, want %q", got, want)
	}
}

func TestContextKeySeparatesSubagentFromParent(t *testing.T) {
	const sid = "76387586-3078-486c-b9f1-768ce1f4d9a5"
	parent := &protocol.HookInput{
		SessionID:      sid,
		TranscriptPath: "/Users/x/.claude/projects/-proj/" + sid + ".jsonl",
	}
	sub := &protocol.HookInput{
		SessionID:      sid,
		TranscriptPath: "/Users/x/.claude/projects/-proj/" + sid + "/subagents/agent-abc123.jsonl",
	}

	if got := ContextKey(parent); got != sid {
		t.Errorf("ContextKey(parent) = %q, want %q", got, sid)
	}
	if ContextKey(sub) == ContextKey(parent) {
		t.Fatal("a subagent must not share the parent's cache key: its context is " +
			"separate and discarded, so the parent never receives what it read")
	}
	if ContextKey(sub) != ContextKey(sub) {
		t.Error("ContextKey must be deterministic")
	}
}

// Two different subagents of the same parent must not share a key either — one
// subagent's read says nothing about what another one received.
func TestContextKeySeparatesSubagentsFromEachOther(t *testing.T) {
	const sid = "s1"
	base := "/p/" + sid + "/subagents/agent-"
	a := &protocol.HookInput{SessionID: sid, TranscriptPath: base + "aaa.jsonl"}
	b := &protocol.HookInput{SessionID: sid, TranscriptPath: base + "bbb.jsonl"}
	if ContextKey(a) == ContextKey(b) {
		t.Errorf("distinct subagents shared a key: %q", ContextKey(a))
	}
}

// Older Claude Code builds may not send transcript_path. Falling back to the
// plain session id preserves today's behavior rather than disabling the cache.
func TestContextKeyFallsBackWithoutTranscriptPath(t *testing.T) {
	const sid = "abc"
	if got := ContextKey(&protocol.HookInput{SessionID: sid}); got != sid {
		t.Errorf("ContextKey = %q, want %q", got, sid)
	}
}

// A malformed path must not produce an empty or session-colliding key.
func TestContextKeyHandlesMalformedPaths(t *testing.T) {
	const sid = "abc"
	for _, tp := range []string{
		"/p/abc/subagents/",        // marker present, no agent id
		"/p/abc/subagents/.jsonl",  // agent id is empty before the extension
		"/p/abc/subagents/agent-x", // no extension
		"subagents/agent-y.jsonl",  // relative
	} {
		got := ContextKey(&protocol.HookInput{SessionID: sid, TranscriptPath: tp})
		if got == "" {
			t.Errorf("ContextKey(%q) returned an empty key", tp)
		}
	}
}
