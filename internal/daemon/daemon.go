// Package daemon implements the qdf-hookd socket serve loop: a long-lived
// process that answers PostToolUse hook requests over a unix socket, sharing
// one in-RAM hookcore.MemStore across all connections instead of the CLI's
// per-invocation disk round trip.
package daemon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/alex60217101990/terse/internal/analytics"
	"github.com/alex60217101990/terse/internal/bytesconv"
	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/hook"
	"github.com/alex60217101990/terse/internal/hookcore"
)

// sunPathMax is the portable unix-socket path budget: darwin's sun_path is
// 104 bytes (linux 108). Exceeding it makes bind fail with the cryptic
// "invalid argument", so we detect and say it plainly.
const sunPathMax = 104

// flushInterval is how often Serve flushes dirty state to disk while running.
const flushInterval = 5 * time.Second

// sweepInterval is how often Serve prunes the blob caches (refs/ and last/)
// down to their size cap and TTL while running.
const sweepInterval = 10 * time.Minute

// pingProbe is how long ping-style dials (version check, no-op detection)
// wait before giving up on a socket that should already be live.
const pingProbe = 200 * time.Millisecond

// ensureReadyTimeout bounds how long Ensure will poll a freshly started
// daemon before giving up.
const ensureReadyTimeout = 2 * time.Second

// connDeadlineNS bounds a single connection's whole read+dispatch+write, in
// nanoseconds. The daemon must never hang a Claude Code hook: without this, an
// unbounded buf.ReadFrom(c) would block forever if a client connected but
// never half-closed (a wedged handler, or an nc variant that accepts but
// no-ops -N), leaking the handler goroutine and — via Serve's wg.Wait()
// — blocking clean shutdown. A deadline turns every such case into a closed
// connection within a bounded time, so nc unblocks and the CLI fallback (or
// passthrough) runs. Set generously above any real in-memory dispatch cost.
//
// It's an atomic int64 rather than a const only so tests can shorten it;
// handlers read it concurrently, so a plain mutable var would be a data race.
// int64-of-nanoseconds (not atomic.Pointer[time.Duration]) keeps it a lock-free
// scalar with no indirection or allocation.
var connDeadlineNS atomic.Int64

func init() { connDeadlineNS.Store(int64(10 * time.Second)) }

// startTime is set at the top of Serve and read by writeStats to report
// daemon uptime over the STATS control word.
var startTime time.Time

// SockPath returns the well-known unix socket path for qdf-hookd:
// ~/.qdf-hook/d.sock. The parent directory is not created here — callers
// that need it to exist (Serve, via net.Listen) create it lazily, matching
// the project's writeFileLazy style (see internal/cache/io.go).
func SockPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".qdf-hook", "d.sock")
}

// SocketPathTooLong reports whether path exceeds the sun_path limit.
func SocketPathTooLong(path string) bool { return len(path) > sunPathMax }

