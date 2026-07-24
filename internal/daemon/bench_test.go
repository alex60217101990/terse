package daemon_test

import (
	"bytes"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/daemon"
)

// benchPayload is the single fixed payload both benchmarks send, so the
// measured delta between them is pure spawn+IPC overhead, not differing
// amounts of pipeline work.
var benchPayload = bashPayload(strings.Repeat("bench log line here\n", 40))

// BenchmarkDaemonRoundtrip isolates the hot daemon-client path: a warm
// qdf-hookd (started once, outside the timed loop) already listening on its
// socket, sharing one in-RAM MemStore across every request. Each iteration
// pays only dial + write + half-close + read — no process spawn, no disk
// I/O for the store. This is the roundtrip cost the daemon design is meant
// to deliver once qdf-hookd is warm.
func BenchmarkDaemonRoundtrip(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	dir := b.TempDir()
	sock := dir + "/s"

	go func() { _ = daemon.Serve(sock, time.Minute, "bench") }()
	waitForSockB(b, sock)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		c, err := net.Dial("unix", sock)
		if err != nil {
			b.Fatalf("dial: %v", err)
		}
		if _, err := io.WriteString(c, benchPayload); err != nil {
			b.Fatalf("write: %v", err)
		}
		if cw, ok := c.(interface{ CloseWrite() error }); ok {
			if err := cw.CloseWrite(); err != nil {
				b.Fatalf("CloseWrite: %v", err)
			}
		}
		if _, err := io.Copy(io.Discard, c); err != nil {
			b.Fatalf("read: %v", err)
		}
		c.Close()
	}
}

// BenchmarkCLIRoundtrip isolates the cost the daemon exists to eliminate:
// spawning a fresh qdf-hook process per hook invocation, exactly what
// Claude Code's PostToolUse hook does today without qdf-hookd. The binary is
// built once (outside the timed loop, mirroring lifecycle_test.go's
// TestMain); each iteration execs it fresh with the same benchPayload on
// stdin and reads its stdout. The delta between this and
// BenchmarkDaemonRoundtrip is process-spawn + disk-backed diskStore
// overhead, not pipeline work — both benchmarks run the identical payload
// through the identical hook.Dispatch logic.
func BenchmarkCLIRoundtrip(b *testing.B) {
	if testExePath == "" {
		b.Fatal("testExePath not built (TestMain in lifecycle_test.go should have built it)")
	}
	home := b.TempDir()
	b.Setenv("HOME", home)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		cmd := exec.Command(testExePath, "post")
		cmd.Env = append(cmd.Env, "HOME="+home)
		cmd.Stdin = strings.NewReader(benchPayload)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			b.Fatalf("exec %s post: %v", testExePath, err)
		}
	}
}

// waitForSockB is waitForSock's *testing.B counterpart (waitForSock in
// daemon_test.go takes *testing.T; benchmarks need their own since
// *testing.B and *testing.T don't share a common Fatalf-with-Helper type).
func waitForSockB(b *testing.B, sockPath string) {
	b.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sockPath); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	b.Fatalf("socket %s never appeared", sockPath)
}
