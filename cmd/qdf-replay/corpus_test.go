package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Archived transcripts already contain hook-compressed output. Replaying those
// through the hook again double-compresses them and manufactures a fake win,
// so they must be excluded from the corpus.
func TestIsHookOutput(t *testing.T) {
	compressed := []string{
		"§ref:abc123§ (7179 bytes, identical to earlier output)",
		"[READ §unchanged:deadbeef§ /tmp/x.go — content identical]",
		"⟦↑ repeat: \"connecting\"⟧",
		"[repeat: \"connecting\"]",
		"[repeat of \"connecting\" except 3->4]",
		"§P=/Users/dev/work/src/acme§",
		"[^=/Users/dev/work/src/acme]",
		"[TABLE 40 rows × 5 cols]",
		"[grep: 12 matches in 3 files]",
		"[full output: qdf-hook expand abc123]",
		"[expand abc123456789]",
	}
	for _, s := range compressed {
		if !IsHookOutput(s) {
			t.Errorf("IsHookOutput(%q) = false, want true", s)
		}
	}

	raw := []string{
		"total 48\ndrwxr-xr-x  5 user staff  160 Aug  7 12:00 .",
		"ok  \tgithub.com/acme/widget\t0.312s",
		"",
	}
	for _, s := range raw {
		if IsHookOutput(s) {
			t.Errorf("IsHookOutput(%q) = true, want false", s)
		}
	}
}

// The recovery footer is appended at the END of a lossy summary, and a table
// or grep summary routinely runs past a head-only scan window. A filter that
// only looked at the head would let exactly the lossiest outputs — the ones
// that would double-compress most dramatically — back into the corpus.
func TestIsHookOutputFindsTrailingRecoveryFooter(t *testing.T) {
	body := strings.Repeat("2026-08-07T10:00:00Z INFO worker=7 handled request\n", 400)
	if len(body) <= scanWindow {
		t.Fatalf("test premise broken: body must exceed the head window (%d)", scanWindow)
	}
	s := body + "[full output: qdf-hook expand a1b2c3d4e5f60718]\n"
	if !IsHookOutput(s) {
		t.Error("IsHookOutput missed a recovery footer past the head scan window")
	}
}

// LoadSessions must pair a tool_result with the tool_use that produced it, and
// keep both the session boundary and the transcript order — the pipeline is
// stateful in both.
func TestLoadSessionsPairsAndOrders(t *testing.T) {
	dir := t.TempDir()
	writeTranscript(t, filepath.Join(dir, "a.jsonl"),
		toolUse("tu_1", "Bash", `{"command":"ls /srv/app"}`),
		toolResult("tu_1", "bin\nconf\nlogs"),
		toolUse("tu_2", "Read", `{"file_path":"/srv/app/main.go"}`),
		toolResult("tu_2", "package main\n"),
	)
	writeTranscript(t, filepath.Join(dir, "b.jsonl"),
		toolUse("tu_3", "Grep", `{"pattern":"handler"}`),
		toolResult("tu_3", "/srv/app/http.go:12:func handler()"),
	)

	c, err := LoadSessions(dir)
	if err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}
	if len(c.Sessions) != 2 {
		t.Fatalf("got %d sessions, want 2", len(c.Sessions))
	}
	if got := c.Triples(); got != 3 {
		t.Fatalf("got %d triples, want 3", got)
	}

	first := c.Sessions[0].Triples
	if first[0].Tool != "Bash" || first[1].Tool != "Read" {
		t.Errorf("transcript order lost: got %q then %q", first[0].Tool, first[1].Tool)
	}
	if first[0].Result != "bin\nconf\nlogs" {
		t.Errorf("result mispaired: got %q", first[0].Result)
	}
	var in struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(first[1].Input, &in); err != nil || in.FilePath != "/srv/app/main.go" {
		t.Errorf("tool input lost: %s (%v)", first[1].Input, err)
	}
}

