package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Per-session isolation is load-bearing: §ref, the Read delta cache and the
// rerun store are all stateful. If one session's state leaks into the next,
// the second session sees phantom cache hits and the report overstates
// savings.
func TestReplayIsolatesSessions(t *testing.T) {
	payload := "alpha bravo charlie delta echo foxtrot golf hotel india juliet\n"
	big := ""
	var bigSb18 strings.Builder
	for range 40 {
		bigSb18.WriteString(payload)
	}
	big += bigSb18.String()
	tr := Triple{Tool: "Bash", Input: []byte(`{"command":"echo hi"}`), Result: big}

	two := []Session{
		{Path: "a.jsonl", Triples: []Triple{tr}},
		{Path: "b.jsonl", Triples: []Triple{tr}},
	}
	rep, err := Replay(two)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	one, err := Replay(two[:1])
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if rep.Total.TokensOut != 2*one.Total.TokensOut {
		t.Fatalf("session state leaked: two sessions emitted %d tokens, "+
			"want exactly twice the single-session %d",
			rep.Total.TokensOut, one.Total.TokensOut)
	}
}

// The inverse of the test above: WITHIN one session the caches must work, or
// the harness would be isolating so hard that it measures nothing. The same
// payload twice in one session is a dedup hit, and the second emission has to
// be far cheaper than the first.
func TestReplayDedupsWithinASession(t *testing.T) {
	big := strings.Repeat("alpha bravo charlie delta echo foxtrot golf hotel india\n", 40)
	tr := Triple{Tool: "Bash", Input: []byte(`{"command":"echo hi"}`), Result: big}

	rep, err := Replay([]Session{{Path: "a.jsonl", Triples: []Triple{tr, tr}}})
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if _, ok := rep.ByHookAction["Bash/ref"]; !ok {
		t.Fatalf("second identical payload in one session did not dedup; categories: %v",
			sortedKeys(rep.ByHookAction))
	}
}

// A replay must never touch the user's own state. Both halves matter: blob
// writes under ~/.qdf-hook would poison a real session's dedup cache, and
// analytics events for tool calls that never happened would corrupt the very
// stats this work exists to make trustworthy.
func TestReplayLeavesRealHomeAlone(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	big := strings.Repeat("alpha bravo charlie delta echo foxtrot golf hotel\n", 40)
	_, err := Replay([]Session{{Path: "a.jsonl", Triples: []Triple{
		{Tool: "Bash", Input: []byte(`{"command":"echo hi"}`), Result: big},
	}}})
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}

	if got := os.Getenv("HOME"); got != home {
		t.Errorf("HOME left at %q, want it restored to %q", got, home)
	}
	if entries, err := os.ReadDir(filepath.Join(home, ".qdf-hook")); err == nil && len(entries) > 0 {
		t.Errorf("replay wrote into the real state dir: %v", entries)
	}
}

// A passthrough emits the ORIGINAL result, not nothing. Scoring it as zero
// output would read as a 100% saving on every payload the tool declined to
// touch — the easiest possible way for this harness to lie.
func TestReplayScoresPassthroughAsNeutral(t *testing.T) {
	// Under the 256-byte floor, so the pipeline passes it through untouched.
	tiny := "ok  github.com/acme/widget  0.312s\n"
	rep, err := Replay([]Session{{Path: "a.jsonl", Triples: []Triple{
		{Tool: "Bash", Input: []byte(`{"command":"go test ./..."}`), Result: tiny},
	}}})
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if rep.Total.TokensIn != rep.Total.TokensOut {
		t.Errorf("passthrough scored %d in / %d out, want them equal",
			rep.Total.TokensIn, rep.Total.TokensOut)
	}
	if rep.Total.TokensOut == 0 {
		t.Error("passthrough scored as zero output — that reads as a 100% saving")
	}
}

func TestCompareBaselineRejectsDifferentCorpus(t *testing.T) {
	base := Report{Fingerprint: Fingerprint{Hash: "aaaa", Files: 2}, ByHookAction: map[string]Cell{}}
	cur := Report{Fingerprint: Fingerprint{Hash: "bbbb", Files: 3}, ByHookAction: map[string]Cell{}}
	var buf bytes.Buffer
	err := compareBaseline(&buf, base, cur)
	if err == nil || !strings.Contains(err.Error(), "fingerprints differ") {
		t.Fatalf("compareBaseline = %v, want a fingerprint mismatch error", err)
	}
}

