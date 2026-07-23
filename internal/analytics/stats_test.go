package analytics_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
)

func makeEvents() []analytics.Event {
	return []analytics.Event{
		{Hook: "read", Action: "unchanged", BytesIn: 8000, BytesOut: 80, DurNS: 300_000},
		{Hook: "read", Action: "unchanged", BytesIn: 5000, BytesOut: 80, DurNS: 400_000},
		{Hook: "read", Action: "delta", BytesIn: 8000, BytesOut: 400, DurNS: 1_200_000},
		{Hook: "read", Action: "full", BytesIn: 8000, BytesOut: 8000, DurNS: 800_000},
		{Hook: "bash", Action: "summary", BytesIn: 50000, BytesOut: 200, DurNS: 20_000_000},
		{Hook: "pretooluse", Action: "pretool-unchanged", BytesIn: 0, BytesOut: 60, DurNS: 100_000},
	}
}

func TestComputeStats_TotalInvocations(t *testing.T) {
	stats := analytics.ComputeStats(makeEvents())
	if stats.TotalInvocations != 6 {
		t.Errorf("expected 6, got %d", stats.TotalInvocations)
	}
}

func TestComputeStats_BytesSaved(t *testing.T) {
	stats := analytics.ComputeStats(makeEvents())
	// Total in: 8000+5000+8000+8000+50000 = 79000
	// Total out: 80+80+400+8000+200+60 = 8820
	// Saved: 70180
	if stats.TotalBytesIn != 79000 {
		t.Errorf("expected BytesIn 79000, got %d", stats.TotalBytesIn)
	}
	if stats.SavedBytes() != 79000-8820 {
		t.Errorf("saved bytes wrong: %d", stats.SavedBytes())
	}
}

func TestComputeStats_SavingsPercent(t *testing.T) {
	stats := analytics.ComputeStats(makeEvents())
	pct := stats.SavingsPercent()
	if pct < 80 || pct > 100 {
		t.Errorf("expected > 80%% savings, got %.1f%%", pct)
	}
}

func TestPrintStats_JSON(t *testing.T) {
	stats := analytics.ComputeStats(makeEvents())
	var buf strings.Builder
	analytics.PrintStats(stats, true, "blocks", &buf)
	var m map[string]any
	if err := json.Unmarshal([]byte(buf.String()), &m); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if _, ok := m["total_invocations"]; !ok {
		t.Error("JSON should contain total_invocations")
	}
}

func TestPrintStats_Text(t *testing.T) {
	stats := analytics.ComputeStats(makeEvents())
	var buf strings.Builder
	analytics.PrintStats(stats, false, "blocks", &buf)
	out := buf.String()
	if !strings.Contains(out, "TOKEN SAVINGS") {
		t.Error("text output should contain TOKEN SAVINGS section")
	}
	if !strings.Contains(out, "LATENCY") {
		t.Error("text output should contain LATENCY section")
	}
}
