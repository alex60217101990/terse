package daemon

import (
	"io"
	"net"
	"strings"
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
// minExchangeDeadline bounds the read+write exchange on an ALREADY-dialed
// connection. It must exceed the daemon's per-connection deadline (10s): the
// client cannot recover after a committed dial (stdin may be consumed), so if
// it gave up first — e.g. on the short dial timeout callers pass — a slow reply
// could be copied to stdout partially and then error, producing a truncated
// hookSpecificOutput. Letting the daemon (which self-bounds and recovers) time
// out first keeps the reply all-or-nothing.
const minExchangeDeadline = 15 * time.Second

func ProxyConn(c net.Conn, in io.Reader, out io.Writer, timeout time.Duration) error {
	defer c.Close()
	// timeout is a dial-scale bound; give the exchange at least
	// minExchangeDeadline so the client never abandons a reply before the daemon.
	if timeout < minExchangeDeadline {
		timeout = minExchangeDeadline
	}
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

// Stats sends the STATS control word and returns the daemon's runtime
// snapshot (see writeStats in daemon.go): one metric per line, for tuning and
// diagnostics — not part of the hook dispatch pipeline.
func Stats(sockPath string, timeout time.Duration) (string, error) {
	var b strings.Builder
	if err := DialAndProxy(sockPath, strings.NewReader("STATS\n"), &b, timeout); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Ping sends the PING control word and returns the daemon's trimmed version
// reply (e.g. "qdf-hookd v0.1.0"). Exported so callers outside this package
// (the `daemon --ping` container-healthcheck flag) can probe liveness without
// reimplementing the PING protocol; it is the same probe Ensure uses
// internally via the unexported ping helper.
func Ping(sockPath string, timeout time.Duration) (string, error) {
	return ping(sockPath, timeout)
}
