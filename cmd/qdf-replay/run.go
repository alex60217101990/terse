package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alex60217101990/terse/internal/analytics"
	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/protocol"
	"github.com/alex60217101990/terse/internal/tokens"
)

// Cell is one row of the report: how many triples fell into a category and
// what they cost before and after the hook ran.
type Cell struct {
	N         int `json:"n"`
	TokensIn  int `json:"tokens_in"`
	TokensOut int `json:"tokens_out"`
}

// Saved is the token difference. Negative means the hook made the output more
// expensive than the raw tool result.
func (c Cell) Saved() int { return c.TokensIn - c.TokensOut }

// SavedPct is Saved as a percentage of TokensIn.
func (c Cell) SavedPct() float64 {
	if c.TokensIn == 0 {
		return 0
	}
	return float64(c.Saved()) / float64(c.TokensIn) * 100
}

func (c *Cell) add(in, out int) {
	c.N++
	c.TokensIn += in
	c.TokensOut += out
}

// Report is one replay's result: what corpus it ran over, what the walk had to
// throw away, and the per-category token ledger.
type Report struct {

	// ByHookAction is keyed "<tool>/<action>", e.g. "Bash/summary".
	//
	// A per-category ledger rather than a single total, because the total hides
	// the failure that matters: a change that improves one detector by more
	// than it breaks another looks like a win in aggregate and is not one.
	ByHookAction map[string]Cell `json:"by_hook_action"`
	Fingerprint  Fingerprint     `json:"fingerprint"`

	Total Cell `json:"total"`

	Sessions int `json:"sessions"`
	Triples  int `json:"triples"`
	Skipped  int `json:"skipped"`
	Filtered int `json:"filtered"`
	Unpaired int `json:"unpaired"`
}

// Replay runs every session's triples through the hook pipeline and reports the
// token cost before and after.
//
// Each session gets a FRESH store rooted in its own temporary HOME. That
// isolation is load-bearing rather than tidy: §ref dedup, the Read delta cache
// and the rerun-delta store are all stateful, so state leaking from one session
// into the next shows up as phantom cache hits and an overstated saving. The
// same redirection is what keeps a replay from reading or writing the real
// ~/.qdf-hook — including its analytics.jsonl, which a replay must never
// pollute with events for tool calls that are not happening.
func Replay(sessions []Session) (Report, error) {
	rep := Report{
		Sessions:     len(sessions),
		ByHookAction: make(map[string]Cell),
	}

	root, err := os.MkdirTemp("", "qdf-replay-")
	if err != nil {
		return rep, fmt.Errorf("temp store root: %w", err)
	}
	defer func() { _ = os.RemoveAll(root) }()

	// A long-lived analytics Writer (what the daemon installs) would hold an fd
	// on the REAL analytics.jsonl and ignore the HOME redirection below. Nothing
	// in this process installs one, but asserting it costs nothing and the
	// failure mode — silently appending replay events to the user's own
	// analytics — is invisible until the stats look wrong.
	analytics.SetWriter(nil)

	defer captureHome()()

	for i, sess := range sessions {
		home := filepath.Join(root, fmt.Sprintf("session-%04d", i))
		if err := os.MkdirAll(home, 0o700); err != nil {
			return rep, fmt.Errorf("session store: %w", err)
		}
		if err := os.Setenv("HOME", home); err != nil {
			return rep, fmt.Errorf("redirect HOME: %w", err)
		}
		if err := replaySession(&rep, sess, i); err != nil {
			return rep, err
		}
	}
	return rep, nil
}

// captureHome saves HOME and returns a function restoring it. Everything the
// hook writes is rooted at os.UserHomeDir, which is the seam a replay uses to
// stay out of the user's real state; leaving HOME pointed at a deleted temp
// directory afterwards would break whatever ran next in the same process.
func captureHome() func() {
	old, had := os.LookupEnv("HOME")
	return func() {
		if had {
			_ = os.Setenv("HOME", old)
			return
		}
		_ = os.Unsetenv("HOME")
	}
}