func TestCompareBaselineFlagsRegressions(t *testing.T) {
	fp := Fingerprint{Hash: "aaaa", Files: 1}
	base := Report{
		Fingerprint:  fp,
		ByHookAction: map[string]Cell{"Bash/summary": {N: 1, TokensIn: 1000, TokensOut: 300}},
		Total:        Cell{N: 1, TokensIn: 1000, TokensOut: 300},
	}

	same := base
	if err := compareBaseline(&bytes.Buffer{}, base, same); err != nil {
		t.Fatalf("identical report reported a regression: %v", err)
	}

	worse := Report{
		Fingerprint:  fp,
		ByHookAction: map[string]Cell{"Bash/summary": {N: 1, TokensIn: 1000, TokensOut: 301}},
		Total:        Cell{N: 1, TokensIn: 1000, TokensOut: 301},
	}
	var buf bytes.Buffer
	if err := compareBaseline(&buf, base, worse); err == nil {
		t.Fatal("a single extra token was not reported; tolerance must be zero")
	}
	if !strings.Contains(buf.String(), "Bash/summary") {
		t.Errorf("the offending row was not printed:\n%s", buf.String())
	}

	better := Report{
		Fingerprint:  fp,
		ByHookAction: map[string]Cell{"Bash/summary": {N: 1, TokensIn: 1000, TokensOut: 250}},
		Total:        Cell{N: 1, TokensIn: 1000, TokensOut: 250},
	}
	if err := compareBaseline(&bytes.Buffer{}, base, better); err != nil {
		t.Fatalf("an improvement was reported as a regression: %v", err)
	}

	// A category that vanishes is a behavior change, not an improvement: the
	// corpus is fixed, so those triples went somewhere else.
	gone := Report{Fingerprint: fp, ByHookAction: map[string]Cell{}, Total: base.Total}
	if err := compareBaseline(&bytes.Buffer{}, base, gone); err == nil {
		t.Error("a disappearing category was not reported")
	}
}

// The committed fixture is the authoritative gate: it is invented, reproducible
// by any contributor, and it is what CI diffs against. If it stops replaying,
// the gate is silently gone.
func TestCommittedFixtureReplays(t *testing.T) {
	corpus, err := LoadSessions(filepath.Join("testdata", "corpus-sessions"))
	if err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}
	if len(corpus.Sessions) < 5 {
		t.Fatalf("fixture has %d sessions, want at least 5", len(corpus.Sessions))
	}
	if corpus.Skipped != 0 || corpus.Unpaired != 0 {
		t.Errorf("fixture is damaged: %d malformed lines, %d unpaired results",
			corpus.Skipped, corpus.Unpaired)
	}
	rep, err := Replay(corpus.Sessions)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if rep.Total.Saved() <= 0 {
		t.Errorf("fixture saved %d tokens; it is meant to exercise the compressing paths",
			rep.Total.Saved())
	}
}

// The recorded baseline is only a gate if it still matches what the fixture
// produces. This is the check CI would otherwise have to shell out for.
func TestBaselineMatchesFixture(t *testing.T) {
	corpus, err := LoadSessions(filepath.Join("testdata", "corpus-sessions"))
	if err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}
	rep, err := Replay(corpus.Sessions)
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	rep.Fingerprint = corpus.Fingerprint
	rep.Skipped, rep.Filtered, rep.Unpaired = corpus.Skipped, corpus.Filtered, corpus.Unpaired

	base, err := readReport(filepath.Join("testdata", "baseline.json"))
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	var buf bytes.Buffer
	if err := compareBaseline(&buf, base, rep); err != nil {
		t.Fatalf("%v\n%s\nre-record with: go run ./cmd/qdf-replay replay "+
			"cmd/qdf-replay/testdata/corpus-sessions --json > cmd/qdf-replay/testdata/baseline.json",
			err, buf.String())
	}
}
