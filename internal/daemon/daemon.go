// Package daemon implements the qdf-hookd socket serve loop: a long-lived
// process that answers PostToolUse hook requests over a unix socket, sharing
// one in-RAM hookcore.MemStore across all connections instead of the CLI's
// per-invocation disk round trip.
package daemon

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
)

// flushInterval is how often Serve flushes dirty state to disk while running.
const flushInterval = 5 * time.Second

// Serve listens on a unix socket at sockPath and answers hook requests until
// no connection has arrived for idle, then does a final flush and returns
// nil. Serve blocks the calling goroutine; callers that want it in the
// background must run it in its own goroutine.
//
// If a live daemon already owns sockPath, Serve returns an error rather than
// clobbering it (probed by dialing); a stale socket left by a dead daemon is
// removed and recreated.
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
func Serve(sockPath string, idle time.Duration) error {
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
		return err
	}
	defer ln.Close()
	ul := ln.(*net.UnixListener)

	store := hookcore.NewMemStore()

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
			return aerr
		}
		lastActivity = time.Now()
		wg.Go(func() {
			handleConn(c, store)
		})
	}
}

// handleConn reads a single request off c to EOF (the client half-closes its
// write side once it's done sending), runs it through the full hook pipeline
// against store, and writes the JSON reply back to c. It never panics the
// daemon: a recover guards the whole handler, and any error (read, dispatch,
// or otherwise) just results in the connection being closed without a reply —
// the client's local-pipeline fallback covers that case.
func handleConn(c net.Conn, store hookcore.StateStore) {
	defer c.Close()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("daemon: recovered panic in handleConn: %v", r)
		}
	}()

	req, err := io.ReadAll(c)
	if err != nil {
		return
	}

	_ = hook.Dispatch(store, bytes.NewReader(req), c)
}
