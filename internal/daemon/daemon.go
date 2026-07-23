// Package daemon implements the qdf-hookd socket serve loop: a long-lived
// process that answers PostToolUse hook requests over a unix socket, sharing
// one in-RAM hookcore.MemStore across all connections instead of the CLI's
// per-invocation disk round trip.
package daemon

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/alex60217101990/qdf-hook/internal/hookcore"
)

// flushInterval is how often Serve flushes dirty state to disk while running.
const flushInterval = 5 * time.Second

// Serve listens on a unix socket at sockPath and answers hook requests until
// no connection has arrived for idle, then does a final flush and returns
// nil. Any stale socket file at sockPath (left behind by a killed prior
// daemon) is removed before listening. Serve blocks the calling goroutine;
// callers that want it in the background must run it in its own goroutine.
//
// One hookcore.MemStore is shared by every connection (they are already
// race-safe against each other per Task 2); each accepted connection is
// handled in its own goroutine so slow/blocked clients never stall others.
// Dirty state is flushed to disk on a ~5s ticker and once more before Serve
// returns.
func Serve(sockPath string, idle time.Duration) error {
	// Remove a stale socket file from a prior, no-longer-running daemon.
	// net.Listen fails with "address already in use" if we don't.
	if _, err := os.Stat(sockPath); err == nil {
		_ = os.Remove(sockPath)
	}

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return err
	}
	defer ln.Close()

	store := hookcore.NewMemStore()

	var lastActivity atomic.Int64
	lastActivity.Store(time.Now().UnixNano())

	// acceptCh delivers newly-accepted connections from the accept loop
	// (running in its own goroutine) to the main select loop below, which
	// also owns the flush ticker and idle check — keeping all timing
	// decisions single-threaded and race-free.
	acceptCh := make(chan net.Conn)
	acceptErrCh := make(chan error, 1)

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				acceptErrCh <- err
				return
			}
			acceptCh <- c
		}
	}()

	var wg sync.WaitGroup
	defer wg.Wait()

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	// idleCheck fires often enough to notice an idle timeout promptly
	// without busy-looping; the actual idle decision compares against
	// lastActivity regardless of how often this fires. Clamped to
	// [10ms, 1s] so both very short (test) and very long (production)
	// idle durations get a sane check cadence.
	idleCheckInterval := max(min(idle/10, time.Second), 10*time.Millisecond)
	idleTicker := time.NewTicker(idleCheckInterval)
	defer idleTicker.Stop()

	for {
		select {
		case c := <-acceptCh:
			lastActivity.Store(time.Now().UnixNano())
			wg.Add(1)
			go func() {
				defer wg.Done()
				handleConn(c, store)
			}()

		case err := <-acceptErrCh:
			// Listener died (e.g. socket removed out from under us); flush
			// and report the error.
			store.FlushDirty()
			return err

		case <-ticker.C:
			store.FlushDirty()

		case <-idleTicker.C:
			last := time.Unix(0, lastActivity.Load())
			if time.Since(last) >= idle {
				store.FlushDirty()
				return nil
			}
		}
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
