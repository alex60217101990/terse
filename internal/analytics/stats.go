package analytics

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/alex60217101990/terse/internal/bytesconv"
)

// HookAgg aggregates calls and bytes for one hook.
type HookAgg struct {
	Count     int `json:"count"`
	BytesIn   int `json:"bytes_in"`
	BytesOut  int `json:"bytes_out"`
	TokensIn  int `json:"tokens_in"`
	TokensOut int `json:"tokens_out"`
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
	TotalTokensIn    int                       `json:"total_tokens_in"`
	TotalTokensOut   int                       `json:"total_tokens_out"`
	// EstimatedTokens is true when any counted event predates token recording,
	// so its tokens had to be estimated as bytes/4. Reported, never hidden: the
	// two eras must not be silently averaged together.
	EstimatedTokens bool `json:"estimated_tokens"`
}

// LatStat holds latency percentiles in milliseconds.
type LatStat struct {
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

func (s Stats) SavedBytes() int     { return s.TotalBytesIn - s.TotalBytesOut }
func (s Stats) SavedTokens() int    { return s.TotalTokensIn - s.TotalTokensOut }
func (s Stats) OriginalTokens() int { return s.TotalTokensIn }

// eventTokens returns an events token counts, falling back to the old bytes/4
// estimate for events recorded before token counting existed. The third result
// reports whether the fallback was used, so a mixed window can say so rather
// than presenting an average of two different units as one number.
func eventTokens(e Event) (in, out int, estimated bool) {
	if e.TokensIn == 0 && e.TokensOut == 0 && (e.BytesIn != 0 || e.BytesOut != 0) {
		return e.BytesIn / 4, e.BytesOut / 4, true
	}
	return e.TokensIn, e.TokensOut, false
}

// humanTokens formats a token count compactly: 1234 -> "1.2k", 4_500_000 -> "4.5M".
func humanTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1e6)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1e3)
	default:
		return strconv.Itoa(n)
	}
}

func (s Stats) SavingsPercent() float64 {
	if s.TotalBytesIn == 0 {
		return 0
	}
	return float64(s.SavedBytes()) / float64(s.TotalBytesIn) * 100
}

// ComputeStats aggregates a slice of events into Stats.
// hookAlias folds legacy lowercase per-tool hook names (written before the
// universal pipeline reported the canonical Claude tool_name) onto their
// canonical form, so old and new analytics records for the same tool don't
// show up as duplicate rows (e.g. "bash" + "Bash").
var hookAlias = map[string]string{
	"read": "Read", "write": "Write", "bash": "Bash", "glob": "Glob", "grep": "Grep",
}

// contextHooks are lifecycle hooks that inject context (a restoration manifest
// across compaction, a session-start preamble) rather than compress a tool's
// output. Their byte deltas are context-management overhead — often a net add
// — so they are excluded from the headline compression ratio and reported
// separately.
var contextHooks = map[string]bool{
	"precompact": true, "postcompact": true, "sessionstart": true,
}

// normalizeHook maps a raw Event.Hook to its canonical display name.
func normalizeHook(h string) string {
	if c, ok := hookAlias[h]; ok {
		return c
	}
	return h
}

