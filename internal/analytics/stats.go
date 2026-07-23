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
	"text/tabwriter"
	"time"
)

// HookAgg aggregates calls and bytes for one hook.
type HookAgg struct {
	Count    int `json:"count"`
	BytesIn  int `json:"bytes_in"`
	BytesOut int `json:"bytes_out"`
}

// Stats holds aggregated analytics data.
type Stats struct {
	ByHookAction     map[string]map[string]int `json:"by_hook_action"` // hook -> action -> count
	ByHook           map[string]HookAgg        `json:"by_hook"`        // hook -> calls + bytes
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

// humanTokens formats a token count compactly: 1234 -> "1.2k", 4_500_000 -> "4.5M".
func humanTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1e3)
	default:
		return fmt.Sprintf("%d", n)
	}
}

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
		ByHook:       make(map[string]HookAgg),
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
		agg := s.ByHook[e.Hook]
		agg.Count++
		agg.BytesIn += e.BytesIn
		agg.BytesOut += e.BytesOut
		s.ByHook[e.Hook] = agg
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

// LoadEvents reads events from analytics.jsonl (and .1 if exists). When days > 0,
// only events newer than that many days are returned; days <= 0 returns all.
func LoadEvents(days int) ([]Event, error) {
	path := AnalyticsPath()
	var cutoff int64
	if days > 0 {
		cutoff = time.Now().Add(-time.Duration(days) * 24 * time.Hour).UnixNano()
	}
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
			if cutoff > 0 && e.TS < cutoff {
				continue
			}
			events = append(events, e)
		}
		if err := scanner.Err(); err != nil {
			_ = f.Close()
			return nil, err
		}
		_ = f.Close()
	}
	return events, nil
}

// colorizer emits ANSI styling, gated so piped/redirected output stays plain.
type colorizer struct{ on, truecolor bool }

func newColorizer() colorizer {
	fi, err := os.Stdout.Stat()
	tty := err == nil && fi.Mode()&os.ModeCharDevice != 0
	on := tty && os.Getenv("NO_COLOR") == ""
	ct := os.Getenv("COLORTERM")
	return colorizer{on: on, truecolor: on && (strings.Contains(ct, "truecolor") || strings.Contains(ct, "24bit"))}
}

func (c colorizer) wrap(code, s string) string {
	if !c.on {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}
func (c colorizer) bold(s string) string { return c.wrap("1", s) }
func (c colorizer) dim(s string) string  { return c.wrap("2", s) }

// partialBlocks are the left-aligned eighth-width block glyphs, index 1..7
// giving a fractional final cell so the bar length is smooth, not chunky.
var partialBlocks = [8]string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// meter renders a rounded "pill" progress bar: half-circle end caps (◖ ◗) around
// a fractional eighth-block fill. Filled cells use a teal→green truecolor
// gradient when supported, else solid green; the remainder is dim. Always
// exactly `width` cells wide so table columns stay aligned.
func (c colorizer) meter(frac float64, width int) string {
	if frac < 0 {
		frac = 0
	} else if frac > 1 {
		frac = 1
	}
	if width < 3 {
		return strings.Repeat("█", width)
	}
	inner := width - 2 // two cells go to the rounded caps

	units := frac * float64(inner)
	full := int(units)
	rem := int((units-float64(full))*8 + 0.5)
	if rem == 8 {
		full++
		rem = 0
	}

	var b strings.Builder
	// Left cap: part of the fill when there's any progress, else dim track.
	if frac > 0 {
		b.WriteString(c.barCell("◖", 0, width))
	} else {
		b.WriteString(c.dim("◖"))
	}
	cells := 0
	for ; cells < full && cells < inner; cells++ {
		b.WriteString(c.barCell("█", cells+1, width))
	}
	if rem > 0 && cells < inner {
		b.WriteString(c.barCell(partialBlocks[rem], cells+1, width))
		cells++
	}
	if cells < inner {
		b.WriteString(c.dim(strings.Repeat("░", inner-cells)))
	}
	// Right cap: filled colour only when the bar is (essentially) full.
	if frac >= 0.999 {
		b.WriteString(c.barCell("◗", width-1, width))
	} else {
		b.WriteString(c.dim("◗"))
	}
	return b.String()
}

// barCell colors one filled cell: a teal→green gradient across the bar under
// truecolor, plain green otherwise, and uncolored when color is off.
func (c colorizer) barCell(ch string, i, width int) string {
	switch {
	case !c.on:
		return ch
	case !c.truecolor:
		return "\x1b[32m" + ch + "\x1b[0m"
	default:
		t := float64(i) / float64(max(width-1, 1))
		r := int(40 + (120-40)*t)
		g := int(200 + (235-200)*t)
		bl := int(165 + (120-165)*t)
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, bl, ch)
	}
}