// Serve listens on a unix socket at sockPath and answers hook requests until
// no connection has arrived for idle, then does a final flush and returns
// nil. Serve blocks the calling goroutine; callers that want it in the
// background must run it in its own goroutine.
//
// If a live daemon already owns sockPath, Serve returns an error rather than
// clobbering it (probed by dialing); a stale socket left by a dead daemon is
// removed and recreated. If the socket's parent directory doesn't exist yet,
// it's created lazily (mirroring internal/cache/io.go's writeFileLazy: try
// first, mkdir only on ENOENT, then retry).
//
// version is reported verbatim to any client that opens a connection and
// sends a bare "PING" request — used by Ensure to detect a stale daemon
// without going through the hook pipeline. A "QUIT" request triggers a clean
// shutdown: the listener is closed (unblocking Accept) and Serve flushes and
// returns nil, exactly like an idle-timeout exit.
//
// One hookcore.MemStore is shared by every connection (they are already
// race-safe against each other per Task 2); each accepted connection is
// handled in its own goroutine so slow/blocked clients never stall others.
//
// Timing is single-threaded and race-free by construction: the accept loop
// runs in the calling goroutine and the listener's own deadline is the idle
// timer. Accept returns either a connection or a timeout — never both — so a
// connection can never be accepted-but-abandoned, and there is no separate
// accept goroutine to leak. Dirty state is flushed at least every
// flushInterval and once more before Serve returns.
func Serve(sockPath string, idle time.Duration, version string) error {
	startTime = time.Now()

	// A unix socket file lingers after its owner dies. Distinguish a live
	// daemon (dial succeeds) — which we must not clobber — from a stale file
	// (dial refused/absent), which we remove so net.Listen won't fail with
	// "address already in use".
	dialer := net.Dialer{Timeout: 100 * time.Millisecond}
	if c, derr := dialer.DialContext(context.Background(), "unix", sockPath); derr == nil {
		_ = c.Close()
		return fmt.Errorf("daemon already running at %s", sockPath)
	}
	// Note: two `daemon --ensure` calls from two sessions starting at the same
	// instant can both pass the dial check and both reach os.Remove+Listen
	// here; the second unlinks the first's freshly bound socket, orphaning the
	// first daemon (it keeps running, flushing, until its 30m idle-exit). The
	// outcome is harmless — last-writer-wins on a rebuildable on-disk cache —
	// but the orphan wastes a goroutine and RAM until it times out. Not worth
	// a cross-process lock for a rare startup race with a self-healing result.
	_ = os.Remove(sockPath)

	if SocketPathTooLong(sockPath) {
		msg := fmt.Sprintf("qdf-hookd: socket path too long (%d bytes > %d): %s", len(sockPath), sunPathMax, sockPath)
		log.Print(msg)
		fmt.Fprintln(os.Stderr, msg)
	}

	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "unix", sockPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if mkErr := os.MkdirAll(filepath.Dir(sockPath), 0o700); mkErr != nil {
				return mkErr
			}
			ln, err = lc.Listen(context.Background(), "unix", sockPath)
		}
		if err != nil {
			return err
		}
	}
	defer func() { _ = ln.Close() }()
	ul, ok := ln.(*net.UnixListener)
	if !ok {
		return fmt.Errorf("daemon: listener for %q is not a *net.UnixListener", sockPath)
	}

	// Open one long-lived analytics appender fd for the whole daemon
	// lifetime and route every hook handler's analytics.Record through it
	// (see analytics.SetWriter), instead of each request paying its own
	// open+stat+write+close under analytics' package mutex. A failure here
	// is non-fatal — analytics is best-effort diagnostics — so on error we
	// log and fall back to the package-level Record's per-call path, same
	// as the CLI.
	var analyticsWriter *analytics.Writer
	if w, aerr := analytics.OpenWriter(); aerr != nil {
		log.Printf("daemon: analytics.OpenWriter: %v (falling back to per-call analytics writes)", aerr)
	} else {
		analyticsWriter = w
		analytics.SetWriter(w)
	}

	store := hookcore.NewMemStore()

	// shutdownRequested distinguishes a QUIT-triggered listener close (clean
	// exit, same as idle timeout) from any other Accept error (real
	// failure). requestShutdown is handed to each connection handler so a
	// QUIT request — arriving on its own goroutine — can signal the
	// single-threaded accept loop to stop: set the flag, then close the
	// listener to unblock a pending Accept.
	var shutdownRequested atomic.Bool
	requestShutdown := func() {
		shutdownRequested.Store(true)
		_ = ul.Close()
	}

	// Graceful shutdown on SIGTERM/SIGINT (docker stop's PID-1 signal,
	// systemd, a bare kill, Ctrl-C). The signal path reuses the QUIT machinery
	// exactly: requestShutdown closes the listener, which unblocks Accept so
	// the loop exits, drains in-flight handlers, and does the final flush —
	// instead of the process being torn down mid-request with up to
	// flushInterval of un-persisted state lost. The watcher goroutine exits
	// when Serve returns (the deferred close(done) fires), so it never
	// outlives a single Serve call, and signal.Stop unregisters the handler.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigCh)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-sigCh:
			requestShutdown()
		case <-done:
		}
	}()

	var wg sync.WaitGroup

	start := time.Now()
	lastActivity := start
	lastFlush := start
	lastSweep := start

	// serveErr carries a real Accept failure out of the loop; a clean exit
	// (idle timeout, or a QUIT/signal-triggered listener close) leaves it nil.
	// Every loop-exit path funnels through the single drain-then-flush below.
	var serveErr error

	for {
		now := time.Now()
		if now.Sub(lastFlush) >= flushInterval {
			store.FlushDirty()
			lastFlush = now
		}
		if now.Sub(lastSweep) >= sweepInterval {
			sweepCache(now.Unix())
			// Prune the in-RAM usage maps on the same tick, using the same
			// TTL cache.SweepBlobs/PruneDir applies to the on-disk sidecars —
			// otherwise FlushDirty's next tick would rewrite the sidecars
			// with entries PruneDir just dropped (see MemStore.PruneUsage).
			store.PruneUsage(cache.CacheTTL())
			lastSweep = now
		}
		if now.Sub(lastActivity) >= idle {
			break
		}

		// Wake no later than whichever comes first: the next flush or the
		// idle deadline. Accept blocks until a connection arrives or that
		// deadline elapses (a Timeout error), so the deadline doubles as both
		// the flush ticker and the idle timer — no second goroutine, no racy
		// select, and an accepted connection is never abandoned.
		wake := lastFlush.Add(flushInterval)
		if d := lastActivity.Add(idle); d.Before(wake) {
			wake = d
		}
		_ = ul.SetDeadline(wake)

		c, aerr := ul.Accept()
		if aerr != nil {
			var ne net.Error
			if errors.As(aerr, &ne) && ne.Timeout() {
				// Nothing arrived before the deadline; loop re-evaluates the
				// flush and idle checks at the top.
				continue
			}
			// A non-timeout Accept error is either our own QUIT/signal close
			// (shutdownRequested — clean, intentional) or a genuine listener
			// failure. Either way, fall out to the drain-then-flush; only a
			// genuine failure is reported.
			if !shutdownRequested.Load() {
				serveErr = aerr
			}
			break
		}
		lastActivity = time.Now()
		wg.Go(func() {
			handleConn(c, store, version, requestShutdown)
		})
	}

	// Drain first, flush second. wg.Wait blocks until every in-flight handler
	// has returned — including any that committed a SaveSession while the
	// accept loop was tearing down — and only THEN does the final FlushDirty
	// run, so a completed-but-unflushed session can't be left on the floor.
	// (The old code flushed at each loop-exit site and only afterwards ran a
	// deferred wg.Wait, so a handler finishing during the drain was never
	// flushed.) FlushDirty also persists the usage indices, so ref/last
	// analytics survive shutdown too.
	wg.Wait()
	store.FlushDirty()

	// Unhook the writer before closing it: SetWriter(nil) is what makes any
	// analytics.Record call still in flight for a small window fall back to
	// the per-call path rather than write through an fd we're about to
	// close. It must come after wg.Wait so no handler goroutine is running
	// concurrently by this point anyway — belt and suspenders.
	if analyticsWriter != nil {
		analytics.SetWriter(nil)
		_ = analyticsWriter.Close()
	}

	return serveErr
}

