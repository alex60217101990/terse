package daemon

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// clientTempSock returns a short-path unix socket for the test. It
// deliberately avoids t.TempDir(): on darwin the sockaddr_un.sun_path limit is
// 104 bytes, and t.TempDir() embeds the (long) test name, so a long test name
// overruns the limit and net.Listen fails silently. os.MkdirTemp with a short
// prefix keeps the whole path well under 104. Mirrors daemon_test.go's
// tempSock (unexported there, in the external daemon_test package, so it's
// not visible from here).
func clientTempSock(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "qdc")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "s")
}

// waitForClientSock polls for sockPath to exist, up to ~1s.
func waitForClientSock(t *testing.T, sockPath string) {
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

// startTestDaemon starts a real Serve loop in a goroutine on a fresh temp
// socket and returns its path once the socket file exists.
func startTestDaemon(t *testing.T) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	sock := clientTempSock(t)
	go func() { _ = Serve(sock, time.Minute, "client-test") }()
	waitForClientSock(t, sock)
	return sock
}

// readFixture builds a realistic PostToolUse JSON payload (Read tool shape)
// for the given hook name. There is no on-disk fixture for "post" yet, so this
// constructs the payload directly — same shape as daemon_test.go's
// readEventPayload/bashPayload helpers.
func readFixture(t *testing.T, hook string) []byte {
	t.Helper()
	inp := map[string]any{
		"session_id":      "sess-client-test",
		"hook_event_name": "PostToolUse",
		"tool_name":       "Read",
		"tool_input":      map[string]any{"file_path": "/tmp/client_test.go"},
		"tool_response":   map[string]any{"content": strings.Repeat("package x\n", 60)},
	}
	b, err := json.Marshal(inp)
	if err != nil {
		t.Fatalf("marshal %s fixture: %v", hook, err)
	}
	return b
}

// rawRoundtrip dials sockPath, writes payload, half-closes the write side, and
// returns everything read back before the daemon closes the connection.
// Mirrors daemon_test.go's roundtrip (not visible from this package).
func rawRoundtrip(t *testing.T, sockPath string, payload []byte) []byte {
	t.Helper()
	c, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()
	if _, err := c.Write(payload); err != nil {
		t.Fatalf("write: %v", err)
	}
	if uc, ok := c.(*net.UnixConn); ok {
		if err := uc.CloseWrite(); err != nil {
			t.Fatalf("CloseWrite: %v", err)
		}
	}
	out, err := io.ReadAll(c)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return out
}

// TestDialAndProxy_MatchesDaemonReply proves DialAndProxy's proxying is
// byte-exact with a direct raw roundtrip. It exercises two independently
// started daemons (each on its own fresh socket/store) rather than sending
// the same payload twice to one daemon: the hook pipeline is stateful
// (Read's mtime cache, generic dedup) keyed on content across calls, so a
// second identical request to the *same* store legitimately gets a different
// (deduped/"unchanged") reply — that's correct dispatch behavior, not a
// proxying bug, and would make this test flaky for the wrong reason. Two
// fresh stores each see the payload as a first-time request, so any
// difference in bytes can only come from the proxy path itself.
func TestDialAndProxy_MatchesDaemonReply(t *testing.T) {
	sockA := startTestDaemon(t)
	sockB := startTestDaemon(t)
	payload := readFixture(t, "post")

	var got bytes.Buffer
	if err := DialAndProxy(sockA, bytes.NewReader(payload), &got, 2*time.Second); err != nil {
		t.Fatalf("DialAndProxy: %v", err)
	}
	want := rawRoundtrip(t, sockB, payload)
	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("proxied reply != daemon reply\n got: %q\nwant: %q", got.Bytes(), want)
	}
}

func TestDialAndProxy_NoDaemon_ReturnsError(t *testing.T) {
	err := DialAndProxy(filepath.Join(t.TempDir(), "absent.sock"),
		strings.NewReader("{}"), io.Discard, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error dialing a non-existent socket")
	}
}
