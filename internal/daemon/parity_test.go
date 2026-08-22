package daemon_test

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/daemon"
	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/hookcore"
)

// genericPayload builds a full PostToolUse JSON payload for toolName with the
// given tool_response content, mirroring bashPayload's shape but allowing the
// tool name to vary (Glob, Grep, Read, mcp__*, ...).
func genericPayload(toolName, output string) string {
	inp := map[string]any{
		"session_id":    "sess-parity-1",
		"tool_name":     toolName,
		"tool_input":    map[string]any{"command": "n/a"},
		"tool_response": map[string]any{"content": output},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

// jsonArrayPayloadContent builds a JSON array of >=5 rows so detect.IsJSONArray
// + AnalyzeJSONArray fire and the generic pipeline takes the "summary" path
// instead of falling through to dedup/squeeze.
//
// The "name" column deliberately uses strictly-decreasing per-value
// frequencies (8,7,6,5,4,3,2,1 occurrences for 8 distinct values) rather than
// one unique name per row. detect.AnalyzeJSONArray's top-5-by-frequency
// selection (internal/detect/json.go) ranges over a Go map to build its
// candidate list before sorting by count — with tied counts (e.g. every name
// appearing exactly once) the map's randomized iteration order makes which
// values land in the top 5 non-deterministic *even between two calls in the
// same process*, independent of daemon vs CLI. Giving every value a distinct
// total count removes the tie, so sorting by frequency is unambiguous and
// the summary is reproducible — which is what this parity test needs to
// isolate the daemon-vs-direct question from that unrelated, pre-existing
// nondeterminism (see the task report's Concerns section).
func jsonArrayPayloadContent() string {
	counts := []int{8, 7, 6, 5, 4, 3, 2, 1}
	var sb strings.Builder
	sb.WriteString("[")
	id, first := 0, true
	for name, n := range counts {
		for range n {
			if !first {
				sb.WriteString(",")
			}
			first = false
			sb.WriteString(`{"id":`)
			sb.WriteString(strconv.Itoa(id))
			sb.WriteString(`,"name":"n`)
			sb.WriteString(strconv.Itoa(name))
			sb.WriteString(`","active":true}`)
			id++
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// goTestOutputContent builds a plausible `go test -v` transcript long enough
// (and matching detect.IsGoTestOutput's pattern) to take the go-test summary
// path.
func goTestOutputContent() string {
	var sb strings.Builder
	for i := range 15 {
		name := "TestThing" + strconv.Itoa(i)
		sb.WriteString("=== RUN   ")
		sb.WriteString(name)
		sb.WriteString("\n")
		sb.WriteString("--- PASS: ")
		sb.WriteString(name)
		sb.WriteString(" (0.00s)\n")
	}
	sb.WriteString("PASS\n")
	sb.WriteString("ok  \tgithub.com/alex60217101990/terse/internal/hook\t0.012s\n")
	return sb.String()
}

// plainLogContent builds unstructured repeated-line output, long enough to
// exceed the generic pipeline's minimum size and exercise the
// dedup/squeeze/passthrough fallback tail rather than a structured
// summarizer.
func plainLogContent() string {
	return strings.Repeat("2026-01-01T00:00:00Z INFO some unstructured log line here\n", 30)
}

// globPathListContent builds a newline-delimited file path list long enough
// to make buildGlobTree's compact form smaller than the raw list.
func globPathListContent() string {
	var sb strings.Builder
	dirs := []string{"internal/hook", "internal/cache", "internal/hookcore", "cmd/qdf-hook"}
	for _, d := range dirs {
		for i := range 8 {
			sb.WriteString(d)
			sb.WriteString("/file")
			sb.WriteString(strconv.Itoa(i))
			sb.WriteString(".go\n")
		}
	}
	return sb.String()
}

// grepMatchesContent builds file:line:text grep output long enough for
// buildGrepSummary to produce a smaller grouped form.
func grepMatchesContent() string {
	var sb strings.Builder
	files := []string{"internal/hook/dispatch.go", "internal/hook/bash.go", "internal/hook/glob.go"}
	for _, f := range files {
		for i := 1; i <= 10; i++ {
			sb.WriteString(f)
			sb.WriteString(":")
			sb.WriteString(strconv.Itoa(i))
			sb.WriteString(":\tsome matching line of source code here\n")
		}
	}
	return sb.String()
}

// readFileContent builds file content long enough to exercise handleRead's
// pipeline (it never needs the path to exist on disk: on a session's first
// read of a path, handleRead passes the content through unchanged regardless
// of whether os.Stat succeeds, so a synthetic path is fine for parity).
func readFileContent() string {
	var sb strings.Builder
	for i := range 50 {
		sb.WriteString("line ")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(": some file content for the parity test\n")
	}
	return sb.String()
}

// readPayload is like genericPayload but with a Read-shaped tool_input
// ({"file_path": ...}), which handleRead requires to extract the path.
func readPayload(path, content string) string {
	inp := map[string]any{
		"session_id":    "sess-parity-1",
		"tool_name":     "Read",
		"tool_input":    map[string]any{"file_path": path},
		"tool_response": map[string]any{"content": content},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

// parityCase is one row of the differential parity table.
type parityCase struct {
	name    string
	payload string
	// silent marks a payload whose correct reply is no reply at all. Pass-through
	// writes nothing so Claude Code records no hook_success attachment for it, so
	// "" is a real answer here, not a malformed fixture.
	silent bool
}

func parityCases() []parityCase {
	return []parityCase{
		{name: "Bash_JSONArray", payload: genericPayload("Bash", jsonArrayPayloadContent())},
		{name: "Bash_GoTestOutput", payload: genericPayload("Bash", goTestOutputContent())},
		{name: "Bash_PlainLog", payload: genericPayload("Bash", plainLogContent())},
		{name: "Glob_PathList", payload: genericPayload("Glob", globPathListContent())},
		{name: "Grep_FileLineText", payload: genericPayload("Grep", grepMatchesContent())},
		// A file's first sighting has nothing cached to compare against, so both
		// sides correctly pass it through in silence.
		{"Read_FileContent", readPayload("/tmp/qdf-parity-fixture.txt", readFileContent()), true},
		{name: "MCP_JSONArray", payload: genericPayload("mcp__x__y", jsonArrayPayloadContent())},
	}
}

// directDispatch runs hook.Dispatch against a fresh on-disk store rooted at a
// fresh HOME, matching the CLI's own "post" subcommand path exactly
// (hookcore.NewDiskStore(), os.Stdin, os.Stdout in cmd/qdf-hook/main.go).
func directDispatch(t *testing.T, payload string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	var buf bytes.Buffer
	if err := hook.Dispatch(hookcore.NewDiskStore(), strings.NewReader(payload), &buf); err != nil {
		t.Fatalf("direct hook.Dispatch: %v", err)
	}
	return buf.String()
}

// daemonDispatch starts a fresh daemon (its own fresh HOME, so its MemStore's
// eventual disk flush can never touch the direct side's state) on a fresh
// socket, sends payload exactly once, and returns the reply.
func daemonDispatch(t *testing.T, payload string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)
	go func() { _ = daemon.Serve(sock, time.Minute, "test") }()
	waitForSock(t, sock)
	return roundtrip(t, sock, payload)
}

// TestParity_DaemonMatchesDirectDispatch is the differential parity test:
// for each representative payload, the daemon's reply (over the socket, via
// its MemStore) must be byte-identical to calling hook.Dispatch directly
// against a fresh diskStore. Both sides see completely fresh, isolated state
// (fresh $HOME each) and each payload touches that state exactly once, so
// neither path can emit a §ref/§unchanged/§delta marker that the other
// wouldn't also emit — a legitimate divergence here is a real bug, not a
// state artifact.
func TestParity_DaemonMatchesDirectDispatch(t *testing.T) {
	for _, tc := range parityCases() {
		t.Run(tc.name, func(t *testing.T) {
			// t.Setenv forbids parallel subtests sharing a process-wide env
			// var across goroutines, so these run sequentially — fine, this
			// is a correctness test, not a benchmark.
			direct := directDispatch(t, tc.payload)
			viaDaemon := daemonDispatch(t, tc.payload)

			if direct != viaDaemon {
				t.Fatalf("daemon reply diverges from direct hook.Dispatch for %s\n--- direct ---\n%s\n--- daemon ---\n%s",
					tc.name, direct, viaDaemon)
			}
			if tc.silent {
				if direct != "" {
					t.Fatalf("%s: expected a silent pass-through, got %q", tc.name, direct)
				}
				return
			}
			if direct == "" {
				t.Fatalf("%s: both sides produced an empty reply (payload likely malformed)", tc.name)
			}
		})
	}
}