// sweepCache prunes the refs/ and last/ blob stores down to their combined
// size cap and TTL, splitting the cap evenly across the two dirs (mirroring
// cache.RunGC's own split). Called periodically from Serve's loop, and
// directly by tests; it never blocks on anything but local disk I/O.
func sweepCache(nowSec int64) {
	cache.SweepBlobs(nowSec)
}

// handleConn reads a single request off c to EOF (the client half-closes its
// write side once it's done sending). Three bare control requests are handled
// before the request ever reaches the hook pipeline:
//
//   - "PING" replies "qdf-hookd <version>\n" — used by Ensure to detect
//     whether a live daemon is running and, if so, whether it's current.
//   - "QUIT" calls requestShutdown (closing the listener so the accept loop
//     in Serve unblocks and exits cleanly) and replies with nothing.
//   - "STATS" replies with a one-metric-per-line runtime snapshot (see
//     writeStats) — for daemon tuning, not correctness.
//
// Anything else is run through the full hook pipeline against store. It
// never panics the daemon: a recover guards the whole handler, and any error
// (read, dispatch, or otherwise) closes the connection with no reply. An empty
// reply is safe by the passthrough equivalence — Claude Code treats empty hook
// stdout as "use the original output". Note the socket-first client does NOT
// re-dispatch locally after a committed dial (stdin may be consumed), so that
// equivalence, not a client fallback, is the real safety net; the dispatch
// error is logged below so a systemic regression is observable.
func handleConn(c net.Conn, store hookcore.StateStore, version string, requestShutdown func()) {
	defer c.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("daemon: recovered panic in handleConn: %v", r)
		}
	}()

	// Bound the whole exchange so a client that connects but never
	// half-closes can't wedge the handler (and, via wg.Wait, shutdown).
	_ = c.SetDeadline(time.Now().Add(time.Duration(connDeadlineNS.Load())))

	buf := readRequest(c)
	defer putRequest(buf)
	req := buf.Bytes()

	// B2S over the trimmed bytes: avoid copying the whole (up to 1 MiB)
	// request just to compare against two 4-byte control words. The view is
	// switch-local and never outlives req (released by defer putRequest).
	switch bytesconv.B2S(bytes.TrimSpace(req)) {
	case "PING":
		fmt.Fprintf(c, "qdf-hookd %s\n", version)
		return
	case "QUIT":
		requestShutdown()
		return
	case "STATS":
		writeStats(c)
		return
	}

	// DispatchBytes decodes the already-buffered request directly, skipping the
	// json.Decoder buffering an io.Reader path would add. An error closes the
	// connection with no reply (safe passthrough); log it to the daemon's
	// stderr (→ ~/.qdf-hook/daemon.log) so a systemic decode/dispatch regression
	// surfaces instead of silently degrading every request to passthrough.
	if err := hook.DispatchBytes(store, req, c); err != nil {
		log.Printf("daemon: dispatch: %v", err)
	}
}

