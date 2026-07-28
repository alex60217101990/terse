package daemon_test

import (
	"bytes"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/daemon"
)

// TestServe_FlushesSessionCompletedDuringShutdownDrain pins the post-drain
// flush ordering. A connection whose handler dispatches (and SaveSessions)
// AFTER the accept loop has already torn down must still be flushed to disk.
//
// The pre-fix code flushed at each loop-exit site and only then ran a deferred
// wg.Wait(), so a handler finishing during the drain was never persisted; this
// test fails deterministically on that code (session file absent).
//
// Determinism: connection A writes only half its request and does NOT
// half-close, so handler A is parked in readRequest and has not saved
// anything. QUIT then fires (identical to the SIGTERM path) and the accept
// loop exits — exactly where the old pre-drain FlushDirty ran. Only THEN does
// A finish and half-close, causing handler A to dispatch+SaveSession after the
// loop is gone. The daemon must drain (wg.Wait) and do the final flush.
func TestServe_FlushesSessionCompletedDuringShutdownDrain(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)

	sid := "sess-drain-flush"
	fp := filepath.Join(t.TempDir(), "drain.go")
	content := strings.Repeat("package drain\n", 80)
	if err := os.WriteFile(fp, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	payload := readEventPayload("PostToolUse", sid, fp, content)

	done := make(chan error, 1)
	go func() { done <- daemon.Serve(sock, time.Minute, "test") }()
	waitForSock(t, sock)

	// Connection A: open and write only HALF the request, no half-close.
	// Handler A parks in readRequest (reads to EOF) — it has NOT dispatched or
	// saved a session yet, but it IS in-flight and counted in Serve's wg.
	a, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("dial A: %v", err)
	}
	defer a.Close()
	half := len(payload) / 2
	if _, err := io.WriteString(a, payload[:half]); err != nil {
		t.Fatalf("write A half: %v", err)
	}

	// Connection B: request shutdown via QUIT — the exact machinery the
	// SIGTERM/SIGINT path reuses. This closes the listener, so the accept loop
	// exits (where the pre-fix pre-drain FlushDirty used to run).
	q, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("dial B: %v", err)
	}
	if _, err := io.WriteString(q, "QUIT\n"); err != nil {
		t.Fatalf("write QUIT: %v", err)
	}
	if cw, ok := q.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
	}
	q.Close()

	// Let the accept loop observe the closed listener and exit before A saves.
	time.Sleep(150 * time.Millisecond)

	// Now finish connection A: send the rest and half-close so handler A hits
	// EOF, dispatches, and SaveSessions — AFTER the accept loop is gone.
	if _, err := io.WriteString(a, payload[half:]); err != nil {
		t.Fatalf("write A rest: %v", err)
	}
	if cw, ok := a.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
	}
	_, _ = io.ReadAll(a) // drain reply; confirms handler A ran to completion

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not shut down after QUIT")
	}

	// The session that completed during the drain must be on disk. On the
	// pre-fix flush-before-wait code this file is absent.
	if _, err := os.Stat(cache.StatePath(sid)); err != nil {
		t.Fatalf("session file missing after shutdown (completed-but-unflushed bug): %v", err)
	}
	st, err := cache.Load(sid)
	if err != nil {
		t.Fatalf("cache.Load(%q): %v", sid, err)
	}
	if len(st.Files) == 0 {
		t.Fatalf("session decoded but has no files; expected the saved read of %s", fp)
	}
}

// TestServe_SIGTERMFlushesSessionE2E is the subprocess end-to-end test: a real
// qdf-hook daemon binary, a real SIGTERM (docker stop / systemd / kill), and a
// real on-disk session that must survive the signal via the graceful drain +
// flush path. Without signal handling the process would die instantly and the
// session (still only in RAM / dirty set) would be lost.
func TestServe_SIGTERMFlushesSessionE2E(t *testing.T) {
	homeDir, err := os.MkdirTemp("", "qd") // short path: darwin sun_path <=104
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(homeDir) })
	// The daemon resolves both its socket and its session dir from HOME; set it
	// so the child (which inherits our env) and cache.StatePath below agree.
	t.Setenv("HOME", homeDir)
	sock := filepath.Join(homeDir, ".qdf-hook", "d.sock")

	cmd := exec.Command(testExePath, "daemon", "--serve")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Skipf("cannot spawn daemon subprocess (sandbox?): %v", err)
	}
	killed := false
	t.Cleanup(func() {
		if !killed && cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})

	// Wait for the daemon's socket to come up.
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, statErr := os.Stat(sock); statErr == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("daemon socket never appeared; stderr:\n%s", stderr.String())
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Send one PostToolUse Read request that saves a session in the daemon.
	sid := "sess-sigterm-e2e"
	fp := filepath.Join(homeDir, "e2e.go")
	content := strings.Repeat("package e2e\n", 80)
	if err := os.WriteFile(fp, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_ = roundtrip(t, sock, readEventPayload("PostToolUse", sid, fp, content))

	// SIGTERM must trigger a graceful shutdown: drain + final flush + exit 0.
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("signal SIGTERM: %v", err)
	}

	// Wait for exit with a deadline — no `timeout` command on macOS.
	waitErr := make(chan error, 1)
	go func() { waitErr <- cmd.Wait() }()
	select {
	case err := <-waitErr:
		killed = true
		if err != nil {
			t.Fatalf("daemon exited non-zero on SIGTERM: %v; stderr:\n%s", err, stderr.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("daemon did not exit within 5s of SIGTERM; stderr:\n%s", stderr.String())
	}

	// The session saved before SIGTERM must be flushed to disk under HOME.
	if _, err := os.Stat(cache.StatePath(sid)); err != nil {
		t.Fatalf("session file missing after SIGTERM (no graceful flush): %v; stderr:\n%s", err, stderr.String())
	}
	st, err := cache.Load(sid)
	if err != nil {
		t.Fatalf("cache.Load(%q): %v", sid, err)
	}
	if len(st.Files) == 0 {
		t.Fatalf("session decoded but has no files; expected the saved read of %s", fp)
	}
}