func replaySession(rep *Report, sess Session, idx int) error {
	store := hookcore.NewMemStore()
	// A synthetic id, not the transcript's own: session state is keyed by it,
	// and the id is the only thing tying one triple's cache entry to the next.
	// Deriving it from the index keeps a replay reproducible even if two
	// archives contain a file of the same name.
	sessionID := fmt.Sprintf("replay-%04d", idx)
	tail := newActionTail()

	var buf bytes.Buffer
	for _, tr := range sess.Triples {
		req, err := buildRequest(sessionID, tr)
		if err != nil {
			return fmt.Errorf("%s: encode %s request: %w", sess.Path, tr.Tool, err)
		}
		buf.Reset()
		if err := hook.DispatchBytes(store, req, &buf); err != nil {
			return fmt.Errorf("%s: dispatch %s: %w", sess.Path, tr.Tool, err)
		}
		emitted, err := emittedOutput(buf.Bytes(), tr.Result)
		if err != nil {
			return fmt.Errorf("%s: decode %s response: %w", sess.Path, tr.Tool, err)
		}

		in := tokens.Count(tr.Result)
		out := tokens.Count(emitted)
		key := tr.Tool + "/" + tail.next()

		cell := rep.ByHookAction[key]
		cell.add(in, out)
		rep.ByHookAction[key] = cell
		rep.Total.add(in, out)
		rep.Triples++
	}
	return nil
}

// replayRequest is the PostToolUse payload shape the hook decodes. It is built
// here rather than reused from protocol.HookInput because that type is written
// for DECODING: its ToolResponse has a custom UnmarshalJSON and no matching
// marshaller, so round-tripping it would not produce the JSON the hook expects.
type replayRequest struct {
	SessionID     string          `json:"session_id"`
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
	ToolResponse  json.RawMessage `json:"tool_response"`
}

func buildRequest(sessionID string, tr Triple) ([]byte, error) {
	resp, err := toolResponse(tr)
	if err != nil {
		return nil, err
	}
	input := tr.Input
	if len(input) == 0 {
		input = json.RawMessage("null")
	}
	return json.Marshal(replayRequest{
		SessionID:     sessionID,
		HookEventName: "PostToolUse",
		ToolName:      tr.Tool,
		ToolInput:     input,
		ToolResponse:  resp,
	})
}

// toolResponse rebuilds the tool_response object for a recorded result.
//
// Read is special-cased because protocol.ToolResponse.Text() resolves the Read
// tool's output from a nested "file" object, and handleRead reads that object's
// window metadata to decide whether the read was partial. Putting a Read result
// under the generic "content" key would route it down a path it never took.
// StartLine/NumLines/TotalLines are deliberately left zero: the transcript does
// not record them, and a guessed window would fabricate a partial read.
func toolResponse(tr Triple) (json.RawMessage, error) {
	if tr.Tool == "Read" {
		var ti struct {
			FilePath string `json:"file_path"`
		}
		_ = json.Unmarshal(tr.Input, &ti)
		return json.Marshal(map[string]any{
			"file": map[string]any{
				"content":  tr.Result,
				"filePath": ti.FilePath,
			},
		})
	}
	return json.Marshal(map[string]any{"content": tr.Result})
}

// emittedOutput returns what the model would actually have seen: the hook's
// replacement, or the original result when the hook passed through.
//
// A passthrough emits the ORIGINAL, not nothing. Scoring it as zero output
// would read as a 100% saving on every payload the tool declined to touch,
// which is the single easiest way to make this harness lie.
func emittedOutput(raw []byte, original string) (string, error) {
	var out protocol.HookOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if out.HookSpecificOutput != nil && out.HookSpecificOutput.UpdatedToolOutput != "" {
		return out.HookSpecificOutput.UpdatedToolOutput, nil
	}
	return original, nil
}

// actionTail reads the action label back out of the analytics events the hook
// just wrote.
//
// The pipeline does not return which branch it took — it writes that to
// analytics and returns only the replacement text. Rather than re-deriving the
// action by pattern-matching the output (which would drift from the pipeline
// the moment a marker changes), the replay reads the label the pipeline itself
// recorded. HOME is redirected per session, so the file being read is the
// replay's own, inside the temp store.
type actionTail struct {
	path string
	off  int64
}

func newActionTail() *actionTail { return &actionTail{path: analytics.AnalyticsPath()} }

// next returns the action of the most recent event appended since the previous
// call, or "none" when the dispatch recorded nothing (a nil tool_response, or a
// handler that returns before its record site).
func (a *actionTail) next() string {
	f, err := os.Open(a.path)
	if err != nil {
		return "none"
	}
	defer f.Close()

	// Record rotates the log past 10 MB by renaming it aside. A replay session
	// never gets close, but a shrunken file means the offset is stale and
	// seeking to it would read garbage or nothing at all.
	if info, err := f.Stat(); err == nil && info.Size() < a.off {
		a.off = 0
	}
	if _, err := f.Seek(a.off, io.SeekStart); err != nil {
		return "none"
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return "none"
	}
	a.off += int64(len(data))

	action := "none"
	for line := range strings.SplitSeq(string(data), "\n") {
		if line == "" {
			continue
		}
		var e analytics.Event
		if json.Unmarshal([]byte(line), &e) != nil || e.Action == "" {
			continue
		}
		action = e.Action
	}
	return action
}
