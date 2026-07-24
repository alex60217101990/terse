package daemon

import (
	"io"
	"net"
	"time"
)

// DialAndProxy connects to the daemon's unix socket, streams in to it,
// half-closes the write side (the daemon reads to EOF — see handleConn), and
// copies the reply to out. A non-nil error means the daemon was unreachable;
// the caller must then dispatch inline so the hook still produces correct
// output (never-worse). This is the pure-Go fast path used by `qdf-hook post`
// and `pretooluse`, and the sole client on platforms without qdf-hookc.
func DialAndProxy(sockPath string, in io.Reader, out io.Writer, timeout time.Duration) error {
	c, err := net.DialTimeout("unix", sockPath, timeout)
	if err != nil {
		return err
	}
	return ProxyConn(c, in, out, timeout)
}

// ProxyConn streams in to an already-dialed connection c, half-closes the
// write side so the daemon's readRequest hits EOF and dispatches, and copies
// the reply to out. c is closed before ProxyConn returns. Split out of
// DialAndProxy so a caller that has already committed to a dial (e.g. the CLI's
// socket-first path, where falling back after a successful dial would mean
// re-reading a consumed os.Stdin) can proxy the same connection without
// re-dialing.
func ProxyConn(c net.Conn, in io.Reader, out io.Writer, timeout time.Duration) error {
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(timeout))

	if _, err := io.Copy(c, in); err != nil {
		return err
	}
	// Half-close so the daemon's readRequest hits EOF and dispatches.
	if uc, ok := c.(*net.UnixConn); ok {
		_ = uc.CloseWrite()
	}
	_, err := io.Copy(out, c)
	return err
}
