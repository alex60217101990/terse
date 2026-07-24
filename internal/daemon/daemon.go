// Package daemon implements the qdf-hookd socket serve loop: a long-lived
// process that answers PostToolUse hook requests over a unix socket, sharing
// one in-RAM hookcore.MemStore across all connections instead of the CLI's
// per-invocation disk round trip.
package daemon

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
)

// flushInterval is how often Serve flushes dirty state to disk while running.
const flushInterval = 5 * time.Second

// pingProbe is how long ping-style dials (version check, no-op detection)
// wait before giving up on a socket that should already be live.
const pingProbe = 200 * time.Millisecond

// ensureReadyTimeout bounds how long Ensure will poll a freshly started
// daemon before giving up.
const ensureReadyTimeout = 2 * time.Second

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
	// A unix socket file lingers after its owner dies. Distinguish a live
	// daemon (dial succeeds) — which we must not clobber — from a stale file
	// (dial refused/absent), which we remove so net.Listen won't fail with
	// "address already in use".
	if c, derr := net.DialTimeout("unix", sockPath, 100*time.Millisecond); derr == nil {
		_ = c.Close()
		return fmt.Errorf("daemon already running at %s", sockPath)
	}
	_ = os.Remove(sockPath)

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if mkErr := os.MkdirAll(filepath.Dir(sockPath), 0o700); mkErr != nil {
				return mkErr
			}
			ln, err = net.Listen("unix", sockPath)
		}
		if err != nil {
			return err
		}
	}
	defer ln.Close()
	ul := ln.(*net.UnixListener)

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

	var wg sync.WaitGroup
	defer wg.Wait()

	start := time.Now()
	lastActivity := start
	lastFlush := start

	for {
		now := time.Now()
		if now.Sub(lastFlush) >= flushInterval {
			store.FlushDirty()
			lastFlush = now
		}
		if now.Sub(lastActivity) >= idle {
			store.FlushDirty()
			return nil
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
			if ne, ok := aerr.(net.Error); ok && ne.Timeout() {
				// Nothing arrived before the deadline; loop re-evaluates the
				// flush and idle checks at the top.
				continue
			}
			store.FlushDirty()
			if shutdownRequested.Load() {
				// QUIT closed the listener out from under us — a clean,
				// intentional shutdown, not an error.
				return nil
			}
			return aerr
		}
		lastActivity = time.Now()
		wg.Go(func() {
			handleConn(c, store, version, requestShutdown)
		})
	}
}

// handleConn reads a single request off c to EOF (the client half-closes its
// write side once it's done sending). Two bare control requests are handled
// before the request ever reaches the hook pipeline:
//
//   - "PING" replies "qdf-hookd <version>\n" — used by Ensure to detect
//     whether a live daemon is running and, if so, whether it's current.
//   - "QUIT" calls requestShutdown (closing the listener so the accept loop
//     in Serve unblocks and exits cleanly) and replies with nothing.
//
// Anything else is run through the full hook pipeline against store. It
// never panics the daemon: a recover guards the whole handler, and any error
// (read, dispatch, or otherwise) just results in the connection being closed
// without a reply — the client's local-pipeline fallback covers that case.
func handleConn(c net.Conn, store hookcore.StateStore, version string, requestShutdown func()) {
	defer c.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("daemon: recovered panic in handleConn: %v", r)
		}
	}()

	buf := readRequest(c)
	defer putRequest(buf)
	req := buf.Bytes()

	switch strings.TrimSpace(string(req)) {
	case "PING":
		fmt.Fprintf(c, "qdf-hookd %s\n", version)
		return
	case "QUIT":
		requestShutdown()
		return
	}

	_ = hook.Dispatch(store, bytes.NewReader(req), c)
}

// ping dials sockPath, sends a PING request, and returns the trimmed reply
// (e.g. "qdf-hookd v0.1.0"). It returns an error if the dial, write, or read
// fails or times out — indistinguishable reasons for treating the daemon as
// not (yet) live.
func ping(sockPath string, timeout time.Duration) (string, error) {
	c, err := net.DialTimeout("unix", sockPath, timeout)
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
	c, err := net.DialTimeout("unix", sockPath, pingProbe)
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
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("unix", sockPath, 50*time.Millisecond)
		if err != nil {
			return
		}
		_ = c.Close()
		time.Sleep(10 * time.Millisecond)
	}
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

	cmd := exec.Command(exePath, "daemon", "--serve")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
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
