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

// meter renders the default progress bar: a fractional eighth-block fill with a
// teal→green truecolor gradient (solid green without truecolor, plain when not
// a TTY); the remainder is a dim track. Always exactly `width` cells wide.
func (c colorizer) meter(frac float64, width int) string {
	frac = clamp01(frac)
	units := frac * float64(width)
	full := int(units)
	rem := int((units-float64(full))*8 + 0.5)
	if rem == 8 {
		full++
		rem = 0
	}
	var b strings.Builder
	cells := 0
	for ; cells < full && cells < width; cells++ {
		b.WriteString(c.barCell("█", cells, width))
	}
	if rem > 0 && cells < width {
		b.WriteString(c.barCell(partialBlocks[rem], cells, width))
		cells++
	}
	if cells < width {
		b.WriteString(c.dim(strings.Repeat("░", width-cells)))
	}
	return b.String()
}

func clamp01(f float64) float64 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// --- alternate bar styles (for `barsdemo`) ---

// meterLine: a thick bar rule (▬, ~50% heavier than ━) fading to a thin dim
// track (─); the thick→thin boundary reads as a soft tip. Cargo/npm-ish.
func (c colorizer) meterLine(frac float64, width int) string {
	frac = clamp01(frac)
	full := int(frac * float64(width))
	var b strings.Builder
	i := 0
	for ; i < full && i < width; i++ {
		b.WriteString(c.barCell("▬", i, width))
	}
	if i < width {
		b.WriteString(c.dim(strings.Repeat("─", width-i)))
	}
	return b.String()
}

// meterShade: solid fill fading through ▓▒░ at the boundary.
func (c colorizer) meterShade(frac float64, width int) string {
	frac = clamp01(frac)
	full := int(frac * float64(width))
	var b strings.Builder
	for i := 0; i < width; i++ {
		switch {
		case i < full:
			b.WriteString(c.barCell("█", i, width))
		case i == full:
			b.WriteString(c.barCell("▓", i, width))
		case i == full+1:
			b.WriteString(c.dim("▒"))
		default:
			b.WriteString(c.dim("░"))
		}
	}
	return b.String()
}

// meterBraille: braille dot-matrix fill (finest sub-cell resolution).
func (c colorizer) meterBraille(frac float64, width int) string {
	frac = clamp01(frac)
	full := int(frac * float64(width))
	var b strings.Builder
	i := 0
	for ; i < full && i < width; i++ {
		b.WriteString(c.barCell("⣿", i, width))
	}
	if i < width && frac > 0 && frac < 1 {
		b.WriteString(c.barCell("⣄", i, width))
		i++
	}
	if i < width {
		b.WriteString(c.dim(strings.Repeat("⠄", width-i)))
	}
	return b.String()
}

// RenderBarGallery prints every bar style at several fill levels so the user can
// pick one by eye in their own terminal (with real color/gradient).
func RenderBarGallery(w io.Writer) {
	c := newColorizer()
	const width = 22
	fracs := []float64{0.12, 0.37, 0.65, 0.88, 1.0}
	styles := []struct {
		n    int
		name string
		fn   func(float64, int) string
	}{
		{1, "blocks  (smooth eighth-block fill)", c.meter},
		{2, "line    (heavy rule + soft tip)", c.meterLine},
		{3, "shade   (solid, ▓▒░ fade tail)", c.meterShade},
		{4, "braille (dot-matrix, finest)", c.meterBraille},
	}
	fmt.Fprintf(w, "\n  %s — pick a style number\n\n", c.bold("qdf-hook bar styles"))
	for _, s := range styles {
		fmt.Fprintf(w, "  %s %s\n", c.bold(fmt.Sprintf("[%d]", s.n)), s.name)
		for _, f := range fracs {
			fmt.Fprintf(w, "       %s  %3.0f%%\n", s.fn(f, width), f*100)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintf(w, "  Set with: %s\n\n", c.dim("qdf-hook stats --style=<blocks|line|shade|braille>  (or tell me your pick)"))
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

// meterFn returns the bar renderer for a style name (default "blocks").
func (c colorizer) meterFn(style string) func(float64, int) string {
	switch style {
	case "line":
		return c.meterLine
	case "shade":
		return c.meterShade
	case "braille":
		return c.meterBraille
	default:
		return c.meter
	}
}

// PrintStats writes formatted stats to w. If jsonOut, writes JSON. style selects
// the progress-bar look (blocks|line|shade|braille).
func PrintStats(s Stats, jsonOut bool, style string, w io.Writer) {
	if jsonOut {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(s)
		return
	}

	c := newColorizer()
	meter := c.meterFn(style)

	fmt.Fprintf(w, "\n  %s  ·  %d invocations\n\n", c.bold("qdf-hook"), s.TotalInvocations)

	// Token savings + efficiency meter.
	fmt.Fprintf(w, "  %s\n", c.bold("TOKEN SAVINGS"))
	fmt.Fprintf(w, "    Original    %9s   ~%s tok\n", FormatBytes(s.TotalBytesIn), humanTokens(s.OriginalTokens()))
	fmt.Fprintf(w, "    Emitted     %9s   ~%s tok\n", FormatBytes(s.TotalBytesOut), humanTokens(s.TotalBytesOut/4))
	fmt.Fprintf(w, "    Saved       %9s   ~%s tok\n", FormatBytes(s.SavedBytes()), humanTokens(s.SavedTokens()))
	fmt.Fprintf(w, "    Efficiency  %s  %.1f%%\n\n", meter(s.SavingsPercent()/100, 24), s.SavingsPercent())

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
				h, a.Count, FormatBytes(saved), pct, meter(float64(saved)/float64(maxSaved), 12))
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