func ComputeStats(events []Event) Stats {
	s := Stats{
		ByHookAction: make(map[string]map[string]int),
		ByHook:       make(map[string]HookAgg),
		Latencies:    make(map[string][]int64),
		LatencyStats: make(map[string]LatStat),
	}
	for _, e := range events {
		hook := normalizeHook(e.Hook)
		s.TotalInvocations++
		// Headline compression totals count only output-compression hooks.
		// Context-injection hooks (precompact/postcompact/sessionstart) add
		// bytes on purpose to survive compaction — folding their overhead into
		// the compression ratio would misreport it. They're reported in their
		// own section instead.
		if !contextHooks[hook] {
			s.TotalBytesIn += e.BytesIn
			s.TotalBytesOut += e.BytesOut
			ti, to, est := eventTokens(e)
			s.TotalTokensIn += ti
			s.TotalTokensOut += to
			s.EstimatedTokens = s.EstimatedTokens || est
		}
		key := hook + "/" + e.Action
		if s.ByHookAction[hook] == nil {
			s.ByHookAction[hook] = make(map[string]int)
		}
		s.ByHookAction[hook][e.Action]++
		agg := s.ByHook[hook]
		agg.Count++
		agg.BytesIn += e.BytesIn
		agg.BytesOut += e.BytesOut
		ati, ato, _ := eventTokens(e)
		agg.TokensIn += ati
		agg.TokensOut += ato
		s.ByHook[hook] = agg
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
		// bufio.Reader (not Scanner) so a single oversized/corrupt line can't
		// abort the whole read: Scanner errors with ErrTooLong past its 64 KB
		// cap and stops, which used to fail `stats` wholesale. Analytics is
		// best-effort diagnostics — skip an unparseable line, keep the rest.
		r := bufio.NewReader(f)
		for {
			line, rerr := r.ReadString('\n')
			if len(line) > 0 {
				var e Event
				// Zero-copy: json.Unmarshal reads the bytes read-only and copies
				// every string field out, so aliasing `line` (a fresh owned
				// string from ReadString) via S2B retains nothing.
				if json.Unmarshal(bytesconv.S2B(line), &e) == nil && (cutoff == 0 || e.TS >= cutoff) {
					events = append(events, e)
				}
			}
			if rerr != nil { // io.EOF or read error: stop this file, keep events
				break
			}
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

func clamp01(f float64) float64 {
	return min(max(f, 0), 1)
}

// impactFrac maps a per-hook saved byte count onto [0,1] with a sqrt
// (perceptual) scale so a single dominant hook does not collapse every other
// bar to empty. A non-zero saved always yields at least one cell (handled by
// the caller's meter width); exactly-zero yields 0.
func impactFrac(saved, maxSaved int) float64 {
	if saved <= 0 || maxSaved <= 0 {
		return 0
	}
	return math.Sqrt(float64(saved)) / math.Sqrt(float64(maxSaved))
}

// --- alternate bar styles (for `barsdemo`) ---

// meterLine: a thick bar rule (▬, ~50% heavier than ━) fading to a thin dim
// track (─); the thick→thin boundary reads as a soft tip. Cargo/npm-ish.
func (c colorizer) meterLine(frac float64, width int) string {
	frac = clamp01(frac)
	full := int(frac * float64(width))
	if frac > 0 && full == 0 {
		full = 1
	}
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

// meterBraille: braille dot-matrix fill (finest sub-cell resolution).
func (c colorizer) meterBraille(frac float64, width int) string {
	frac = clamp01(frac)
	full := int(frac * float64(width))
	if frac > 0 && full == 0 {
		full = 1
	}
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

// meterFn returns the bar renderer for a style name. Default is the thick line;
// "braille" switches to the dot-matrix style.
func (c colorizer) meterFn(style string) func(float64, int) string {
	if style == "braille" {
		return c.meterBraille
	}
	return c.meterLine
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

	// Per-hook tables. Compression hooks (tool output) carry a saved/%/impact
	// bar; context-injection hooks (precompact/postcompact/sessionstart) add
	// bytes on purpose, so they get their own "added" column and never
	// contaminate the impact scale. Impact is the last column so its ANSI
	// codes don't throw off tabwriter's alignment.
	if len(s.ByHook) > 0 {
		var comp, ctx []string
		maxSaved := 1
		for h, a := range s.ByHook {
			if contextHooks[h] {
				ctx = append(ctx, h)
				continue
			}
			comp = append(comp, h)
			if sv := a.BytesIn - a.BytesOut; sv > maxSaved {
				maxSaved = sv
			}
		}
		slices.Sort(comp)
		slices.Sort(ctx)

		if len(comp) > 0 {
			fmt.Fprintf(w, "  %s\n", c.bold("BY HOOK"))
			tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
			fmt.Fprintln(tw, "    hook\tcalls\tsaved\t%\timpact")
			for _, h := range comp {
				a := s.ByHook[h]
				saved := a.BytesIn - a.BytesOut
				pct := 0.0
				if a.BytesIn > 0 {
					pct = float64(saved) / float64(a.BytesIn) * 100
				}
				fmt.Fprintf(tw, "    %s\t%d\t%s\t%.0f%%\t%s\n",
					h, a.Count, FormatBytes(saved), pct, meter(impactFrac(saved, maxSaved), 12))
			}
			_ = tw.Flush()
			fmt.Fprintln(w)
		}

		if len(ctx) > 0 {
			fmt.Fprintf(w, "  %s\n", c.bold("CONTEXT INJECTION"))
			tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
			fmt.Fprintln(tw, "    hook\tcalls\tadded")
			for _, h := range ctx {
				a := s.ByHook[h]
				// Manifests add context deliberately; report bytes added
				// (out-in), so this reads as cost, not a bogus "saving".
				fmt.Fprintf(tw, "    %s\t%d\t%s\n", h, a.Count, FormatBytes(a.BytesOut-a.BytesIn))
			}
			_ = tw.Flush()
			fmt.Fprintln(w)
		}
	}

	// Latency percentiles.
	if len(s.LatencyStats) > 0 {
		fmt.Fprintf(w, "  %s  p50 / p95 / p99\n", c.bold("LATENCY"))
		keys := slices.Sorted(maps.Keys(s.LatencyStats))
		for _, k := range keys {
			ls := s.LatencyStats[k]
			fmt.Fprintf(w, "    %-28s %.1f / %.1f / %.1f ms\n", k, ls.P50, ls.P95, ls.P99)
		}
		fmt.Fprintln(w)
	}
}