// writeStats emits a compact one-metric-per-line runtime snapshot for tuning
// (peer of PING/QUIT). Cheap: reads MemStats once, no profiling machinery.
func writeStats(c net.Conn) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Fprintf(c, "heap_alloc %d\nheap_inuse %d\nnum_gc %d\npause_total_ns %d\nnum_goroutine %d\nuptime %s\n",
		m.HeapAlloc, m.HeapInuse, m.NumGC, m.PauseTotalNs, runtime.NumGoroutine(), time.Since(startTime))
}

// ping dials sockPath, sends a PING request, and returns the trimmed reply
// (e.g. "qdf-hookd v0.1.0"). It returns an error if the dial, write, or read
// fails or times out — indistinguishable reasons for treating the daemon as
// not (yet) live.
func ping(sockPath string, timeout time.Duration) (string, error) {
	dialer := net.Dialer{Timeout: timeout}
	c, err := dialer.DialContext(context.Background(), "unix", sockPath)
	if err != nil {
		return "", err
	}
	defer c.Close()

	if _, err := io.WriteString(c, "PING\n"); err != nil {
		return "", err
	}
	if cw, ok := c.(interface{ CloseWrite() error }); ok {
		if err := cw.CloseWrite(); err != nil {
			return "", err
		}
	}
	_ = c.SetReadDeadline(time.Now().Add(timeout))

	out, err := io.ReadAll(c)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// sendQuit dials sockPath and sends a QUIT request, telling a live daemon to
// shut down cleanly. Errors are not fatal to the caller: if the daemon is
// already gone there's nothing to quit, and Ensure's subsequent wait/replace
// logic handles that either way.
func sendQuit(sockPath string) {
	dialer := net.Dialer{Timeout: pingProbe}
	c, err := dialer.DialContext(context.Background(), "unix", sockPath)
	if err != nil {
		return
	}
	defer c.Close()

	if _, err := io.WriteString(c, "QUIT\n"); err != nil {
		return
	}
	if cw, ok := c.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
	}
	_, _ = io.ReadAll(c) // drain; the daemon replies with nothing and closes.
}

// waitGone polls sockPath until dialing it fails (the previous daemon has
// released the socket) or timeout elapses.
func waitGone(sockPath string, timeout time.Duration) {
	dialer := net.Dialer{Timeout: 50 * time.Millisecond}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c, err := dialer.DialContext(context.Background(), "unix", sockPath)
		if err != nil {
			return
		}
		_ = c.Close()
		time.Sleep(10 * time.Millisecond)
	}
}

// openDaemonLog opens the daemon's stderr log file (daemon.log, beside the
// socket) for appending, creating the parent directory if needed. Used so a
// detached daemon's panic output has somewhere to go instead of /dev/null.
func openDaemonLog(sockPath string) (*os.File, error) {
	dir := filepath.Dir(sockPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(filepath.Join(dir, "daemon.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
}

// Ensure makes sure a live, current qdf-hookd is serving sockPath, starting
// or replacing it as needed:
//
//   - If a daemon is already live and answers PING with exactly
//     "qdf-hookd <version>", Ensure is a no-op: it returns nil without
//     starting anything.
//   - If a daemon is live but reports a different version (stale, e.g. after
//     a qdf-hook upgrade), Ensure sends it QUIT and waits briefly for it to
//     release the socket — Serve itself refuses to bind over a still-live
//     daemon, so the old one must step aside before a new one can start.
//   - If no daemon answers at all, Ensure starts one directly.
//
// In both replacement cases, Ensure execs `exePath daemon --serve` detached
// (new session, not waited on) and polls PING for up to ~2s so it only
// returns once the new daemon is actually serving.
func Ensure(sockPath, exePath, version string) error {
	want := "qdf-hookd " + version

	if reply, err := ping(sockPath, pingProbe); err == nil {
		if reply == want {
			return nil // already live and current
		}
		// Live but stale: ask it to step aside before we start a new one.
		sendQuit(sockPath)
		waitGone(sockPath, time.Second)
	}

	cmd := exec.CommandContext(context.Background(), exePath, "daemon", "--serve")
	cmd.SysProcAttr = detachSysProcAttr()
	// stdin/stdout are left nil (exec connects them to /dev/null) so the
	// daemon never holds the hook's pipes open. stderr goes to a log file
	// rather than /dev/null so a daemon panic (handleConn's recover log) is
	// diagnosable; on any failure to open it we fall back to /dev/null.
	if logf, lerr := openDaemonLog(sockPath); lerr == nil {
		cmd.Stderr = logf
		defer logf.Close()
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("daemon: start %s: %w", exePath, err)
	}
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("daemon: release %s: %w", exePath, err)
	}

	deadline := time.Now().Add(ensureReadyTimeout)
	for time.Now().Before(deadline) {
		if reply, err := ping(sockPath, 100*time.Millisecond); err == nil && reply == want {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("daemon: %s did not start serving %s within %s", exePath, sockPath, ensureReadyTimeout)
}
