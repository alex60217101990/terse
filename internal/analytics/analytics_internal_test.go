package analytics

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestWriter_RecordsReadableByLoadEvents proves the Writer's Record writes
// land in the same file LoadEvents reads back — the whole point of routing
// the daemon's analytics through a held fd instead of per-call Record.
func TestWriter_RecordsReadableByLoadEvents(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	w, err := OpenWriter()
	if err != nil {
		t.Fatalf("OpenWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	for i := range 3 {
		e := Event{Hook: "read", Action: "full", BytesIn: 100 + i, BytesOut: 10}
		if err := w.Record(e); err != nil {
			t.Fatalf("Record %d: %v", i, err)
		}
	}

	events, err := LoadEvents(0)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d: %+v", len(events), events)
	}
}

// TestWriter_RotatesAtThreshold lowers rotateAt to something a handful of
// small writes will cross, then writes past checkEvery so the amortized
// check actually fires, and confirms the old file was renamed aside and new
// writes land in a fresh path.
func TestWriter_RotatesAtThreshold(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	origRotateAt := rotateAt
	rotateAt = 200 // small: a few dozen JSON lines already exceed this
	t.Cleanup(func() { rotateAt = origRotateAt })

	w, err := OpenWriter()
	if err != nil {
		t.Fatalf("OpenWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	path := AnalyticsPath()

	// checkEvery bounds how often the rotation check runs; write past it so
	// the check is guaranteed to fire at least once, well after the file
	// has grown past the lowered rotateAt.
	for i := range checkEvery + 10 {
		e := Event{Hook: "rot", Action: "x", BytesIn: i}
		if err := w.Record(e); err != nil {
			t.Fatalf("Record %d: %v", i, err)
		}
	}

	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("expected rotated file %s.1 to exist: %v", path, err)
	}

	// The live path must exist too (post-rotation writes reopened it) and
	// be smaller than the pre-rotation accumulation, since it only holds
	// whatever landed after the rotation point.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected current path %s to exist: %v", path, err)
	}
	if info.Size() <= 0 {
		t.Fatalf("expected some writes to have landed post-rotation, got size %d", info.Size())
	}
}

// TestWriter_ReopensAfterExternalRename simulates another process rotating
// the analytics file out from under our held fd (renaming it aside without
// going through this Writer). The next periodic check must notice the
// mismatch via os.SameFile and reopen path, so subsequent writes land in a
// fresh file rather than the orphaned renamed one.
func TestWriter_ReopensAfterExternalRename(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	w, err := OpenWriter()
	if err != nil {
		t.Fatalf("OpenWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	path := AnalyticsPath()
	if err := w.Record(Event{Hook: "a", Action: "before-rename"}); err != nil {
		t.Fatalf("Record a: %v", err)
	}

	if err := os.Rename(path, path+".moved"); err != nil {
		t.Fatalf("external rename: %v", err)
	}

	// Force the next Record's periodic check to run regardless of the
	// write-count cadence, mirroring what "or 30s, whichever first" means
	// in a test that can't wait 30 real seconds.
	w.mu.Lock()
	w.lastCheck = time.Now().Add(-2 * checkInterval)
	w.mu.Unlock()

	if err := w.Record(Event{Hook: "b", Action: "after-rename"}); err != nil {
		t.Fatalf("Record b: %v", err)
	}

	reopened, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected path to be reopened and recreated: %v", err)
	}
	if !strings.Contains(string(reopened), `"hook":"b"`) {
		t.Fatalf("expected event b in reopened path, got %q", reopened)
	}

	moved, err := os.ReadFile(path + ".moved")
	if err != nil {
		t.Fatalf("expected the externally renamed file to still hold event a: %v", err)
	}
	if !strings.Contains(string(moved), `"hook":"a"`) {
		t.Fatalf("expected event a in the renamed-aside file, got %q", moved)
	}
}

// TestWriter_ConcurrentRecordsRaceFree exercises the Writer the way the
// daemon does: many goroutines calling Record concurrently. Run with -race;
// also checks no event is lost under concurrent writes.
func TestWriter_ConcurrentRecordsRaceFree(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	w, err := OpenWriter()
	if err != nil {
		t.Fatalf("OpenWriter: %v", err)
	}
	defer func() { _ = w.Close() }()

	const goroutines = 16
	const perGoroutine = 50

	var wg sync.WaitGroup
	for g := range goroutines {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := range perGoroutine {
				_ = w.Record(Event{Hook: "concurrent", Action: "x", BytesIn: g*1000 + i})
			}
		}(g)
	}
	wg.Wait()

	events, err := LoadEvents(0)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(events) != goroutines*perGoroutine {
		t.Fatalf("expected %d events, got %d", goroutines*perGoroutine, len(events))
	}
}

// TestRecord_UsesInstalledWriter proves the package-level Record routes
// through an installed Writer (what Serve does for its whole lifetime)
// instead of the per-call open/stat/write/close path.
func TestRecord_UsesInstalledWriter(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	w, err := OpenWriter()
	if err != nil {
		t.Fatalf("OpenWriter: %v", err)
	}
	SetWriter(w)
	defer func() {
		SetWriter(nil)
		_ = w.Close()
	}()

	if err := Record(Event{Hook: "via-writer", Action: "x"}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	events, err := LoadEvents(0)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(events) != 1 || events[0].Hook != "via-writer" {
		t.Fatalf("expected 1 via-writer event, got %+v", events)
	}
}

// TestRecord_NilWriterFallsBackToPerCallPath is the CLI regression case:
// with no Writer installed (activeWriter nil, the package's zero value and
// what the CLI always sees since it never calls SetWriter), Record must
// behave exactly as before this change.
func TestRecord_NilWriterFallsBackToPerCallPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	SetWriter(nil) // defensive: guard against test-order leakage of a Writer

	if err := Record(Event{Hook: "cli", Action: "full", BytesIn: 42}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	events, err := LoadEvents(0)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(events) != 1 || events[0].Hook != "cli" {
		t.Fatalf("expected 1 cli event, got %+v", events)
	}
}
