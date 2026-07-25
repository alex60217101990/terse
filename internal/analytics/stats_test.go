package analytics_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alex60217101990/terse/internal/analytics"
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

func TestComputeStats_FoldsLegacyToolAliases(t *testing.T) {
	// Legacy lowercase per-tool records must merge with canonical tool_name
	// records instead of showing as duplicate rows (bash + Bash).
	ev := []analytics.Event{
		{Hook: "bash", Action: "summary", BytesIn: 1000, BytesOut: 100},
		{Hook: "Bash", Action: "summary", BytesIn: 2000, BytesOut: 200},
	}
	s := analytics.ComputeStats(ev)
	if _, dup := s.ByHook["bash"]; dup {
		t.Error("legacy 'bash' must fold into 'Bash', not appear separately")
	}
	agg, ok := s.ByHook["Bash"]
	if !ok || agg.Count != 2 || agg.BytesIn != 3000 {
		t.Errorf("folded Bash agg wrong: %+v ok=%v", agg, ok)
	}
}

func TestComputeStats_ContextHooksExcludedFromSavings(t *testing.T) {
	ev := []analytics.Event{
		{Hook: "Bash", Action: "summary", BytesIn: 1000, BytesOut: 100},          // compression
		{Hook: "postcompact", Action: "postcompact", BytesIn: 0, BytesOut: 5000}, // context add
	}
	s := analytics.ComputeStats(ev)
	// The 5000-byte manifest must NOT drag the headline compression ratio.
	if s.TotalBytesIn != 1000 || s.TotalBytesOut != 100 {
		t.Errorf("context hook leaked into headline totals: in=%d out=%d", s.TotalBytesIn, s.TotalBytesOut)
	}
	if _, ok := s.ByHook["postcompact"]; !ok {
		t.Error("postcompact should still be tracked in ByHook")
	}
	if s.TotalInvocations != 2 {
		t.Errorf("all invocations still counted, got %d", s.TotalInvocations)
	}
}

func TestFormatBytes_Negative(t *testing.T) {
	if got := analytics.FormatBytes(-14344); got != "-14.0 KB" {
		t.Errorf("negative bytes: got %q, want %q", got, "-14.0 KB")
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

func TestImpactBar_SkewedDistributionStaysVisible(t *testing.T) {
	// Read dominates; Edit saves far less in absolute bytes but is non-zero.
	// Before the fix, Edit's bar floored to zero cells. After: >=1 cell.
	ev := []analytics.Event{
		{Hook: "Read", Action: "unchanged", BytesIn: 24_000_000, BytesOut: 900_000},
		{Hook: "Edit", Action: "delta", BytesIn: 907_000, BytesOut: 8_400},
		{Hook: "Agent", Action: "full", BytesIn: 1000, BytesOut: 1000},
	}
	stats := analytics.ComputeStats(ev)
	var buf strings.Builder
	analytics.PrintStats(stats, false, "blocks", &buf)
	out := buf.String()

	editLine := lineContaining(t, out, "Edit")
	if !strings.Contains(editLine, "▬") {
		t.Errorf("non-zero hook Edit must show >=1 filled cell, got:\n%s", editLine)
	}
	agentLine := lineContaining(t, out, "Agent")
	if strings.Contains(agentLine, "▬") {
		t.Errorf("zero-saved hook Agent must show no filled cell, got:\n%s", agentLine)
	}
}

func lineContaining(t *testing.T, s, sub string) string {
	t.Helper()
	for line := range strings.SplitSeq(s, "\n") {
		if strings.Contains(line, sub) {
			return line
		}
	}
	t.Fatalf("no line containing %q in:\n%s", sub, s)
	return ""
}

// TestLoadEvents_OversizedLineDoesNotAbort: a single >64KB (corrupt) line in
// analytics.jsonl must not fail the whole `stats` read — the valid events
// around it are still returned. (bufio.Scanner's 64KB cap used to abort.)
func TestLoadEvents_OversizedLineDoesNotAbort(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := analytics.AnalyticsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	ts := time.Now().UnixNano()
	good1 := fmt.Sprintf(`{"sid":"s","hook":"Bash","action":"ref","ts":%d,"bi":100,"bo":10,"dur":1}`, ts)
	good2 := fmt.Sprintf(`{"sid":"s","hook":"Read","action":"full","ts":%d,"bi":200,"bo":200,"dur":2}`, ts)
	huge := strings.Repeat("x", 200*1024) // 200 KB garbage line
	content := good1 + "\n" + huge + "\n" + good2 + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	events, err := analytics.LoadEvents(0)
	if err != nil {
		t.Fatalf("LoadEvents must not fail on an oversized line: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("want 2 valid events around the garbage line, got %d", len(events))
	}
}
