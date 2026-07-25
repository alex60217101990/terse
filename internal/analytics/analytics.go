package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// recordMu serializes Record. The one-shot CLI only ever called Record once
// per process, but qdf-hookd calls it from many concurrent connection
// handlers, so the stat+rename rotation check would otherwise race (two
// goroutines both seeing >10MB and both renaming). The O_APPEND line writes
// are individually atomic; this mutex only needs to make rotation and the
// write mutually exclusive.
var recordMu sync.Mutex

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

// Record appends one event to the analytics JSONL file.
// Errors are non-fatal — hooks must not crash on analytics failure.
func Record(e Event) error {
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

	// Rotate if > 10MB, checked via the open fd rather than a second path
	// stat. Reopen the fresh file after renaming the old one aside.
	if info, e := f.Stat(); e == nil && info.Size() > 10*1024*1024 {
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
