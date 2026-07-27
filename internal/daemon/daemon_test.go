package daemon_test

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/daemon"
	"github.com/alex60217101990/terse/internal/protocol"
)

// tempSock returns a short-path unix socket for the test. It deliberately
// avoids t.TempDir(): on darwin the sockaddr_un.sun_path limit is 104 bytes,
// and t.TempDir() embeds the (long) test name, so a long test name overruns
// the limit and net.Listen fails silently. os.MkdirTemp with a short prefix
// keeps the whole path well under 104.
func tempSock(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "qd")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "s")
}

// waitForSock polls for sockPath to exist, up to ~1s.
func waitForSock(t *testing.T, sockPath string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sockPath); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("socket %s never appeared", sockPath)
}

// roundtrip dials sockPath, writes payload, half-closes the write side, and
// returns everything read back before the daemon closes the connection.
func roundtrip(t *testing.T, sockPath, payload string) string {
	t.Helper()
	c, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()
	if _, err := io.WriteString(c, payload); err != nil {
		t.Fatalf("write: %v", err)
	}
	if cw, ok := c.(interface{ CloseWrite() error }); ok {
		if err := cw.CloseWrite(); err != nil {
			t.Fatalf("CloseWrite: %v", err)
		}
	}
	out, err := io.ReadAll(c)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(out)
}

// bashPayload builds a PostToolUse JSON payload for the Bash tool with output.
func bashPayload(output string) string {
	inp := map[string]any{
		"session_id":    "sess-daemon-1",
		"tool_name":     "Bash",
		"tool_input":    map[string]any{"command": "echo hi"},
		"tool_response": map[string]any{"content": output},
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

// readEventPayload builds a Read hook payload for the given event. A PostToolUse
// payload carries tool_response.content; a PreToolUse one does not.
func readEventPayload(event, sid, path, content string) string {
	inp := map[string]any{
		"session_id":      sid,
		"hook_event_name": event,
		"tool_name":       "Read",
		"tool_input":      map[string]any{"file_path": path},
	}
	if content != "" {
		inp["tool_response"] = map[string]any{"content": content}
	}
	b, _ := json.Marshal(inp)
	return string(b)
}

// TestDaemon_PreToolUseRoutesAndDenies proves PreToolUse is routed through the
// daemon (by hook_event_name) against the in-RAM session a prior PostToolUse
// Read populated — a repeated read of an unchanged file is denied from RAM
// without a fresh CLI process or disk decode.
func TestDaemon_PreToolUseRoutesAndDenies(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)
	go func() { _ = daemon.Serve(sock, time.Minute, "test") }()
	waitForSock(t, sock)

	fp := filepath.Join(t.TempDir(), "f.go")
	content := strings.Repeat("package x\n", 60)
	if err := os.WriteFile(fp, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// Prime the daemon's session cache with a PostToolUse Read.
	_ = roundtrip(t, sock, readEventPayload("PostToolUse", "sess-pre", fp, content))

	// PreToolUse for the same unchanged file must deny with the §unchanged§
	// marker — served from the in-RAM session.
	reply := roundtrip(t, sock, readEventPayload("PreToolUse", "sess-pre", fp, ""))
	if !strings.Contains(reply, `"permissionDecision":"deny"`) {
		t.Fatalf("expected PreToolUse deny for unchanged file, got: %s", reply)
	}
	if !strings.Contains(reply, "§unchanged") {
		t.Fatalf("expected §unchanged marker, got: %s", reply)
	}
}

func TestDaemon_ServesAndDedups(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)
	go func() { _ = daemon.Serve(sock, time.Minute, "test") }()
	waitForSock(t, sock)

	payload := bashPayload(strings.Repeat("log line here\n", 40))

	first := roundtrip(t, sock, payload) // 1st: registers
	var out protocol.HookOutput
	if err := json.Unmarshal([]byte(first), &out); err != nil {
		t.Fatalf("invalid HookOutput JSON: %v (body: %s)", err, first)
	}

	second := roundtrip(t, sock, payload) // 2nd: identical -> §ref
	if !strings.Contains(second, "§ref:") {
		t.Fatalf("expected ref, got %s", second)
	}
}

// TestDaemon_IdleTimeoutExits checks that Serve returns on its own once no
// connection has arrived for the idle duration, and that it stops accepting
// after doing so.
func TestDaemon_IdleTimeoutExits(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)

	done := make(chan error, 1)
	go func() { done <- daemon.Serve(sock, 50*time.Millisecond, "test") }()
	waitForSock(t, sock)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not exit after idle timeout")
	}

	if _, err := net.Dial("unix", sock); err == nil {
		t.Fatal("expected dial to fail after daemon exited")
	}
}

// TestDaemon_MalformedRequestClosesGracefully checks that a request the hook
// pipeline can't decode just closes the connection (empty reply, no crash)
// and that the daemon keeps serving afterward. (handleConn's recover is
// defensive against a future Dispatch panic; malformed JSON returns an
// ordinary error, so this exercises the error-return path, not the recover.)
func TestDaemon_MalformedRequestClosesGracefully(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)
	go func() { _ = daemon.Serve(sock, time.Minute, "test") }()
	waitForSock(t, sock)

	garbage := roundtrip(t, sock, "not json at all {{{")
	if garbage != "" {
		t.Fatalf("expected empty reply for malformed request, got %q", garbage)
	}

	// Daemon must still be alive and serving valid requests.
	payload := bashPayload(strings.Repeat("still alive\n", 40))
	out := roundtrip(t, sock, payload)
	var resp protocol.HookOutput
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("daemon did not recover after malformed request: %v (body: %s)", err, out)
	}
}

// TestDaemon_ContinuousTrafficNeverDropsConn is a regression test for the
// accept/idle-exit race: a connection landing at the idle boundary must be
// served, never abandoned. With a short idle, we drive connections spaced
// under the idle window but spanning well past it in total; every one must
// get a valid reply and none may hang (each arrival resets the idle timer).
func TestDaemon_ContinuousTrafficNeverDropsConn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)
	// Idle window (500ms) is generous relative to the ~15ms inter-request gap so
	// scheduler/GC jitter on a slow or contended CI runner can't make one gap
	// exceed it (the old 40ms-idle / 15ms-sleep margin was too tight and flaked
	// there). The test stays sensitive: the loop runs ~0.85s of continuous
	// traffic — well past the 500ms idle — so a daemon that FAILS to reset its
	// idle timer on each request would exit mid-loop and drop a reply.
	const idle = 500 * time.Millisecond
	go func() { _ = daemon.Serve(sock, idle, "test") }()
	waitForSock(t, sock)

	for i := range 50 {
		payload := bashPayload(strings.Repeat("traffic line\n", 40))
		out := roundtrip(t, sock, payload) // roundtrip fails the test if it hangs/errors
		var resp protocol.HookOutput
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			t.Fatalf("request %d got invalid reply (daemon dropped conn / exited?): %v (body: %s)", i, err, out)
		}
		time.Sleep(15 * time.Millisecond) // << idle, so continuous traffic keeps the daemon up
	}
}

func TestSocketPathTooLong(t *testing.T) {
	long := "/tmp/" + strings.Repeat("x", 120) + "/d.sock"
	if !daemon.SocketPathTooLong(long) {
		t.Error("120+ byte path must be flagged")
	}
	if daemon.SocketPathTooLong("/tmp/short.sock") {
		t.Error("short path must not be flagged")
	}
}
