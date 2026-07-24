package daemon

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestConnDeadline_StalledClientDoesNotWedgeShutdown is the regression test for
// the final-review Important finding: a client that connects but never
// half-closes must not block the daemon forever. With the per-connection
// deadline, the stalled handler returns on its own, so a QUIT still shuts the
// daemon down promptly (Serve's defer wg.Wait() no longer waits forever).
// Without the deadline this test hangs until the -timeout fires.
func TestConnDeadline_StalledClientDoesNotWedgeShutdown(t *testing.T) {
	old := connDeadlineNS.Load()
	connDeadlineNS.Store(int64(200 * time.Millisecond))
	t.Cleanup(func() { connDeadlineNS.Store(old) })

	t.Setenv("HOME", t.TempDir())
	dir, err := os.MkdirTemp("", "qd")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "s")

	done := make(chan error, 1)
	go func() { done <- Serve(sock, time.Minute, "test") }()

	// Wait for the socket to appear.
	deadline := time.Now().Add(time.Second)
	for {
		if _, err := os.Stat(sock); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("socket never appeared")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Stalled client: connect, send nothing, never half-close.
	stall, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("dial stall: %v", err)
	}
	defer stall.Close()

	// Ask the daemon to quit via a separate connection.
	q, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatalf("dial quit: %v", err)
	}
	_, _ = q.Write([]byte("QUIT\n"))
	if cw, ok := q.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
	}
	q.Close()

	// Serve must return well within a few multiples of the (short) deadline;
	// without the deadline the stalled handler blocks wg.Wait() forever.
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not shut down: stalled connection wedged wg.Wait()")
	}
}
