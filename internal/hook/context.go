package hook

import (
	"strings"

	"github.com/alex60217101990/terse/internal/protocol"
)

// subagentMarker is the path segment Claude Code uses for subagent transcripts:
// <project>/<session-id>/subagents/agent-<id>.jsonl, where the main thread's
// transcript is <project>/<session-id>.jsonl. Only consulted when agent_id is
// absent.
const subagentMarker = "/subagents/"

// ContextKey returns the state-store key for one hook invocation.
//
// It is deliberately NOT the session id. Claude Code reports a subagent's tool
// calls under the PARENT's session_id, but a subagent has its own context that
// is thrown away when it finishes — nothing it reads ever reaches the parent.
// Keying the file cache on session_id alone therefore lets a subagent's read
// deny the parent's, handing back "no re-read needed" for content the parent
// never saw. That is the one failure this cache must not have: a deny cannot be
// downgraded to a safe passthrough, because there is no content in hand.
//
// Splitting the key keeps each context honest, and still lets a subagent dedup
// against its own earlier reads.
func ContextKey(inp *protocol.HookInput) string {
	if inp.AgentID != "" {
		return inp.SessionID + "/" + inp.AgentID
	}
	if id := agentFromTranscript(inp.TranscriptPath); id != "" {
		return inp.SessionID + "/" + id
	}
	// Main thread, or a Claude Code build sending neither field. Falling back to
	// the session id preserves the existing behavior rather than disabling the
	// cache outright.
	return inp.SessionID
}

// agentFromTranscript extracts a subagent identifier from a transcript path,
// returning "" for a main-thread path. It is a fallback for builds that do not
// send agent_id, so it errs toward "" — a wrong split would scatter a single
// context's cache across many keys, which costs tokens but never correctness.
func agentFromTranscript(path string) string {
	_, after, ok := strings.Cut(path, subagentMarker)
	if !ok {
		return ""
	}
	id := after
	if j := strings.IndexByte(id, '/'); j >= 0 {
		id = id[:j]
	}
	if j := strings.IndexByte(id, '.'); j >= 0 {
		id = id[:j]
	}
	return id
}