// A corrupt archive must not look identical to a clean one. Three distinct
// failures are counted separately because they mean different things: a bad
// line means the file is damaged, a filtered triple means the archive is
// already compressed, and an unpaired result means the transcript is truncated.
func TestLoadSessionsCountsSkippedFilteredAndUnpaired(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mixed.jsonl")
	lines := []string{
		toolUse("tu_1", "Bash", `{"command":"echo hi"}`),
		toolResult("tu_1", "hi"),
		`{"type":"assistant","message":{`, // truncated JSON
		`not json at all`,
		toolUse("tu_2", "Bash", `{"command":"cat log"}`),
		toolResult("tu_2", "§ref:abc123§ (7179 bytes, identical to earlier output)"),
		toolResult("tu_missing", "orphaned result"),
		`{"type":"attachment","content":"meta line, not malformed"}`,
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	c, err := LoadSessions(dir)
	if err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}
	if c.Skipped != 2 {
		t.Errorf("Skipped = %d, want 2", c.Skipped)
	}
	if c.Filtered != 1 {
		t.Errorf("Filtered = %d, want 1", c.Filtered)
	}
	if c.Unpaired != 1 {
		t.Errorf("Unpaired = %d, want 1", c.Unpaired)
	}
	if got := c.Triples(); got != 1 {
		t.Errorf("Triples = %d, want 1", got)
	}
}

// Two reports taken over different corpora are not comparable, so --baseline
// refuses to diff them. The fingerprint is what it checks, and it has to move
// when the archive does.
func TestFingerprintChangesWithCorpus(t *testing.T) {
	dir := t.TempDir()
	writeTranscript(t, filepath.Join(dir, "a.jsonl"),
		toolUse("tu_1", "Bash", `{"command":"ls"}`),
		toolResult("tu_1", "bin\nconf"),
	)
	before, err := LoadSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	again, err := LoadSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	if before.Fingerprint != again.Fingerprint {
		t.Fatalf("fingerprint is not stable: %+v vs %+v", before.Fingerprint, again.Fingerprint)
	}

	writeTranscript(t, filepath.Join(dir, "b.jsonl"),
		toolUse("tu_2", "Bash", `{"command":"pwd"}`),
		toolResult("tu_2", "/srv/app"),
	)
	after, err := LoadSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	if after.Fingerprint.Hash == before.Fingerprint.Hash {
		t.Error("fingerprint did not change after a file was added to the corpus")
	}
	if after.Fingerprint.Files != 2 {
		t.Errorf("Files = %d, want 2", after.Fingerprint.Files)
	}
}

// A single transcript line legitimately runs to many megabytes when a tool
// dumped a large result. bufio.Scanner would error out and abandon the rest of
// the file, silently shrinking the corpus; the reader must not.
func TestLoadSessionsReadsOversizedLine(t *testing.T) {
	dir := t.TempDir()
	huge := strings.Repeat("alpha bravo charlie delta echo foxtrot\n", 40000) // ~1.5 MB
	writeTranscript(t, filepath.Join(dir, "big.jsonl"),
		toolUse("tu_1", "Bash", `{"command":"cat big.log"}`),
		toolResult("tu_1", huge),
		toolUse("tu_2", "Bash", `{"command":"echo done"}`),
		toolResult("tu_2", "done"),
	)
	c, err := LoadSessions(dir)
	if err != nil {
		t.Fatalf("LoadSessions: %v", err)
	}
	if got := c.Triples(); got != 2 {
		t.Fatalf("Triples = %d, want 2 — the tail of the file was dropped", got)
	}
	if c.Sessions[0].Triples[0].Result != huge {
		t.Error("oversized result was truncated")
	}
}

// --- helpers ---

func writeTranscript(t *testing.T, path string, lines ...string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func toolUse(id, name, input string) string {
	b, err := json.Marshal(map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"role": "assistant",
			"content": []any{map[string]any{
				"type": "tool_use", "id": id, "name": name,
				"input": json.RawMessage(input),
			}},
		},
	})
	if err != nil {
		panic(err)
	}
	return string(b)
}

func toolResult(id, content string) string {
	b, err := json.Marshal(map[string]any{
		"type": "user",
		"message": map[string]any{
			"role": "user",
			"content": []any{map[string]any{
				"type": "tool_result", "tool_use_id": id, "content": content,
			}},
		},
	})
	if err != nil {
		panic(err)
	}
	return string(b)
}
