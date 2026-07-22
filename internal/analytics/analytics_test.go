package analytics_test

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
)

func TestRecord_WritesJSONL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	e := analytics.Event{
		TS: 1753182000000000000, SID: "abc123",
		Hook: "read", Action: "unchanged",
		BytesIn: 8192, BytesOut: 95, DurNS: 312000,
	}
	if err := analytics.Record(e); err != nil {
		t.Fatalf("Record: %v", err)
	}

	data, err := os.ReadFile(analytics.AnalyticsPath())
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	var got analytics.Event
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Hook != "read" || got.Action != "unchanged" || got.BytesIn != 8192 {
		t.Errorf("unexpected event: %+v", got)
	}
}

func TestRecord_MultipleEvents(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for range 5 {
		_ = analytics.Record(analytics.Event{Hook: "bash", Action: "summary", BytesIn: 1000, BytesOut: 50})
	}
	f, _ := os.Open(analytics.AnalyticsPath())
	defer f.Close()
	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	if count != 5 {
		t.Errorf("expected 5 lines, got %d", count)
	}
}

func TestRecord_SIDTruncated(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	e := analytics.Event{SID: "abcdefghijklmnopqrstuvwxyz", Hook: "read", Action: "full"}
	_ = analytics.Record(e)
	data, _ := os.ReadFile(analytics.AnalyticsPath())
	var got analytics.Event
	_ = json.Unmarshal([]byte(strings.TrimSpace(string(data))), &got)
	if len(got.SID) > 16 {
		t.Errorf("SID should be truncated to 16 chars, got %q", got.SID)
	}
}

func BenchmarkRecord(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	e := analytics.Event{
		TS: 1753182000000000000, SID: "abc123",
		Hook: "read", Action: "unchanged",
		BytesIn: 8192, BytesOut: 95, DurNS: 312000,
	}
	b.ResetTimer()
	for b.Loop() {
		_ = analytics.Record(e)
	}
}
