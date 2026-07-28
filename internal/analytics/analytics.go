package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// recordMu serializes the package-level Record's per-call open/stat/write/
// close path (the CLI: one process, one call, then exit). The daemon
// instead routes through a Writer (see OpenWriter/SetWriter below), which
// holds its own mutex and its own long-lived fd — recordMu is never taken
// on that path.
var recordMu sync.Mutex

// rotateAt is the file-size threshold above which the analytics log is
// rotated (renamed aside to path+".1") before the next write lands. It's a
// var, not a const, purely so tests can lower it to something reachable
// without actually writing >10MB.
var rotateAt int64 = 10 * 1024 * 1024

// activeWriter is the daemon's installed Writer, if any. nil (the default,
// and what the CLI always sees) means Record falls back to its per-call
// open/write/close path. An atomic.Pointer keeps the common-case load
// lock-free on this every-request hot path; SetWriter is the only writer.
var activeWriter atomic.Pointer[Writer]

// SetWriter installs w as the destination for the package-level Record.
// Pass nil to restore the per-call path — what Serve does on shutdown,
// after its final flush, so no straggling handler goroutine can write
// through a Writer whose fd is about to be closed.
func SetWriter(w *Writer) {
	activeWriter.Store(w)
}

// Event records one hook invocation for analytics purposes.
type Event struct {
	SID      string `json:"sid"` // first 16 chars of session_id
	Hook     string `json:"hook"`
	Action   string `json:"action"`
	TS       int64  `json:"ts"` // unix nanoseconds
	BytesIn  int    `json:"bi"`
	BytesOut int    `json:"bo"`
	DurNS    int64  `json:"dur"` // hook duration in nanoseconds
}

// AnalyticsPath returns the path to the analytics JSONL file.
func AnalyticsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".qdf-hook", "analytics.jsonl")
}

// Record appends one event to the analytics JSONL file. If a Writer has
// been installed via SetWriter (the daemon does this for its whole
// lifetime), the event is routed through its held fd instead — see
// Writer.Record. Errors are non-fatal — hooks must not crash on analytics
// failure.
func Record(e Event) error {
	if w := activeWriter.Load(); w != nil {
		return w.Record(e)
	}

	if len(e.SID) > 16 {
		e.SID = e.SID[:16]
	}
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	recordMu.Lock()
	defer recordMu.Unlock()

	path := AnalyticsPath()

	// Open first; only MkdirAll on ENOENT (the dir already exists on every
	// call but the first). Mirrors cache/io.go's writeFileLazy — drops an
	// os.MkdirAll from this per-hook hot path.
	f, err := openAppend(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if mkErr := os.MkdirAll(filepath.Dir(path), 0o700); mkErr != nil {
			return mkErr
		}
		if f, err = openAppend(path); err != nil {
			return err
		}
	}

	// Rotate if > rotateAt, checked via the open fd rather than a second
	// path stat. Reopen the fresh file after renaming the old one aside.
	if info, e := f.Stat(); e == nil && info.Size() > rotateAt {
		_ = f.Close()
		_ = os.Rename(path, path+".1")
		if f, err = openAppend(path); err != nil {
			return err
		}
	}
	defer f.Close()
	_, err = f.Write(line)
	return err
}

func openAppend(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
}

// checkEvery and checkInterval bound how often Writer.Record pays for a
// stat-based rotation/reopen check: every checkEvery writes or
// checkInterval elapsed, whichever comes first — not on every call, unlike
// the package-level Record's per-call stat.
const checkEvery = 256

const checkInterval = 30 * time.Second

// Writer owns one long-lived O_APPEND fd for the analytics log, amortizing
// the open/stat/close cost that the package-level Record pays on every call
// across many writes. Meant for the daemon, whose process (and therefore
// its analytics writes) lives far longer than the CLI's one-shot Record.
//
// All exported methods are goroutine-safe: qdf-hookd's connection handlers
// call Record concurrently from many goroutines, so the internal mutex is
// load-bearing, not incidental.
type Writer struct {
	lastCheck time.Time
	f         *os.File
	path      string
	writes    int
	mu        sync.Mutex
}

// OpenWriter opens (creating the parent dir if needed) the analytics log
// for appending and returns a Writer that keeps the fd open across calls.
// Callers must Close it when done.
func OpenWriter() (*Writer, error) {
	path := AnalyticsPath()
	f, err := openAppend(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if mkErr := os.MkdirAll(filepath.Dir(path), 0o700); mkErr != nil {
			return nil, mkErr
		}
		if f, err = openAppend(path); err != nil {
			return nil, err
		}
	}
	return &Writer{path: path, f: f, lastCheck: time.Now()}, nil
}

// Record appends one event through the writer's held fd. Every checkEvery
// calls, or every checkInterval, whichever comes first, it also runs
// checkLocked to catch external rotation and to rotate the file itself past
// rotateAt — the same rotation semantics as the package-level Record, just
// amortized instead of paid on every call.
func (w *Writer) Record(e Event) error {
	if len(e.SID) > 16 {
		e.SID = e.SID[:16]
	}
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	w.mu.Lock()
	defer w.mu.Unlock()

	w.writes++
	if w.writes >= checkEvery || time.Since(w.lastCheck) >= checkInterval {
		if cerr := w.checkLocked(); cerr != nil {
			return cerr
		}
		w.writes = 0
		w.lastCheck = time.Now()
	}

	_, err = w.f.Write(line)
	return err
}

// checkLocked reopens the file if some other process rotated or removed it
// out from under our held fd (detected by comparing our fd's identity
// against a fresh stat of path via os.SameFile — cheap, and only ever run
// on the amortized schedule above), then rotates our own file aside if it
// has grown past rotateAt. Must be called with w.mu held.
func (w *Writer) checkLocked() error {
	fi, ferr := w.f.Stat()
	if ferr == nil {
		if pi, perr := os.Stat(w.path); perr != nil || !os.SameFile(fi, pi) {
			_ = w.f.Close()
			f, err := openAppend(w.path)
			if err != nil {
				return err
			}
			w.f = f
			fi, ferr = w.f.Stat()
		}
	}

	if ferr == nil && fi.Size() > rotateAt {
		_ = w.f.Close()
		_ = os.Rename(w.path, w.path+".1")
		f, err := openAppend(w.path)
		if err != nil {
			return err
		}
		w.f = f
	}
	return nil
}

// Close releases the writer's fd.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.f.Close()
}

// SavedTokens estimates tokens saved from an event (4 bytes per token).
func SavedTokens(e Event) int {
	saved := e.BytesIn - e.BytesOut
	if saved < 0 {
		return 0
	}
	return saved / 4
}

// FormatBytes formats byte counts for display.
func FormatBytes(n int) string {
	if n < 0 {
		return "-" + FormatBytes(-n)
	}
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	case n >= 1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
