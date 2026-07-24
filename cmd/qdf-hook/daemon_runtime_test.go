package main

import (
	"runtime"
	"runtime/debug"
	"testing"
)

// TestRestoreDaemonRuntime verifies the daemon undoes the one-shot-CLI tuning
// (GOMAXPROCS(1) + SetGCPercent(-1)) so the long-lived process garbage-collects
// and uses all Ps. It mutates process-global runtime state — keep it isolated.
func TestRestoreDaemonRuntime(t *testing.T) {
	// Simulate main.init() having disabled GC + pinned to 1 P.
	runtime.GOMAXPROCS(1)
	debug.SetGCPercent(-1)

	restoreDaemonRuntime()

	if got := runtime.GOMAXPROCS(0); got != runtime.NumCPU() {
		t.Errorf("GOMAXPROCS = %d, want NumCPU %d", got, runtime.NumCPU())
	}
	// SetGCPercent returns the PREVIOUS value; after restore it should be 100.
	if prev := debug.SetGCPercent(100); prev != 100 {
		t.Errorf("GC percent after restore = %d, want 100", prev)
	}
}
