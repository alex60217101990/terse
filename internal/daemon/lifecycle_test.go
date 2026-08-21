package daemon_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/daemon"
)

// appVersion mirrors cmd/qdf-hook's default appVersion (main.go), which is
// "dev" for an un-stamped build. TestMain builds the test binary with a plain
// `go build` (no -ldflags -X), so the binary reports this default over PING;
// the two must stay in sync for TestEnsure_VersionMismatchReplacesDaemon to
// mean anything.
const appVersion = "dev"

// testExePath is a real, freshly built qdf-hook binary, built once in
// TestMain and reused by every test in this file that needs to exercise
// Ensure's real (detached exec) daemon-start path.
var testExePath string

func TestMain(m *testing.M) {
	os.Exit(runLifecycleTests(m))
}

func runLifecycleTests(m *testing.M) int {
	dir, err := os.MkdirTemp("", "qdfbin")
	if err != nil {
		fmt.Fprintln(os.Stderr, "lifecycle_test: MkdirTemp:", err)
		return 1
	}
	defer os.RemoveAll(dir)

	exe := filepath.Join(dir, "qdf-hook")
	build := exec.Command("go", "build", "-o", exe, "github.com/alex60217101990/terse/cmd/qdf-hook")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "lifecycle_test: building qdf-hook: %v\n%s\n", err, out)
		return 1
	}
	testExePath = exe

	return m.Run()
}

// quitDaemon best-effort tells whatever is listening on sockPath to shut
// down, so a real (exec'd) daemon started by a test doesn't linger past it.
func quitDaemon(sockPath string) {
	c, err := net.DialTimeout("unix", sockPath, 200*time.Millisecond)
	if err != nil {
		return
	}
	defer c.Close()
	_, _ = c.Write([]byte("QUIT\n"))
	if cw, ok := c.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
	}
}

// homeSock points $HOME at a fresh, short-path temp dir (for the darwin
// sun_path 104-byte limit — see tempSock's doc comment) and returns
// daemon.SockPath() resolved under it. Ensure's exec'd daemon reads $HOME
// itself (via daemon.SockPath(), the "daemon --serve" subcommand takes no
// socket-path flag), so any test that starts a *real* daemon process must
// use this — not an arbitrary sockPath — for the two to agree on where to
// listen.
func homeSock(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "qd")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	t.Setenv("HOME", dir)
	return daemon.SockPath()
}

// sockInode returns the inode of the file at path, used as a sentinel for
// "is this the same underlying socket file, or was it recreated". Serve
// always os.Remove + re-creates the file when it (re)binds, so a changed
// inode reliably means a new Listen happened — i.e. a new daemon started.
func sockInode(t *testing.T, path string) uint64 {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("sockInode: unsupported platform (no *syscall.Stat_t)")
	}
	return st.Ino
}

func TestEnsure_StartsDaemon_NoneRunning(t *testing.T) {
	sock := homeSock(t)
	t.Cleanup(func() { quitDaemon(sock) })

	if _, err := os.Stat(sock); err == nil {
		t.Fatalf("socket %s already exists before Ensure", sock)
	}

	if err := daemon.Ensure(sock, testExePath, appVersion); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	reply := roundtrip(t, sock, "PING\n")
	want := "qdf-hookd " + appVersion
	if strings.TrimSpace(reply) != want {
		t.Fatalf("PING reply = %q, want %q", reply, want)
	}
}

func TestEnsure_SecondCallIsNoOp(t *testing.T) {
	sock := homeSock(t)
	t.Cleanup(func() { quitDaemon(sock) })

	if err := daemon.Ensure(sock, testExePath, appVersion); err != nil {
		t.Fatalf("first Ensure: %v", err)
	}
	ino1 := sockInode(t, sock)

	if err := daemon.Ensure(sock, testExePath, appVersion); err != nil {
		t.Fatalf("second Ensure: %v", err)
	}
	ino2 := sockInode(t, sock)

	if ino1 != ino2 {
		t.Fatalf("second Ensure replaced the socket (inode %d -> %d); expected a no-op", ino1, ino2)
	}

	// Still answers correctly, proving the same daemon is still serving.
	reply := roundtrip(t, sock, "PING\n")
	want := "qdf-hookd " + appVersion
	if strings.TrimSpace(reply) != want {
		t.Fatalf("PING reply after second Ensure = %q, want %q", reply, want)
	}
}

func TestEnsure_VersionMismatchReplacesDaemon(t *testing.T) {
	sock := homeSock(t)
	t.Cleanup(func() { quitDaemon(sock) })

	// Start an in-process "stale" daemon reporting an old version — Ensure
	// must detect the mismatch, QUIT it, and replace it with a fresh one at
	// appVersion (the real qdf-hook binary's compiled-in version).
	oldDone := make(chan error, 1)
	go func() { oldDone <- daemon.Serve(sock, time.Minute, "old-version") }()
	waitForSock(t, sock)

	if reply, err := ensurePing(sock); err != nil || reply != "qdf-hookd old-version" {
		t.Fatalf("precondition: stale daemon PING = %q, err %v", reply, err)
	}

	if err := daemon.Ensure(sock, testExePath, appVersion); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	select {
	case err := <-oldDone:
		if err != nil {
			t.Fatalf("stale Serve returned error instead of clean QUIT exit: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stale daemon did not exit after Ensure's version-mismatch QUIT")
	}

	reply := roundtrip(t, sock, "PING\n")
	want := "qdf-hookd " + appVersion
	if strings.TrimSpace(reply) != want {
		t.Fatalf("PING after replace = %q, want %q", reply, want)
	}
}

func TestPING_ReturnsVersion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)
	go func() { _ = daemon.Serve(sock, time.Minute, "v-ping-test") }()
	waitForSock(t, sock)

	reply := roundtrip(t, sock, "PING\n")
	if strings.TrimSpace(reply) != "qdf-hookd v-ping-test" {
		t.Fatalf("PING reply = %q", reply)
	}

	// Bare "PING" (no trailing newline) must work too.
	reply2 := roundtrip(t, sock, "PING")
	if strings.TrimSpace(reply2) != "qdf-hookd v-ping-test" {
		t.Fatalf("bare PING reply = %q", reply2)
	}
}

func TestQuit_StopsServeCleanly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sock := tempSock(t)

	done := make(chan error, 1)
	go func() { done <- daemon.Serve(sock, time.Minute, "v1") }()
	waitForSock(t, sock)

	// QUIT gets no reply; the daemon just closes the listener and exits.
	_ = roundtrip(t, sock, "QUIT\n")

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve returned error after QUIT: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not exit after QUIT")
	}

	if _, err := net.Dial("unix", sock); err == nil {
		t.Fatal("expected dial to fail after QUIT-triggered exit")
	}
}

// ensurePing is a small variant of roundtrip that returns an error instead of
// failing the test, for a precondition check that's expected to sometimes
// need diagnosing rather than an immediate t.Fatal.
func ensurePing(sockPath string) (string, error) {
	c, err := net.DialTimeout("unix", sockPath, time.Second)
	if err != nil {
		return "", err
	}
	defer c.Close()
	if _, err := c.Write([]byte("PING\n")); err != nil {
		return "", err
	}
	if cw, ok := c.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
	}
	buf := make([]byte, 256)
	n, _ := c.Read(buf)
	return strings.TrimSpace(string(buf[:n])), nil
}