// PrintStats writes formatted stats to w. If jsonOut, writes JSON.
func PrintStats(s Stats, jsonOut bool, w io.Writer) {
	if jsonOut {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(s)
		return
	}

	c := newColorizer()

	fmt.Fprintf(w, "\n  %s  ·  %d invocations\n\n", c.bold("qdf-hook"), s.TotalInvocations)

	// Token savings + efficiency meter.
	fmt.Fprintf(w, "  %s\n", c.bold("TOKEN SAVINGS"))
	fmt.Fprintf(w, "    Original    %9s   ~%s tok\n", FormatBytes(s.TotalBytesIn), humanTokens(s.OriginalTokens()))
	fmt.Fprintf(w, "    Emitted     %9s   ~%s tok\n", FormatBytes(s.TotalBytesOut), humanTokens(s.TotalBytesOut/4))
	fmt.Fprintf(w, "    Saved       %9s   ~%s tok\n", FormatBytes(s.SavedBytes()), humanTokens(s.SavedTokens()))
	fmt.Fprintf(w, "    Efficiency  %s  %.1f%%\n\n", c.meter(s.SavingsPercent()/100, 24), s.SavingsPercent())

	// Per-hook table with impact bars (impact = saved bytes relative to the
	// busiest hook). Impact is the last column so its ANSI codes don't throw
	// off tabwriter's column alignment.
	if len(s.ByHook) > 0 {
		fmt.Fprintf(w, "  %s\n", c.bold("BY HOOK"))
		hooks := make([]string, 0, len(s.ByHook))
		maxSaved := 1
		for h, a := range s.ByHook {
			hooks = append(hooks, h)
			if sv := a.BytesIn - a.BytesOut; sv > maxSaved {
				maxSaved = sv
			}
		}
		sort.Strings(hooks)
		tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
		fmt.Fprintln(tw, "    hook\tcalls\tsaved\t%\timpact")
		for _, h := range hooks {
			a := s.ByHook[h]
			saved := a.BytesIn - a.BytesOut
			pct := 0.0
			if a.BytesIn > 0 {
				pct = float64(saved) / float64(a.BytesIn) * 100
			}
			fmt.Fprintf(tw, "    %s\t%d\t%s\t%.0f%%\t%s\n",
				h, a.Count, FormatBytes(saved), pct, c.meter(float64(saved)/float64(maxSaved), 12))
		}
		_ = tw.Flush()
		fmt.Fprintln(w)
	}

	// Latency percentiles.
	if len(s.LatencyStats) > 0 {
		fmt.Fprintf(w, "  %s  p50 / p95 / p99\n", c.bold("LATENCY"))
		keys := make([]string, 0, len(s.LatencyStats))
		for k := range s.LatencyStats {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			ls := s.LatencyStats[k]
			fmt.Fprintf(w, "    %-28s %.1f / %.1f / %.1f ms\n", k, ls.P50, ls.P95, ls.P99)
		}
		fmt.Fprintln(w)
	}
}
