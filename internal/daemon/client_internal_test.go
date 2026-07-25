package daemon

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestProxyConn_ExchangeOutlastsDialTimeout: callers pass a short dial-scale
// timeout, but the read+write exchange must not be bounded by it — otherwise a
// reply that arrives after that timeout is copied to stdout partially and then
// errors (a truncated hookSpecificOutput). The server here replies only AFTER
// the passed 200ms elapses; ProxyConn must still deliver the full reply.
func TestProxyConn_ExchangeOutlastsDialTimeout(t *testing.T) {
	// Short path: macOS caps unix socket paths at ~104 bytes, and t.TempDir()
	// under $TMPDIR is too long.
	sock := filepath.Join("/tmp", fmt.Sprintf("qdfpx%d.sock", os.Getpid()))
	_ = os.Remove(sock)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	defer os.Remove(sock)

	const reply = "FULL-REPLY-PAYLOAD\n"
	var wg sync.WaitGroup
	wg.Go(func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_, _ = io.ReadAll(c)               // request, to EOF (client half-closes)
		time.Sleep(400 * time.Millisecond) // reply lands after the 200ms passed timeout
		_, _ = c.Write([]byte(reply))
	})

	c, err := net.DialTimeout("unix", sock, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := ProxyConn(c, strings.NewReader("REQ\n"), &out, 200*time.Millisecond); err != nil {
		t.Fatalf("ProxyConn: %v", err)
	}
	if out.String() != reply {
		t.Fatalf("late reply lost to dial-scale deadline: got %q want %q", out.String(), reply)
	}
	wg.Wait()
}
