package analytics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"sort"
	"strings"
)

// Stats holds aggregated analytics data.
type Stats struct {
	ByHookAction     map[string]map[string]int `json:"by_hook_action"` // hook -> action -> count
	Latencies        map[string][]int64        `json:"-"`              // hook/action -> []DurNS (for percentile calc)
	LatencyStats     map[string]LatStat        `json:"latency_ms"`
	TotalInvocations int                       `json:"total_invocations"`
	TotalBytesIn     int                       `json:"total_bytes_in"`
	TotalBytesOut    int                       `json:"total_bytes_out"`
}

// LatStat holds latency percentiles in milliseconds.
type LatStat struct {
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

func (s Stats) SavedBytes() int     { return s.TotalBytesIn - s.TotalBytesOut }
func (s Stats) SavedTokens() int    { return s.SavedBytes() / 4 }
func (s Stats) OriginalTokens() int { return s.TotalBytesIn / 4 }

func (s Stats) SavingsPercent() float64 {
	if s.TotalBytesIn == 0 {
		return 0
	}
	return float64(s.SavedBytes()) / float64(s.TotalBytesIn) * 100
}

// ComputeStats aggregates a slice of events into Stats.
func ComputeStats(events []Event) Stats {
	s := Stats{
		ByHookAction: make(map[string]map[string]int),
		Latencies:    make(map[string][]int64),
		LatencyStats: make(map[string]LatStat),
	}
	for _, e := range events {
		s.TotalInvocations++
		s.TotalBytesIn += e.BytesIn
		s.TotalBytesOut += e.BytesOut
		key := e.Hook + "/" + e.Action
		if s.ByHookAction[e.Hook] == nil {
			s.ByHookAction[e.Hook] = make(map[string]int)
		}
		s.ByHookAction[e.Hook][e.Action]++
		s.Latencies[key] = append(s.Latencies[key], e.DurNS)
	}
	// Compute latency percentiles.
	for key, durs := range s.Latencies {
		slices.Sort(durs)
		s.LatencyStats[key] = LatStat{
			P50: pctile(durs, 0.50) / 1e6,
			P95: pctile(durs, 0.95) / 1e6,
			P99: pctile(durs, 0.99) / 1e6,
		}
	}
	return s
}

func pctile(sorted []int64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := min(max(int(math.Ceil(p*float64(len(sorted))))-1, 0), len(sorted)-1)
	return float64(sorted[idx])
}

// LoadEvents reads events from analytics.jsonl (and .1 if exists), filtered by days.
func LoadEvents(days int) ([]Event, error) {
	path := AnalyticsPath()
	var events []Event
	for _, p := range []string{path + ".1", path} {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			var e Event
			if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
				continue
			}
			events = append(events, e)
		}
		f.Close()
	}
	return events, nil
}

// PrintStats writes formatted stats to w. If jsonOut, writes JSON.
func PrintStats(s Stats, jsonOut bool, w io.Writer) {
	if jsonOut {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(s)
		return
	}

	fmt.Fprintf(w, "\nqdf-hook  (%d invocations)\n\n", s.TotalInvocations)

	fmt.Fprintf(w, "TOKEN SAVINGS\n")
	fmt.Fprintf(w, "  Original:  %s  (~%dM tokens)\n", FormatBytes(s.TotalBytesIn), s.OriginalTokens()/1_000_000)
	fmt.Fprintf(w, "  Emitted:   %s  (~%dM tokens)\n", FormatBytes(s.TotalBytesOut), s.TotalBytesOut/4/1_000_000)
	fmt.Fprintf(w, "  Saved:     %s  %.1f%%  (~%dM tokens)\n\n",
		FormatBytes(s.SavedBytes()), s.SavingsPercent(), s.SavedTokens()/1_000_000)

	fmt.Fprintf(w, "LATENCY  p50 / p95 / p99\n")
	keys := make([]string, 0, len(s.LatencyStats))
	for k := range s.LatencyStats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		ls := s.LatencyStats[k]
		fmt.Fprintf(w, "  %-30s %.1fms / %.1fms / %.1fms\n", k+":", ls.P50, ls.P95, ls.P99)
	}

	fmt.Fprintf(w, "\nBY HOOK\n")
	hooks := make([]string, 0, len(s.ByHookAction))
	for h := range s.ByHookAction {
		hooks = append(hooks, h)
	}
	sort.Strings(hooks)
	for _, h := range hooks {
		actions := s.ByHookAction[h]
		total := 0
		for _, n := range actions {
			total += n
		}
		var parts []string
		for action, n := range actions {
			parts = append(parts, fmt.Sprintf("%s: %d", action, n))
		}
		sort.Strings(parts)
		fmt.Fprintf(w, "  %-15s %5d  (%s)\n", h+":", total, strings.Join(parts, "  "))
	}
	fmt.Fprintln(w)
}
