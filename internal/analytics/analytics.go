package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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

	path := AnalyticsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	// Rotate if > 10MB.
	if info, err := os.Stat(path); err == nil && info.Size() > 10*1024*1024 {
		_ = os.Rename(path, path+".1")
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(line)
	return err
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
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	case n >= 1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
