package main

import (
	"strings"
	"testing"
)

// The transcript carries several lines whose type is "result": one per hook
// invocation, plus the session's own. Only the last one carries usage, and that
// is the one the report is built from — picking the first "result" line would
// report a session that cost nothing.
func TestParseResult_PicksTheLineCarryingUsage(t *testing.T) {
	stream := strings.Join([]string{
		`{"type":"system","subtype":"hook_started","hook_name":"SessionStart:startup"}`,
		`{"type":"result","duration_ms":12,"uuid":"hook-1"}`,
		`{"type":"assistant","message":{"id":"msg_01","usage":{"input_tokens":10,"cache_read_input_tokens":21820,"output_tokens":3}}}`,
		`{"type":"result","duration_ms":9,"uuid":"hook-2"}`,
		`{"is_error":false,"num_turns":3,"total_cost_usd":0.0494,"usage":{"input_tokens":10,` +
			`"cache_creation_input_tokens":23422,"cache_read_input_tokens":21820,"output_tokens":87},` +
			`"subtype":"success","type":"result","duration_ms":3205}`,
	}, "\n") + "\n"

	res, err := parseResult([]byte(stream))
	if err != nil {
		t.Fatalf("parseResult: %v", err)
	}
	if res.NumTurns != 3 || res.TotalCostUSD != 0.0494 {
		t.Errorf("wrong result line: %+v", res)
	}
	if got, want := res.Usage.Billed(), 10+23422+21820+87; got != want {
		t.Errorf("billed = %d, want %d", got, want)
	}
}

func TestParseResult_RefusesATranscriptWithNoUsage(t *testing.T) {
	stream := `{"type":"result","duration_ms":12,"uuid":"hook-1"}` + "\n"
	if _, err := parseResult([]byte(stream)); err == nil {
		t.Fatal("a transcript with no usage line must be an error, not a zero report")
	}
}

// A line longer than bufio's default 64KB buffer is the normal case, not an edge
// case: an assistant message carries whole tool results.
func TestParseResult_HandlesLongLines(t *testing.T) {
	fat := `{"type":"assistant","message":{"id":"msg_01","content":"` + strings.Repeat("x", 300_000) + `"}}`
	stream := fat + "\n" +
		`{"is_error":false,"num_turns":1,"total_cost_usd":0.01,` +
		`"usage":{"input_tokens":5,"output_tokens":5},"type":"result"}` + "\n"

	res, err := parseResult([]byte(stream))
	if err != nil {
		t.Fatalf("parseResult: %v", err)
	}
	if res.Usage.OutputTokens != 5 {
		t.Errorf("lost the result line after a long one: %+v", res)
	}
}

// Runs arrive interleaved by attempt; the report pairs each task's baseline with
// the treatment run of the same attempt, never across attempts.
func TestPairs_MatchesByTaskAndAttempt(t *testing.T) {
	runs := []Run{
		{Task: "a", Variant: variantOff, Attempt: 1},
		{Task: "a", Variant: variantOn, Attempt: 1},
		{Task: "b", Variant: variantOff, Attempt: 1},
		{Task: "b", Variant: variantOn, Attempt: 1},
		{Task: "a", Variant: variantOff, Attempt: 2},
		{Task: "a", Variant: variantOn, Attempt: 2},
	}
	got := pairs(runs)
	if len(got) != 3 {
		t.Fatalf("paired %d runs, want 3: %+v", len(got), got)
	}
	for _, p := range got {
		if p.off.Task != p.on.Task || p.off.Attempt != p.on.Attempt {
			t.Errorf("mismatched pair: %+v", p)
		}
	}
}

// An unmatched baseline is dropped rather than paired with someone else's run.
func TestPairs_DropsAnUnmatchedBaseline(t *testing.T) {
	runs := []Run{
		{Task: "a", Variant: variantOff, Attempt: 1},
		{Task: "b", Variant: variantOn, Attempt: 1},
	}
	if got := pairs(runs); len(got) != 0 {
		t.Errorf("paired across tasks: %+v", got)
	}
}

func TestPct(t *testing.T) {
	cases := map[string]string{
		pct(100, 80):  "-20.0%",
		pct(100, 130): "+30.0%",
		pct(0, 0):     "-",
		pct(0, 5):     "new",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
