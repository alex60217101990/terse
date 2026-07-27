package detect

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// A large single JSON object (a config or API dump) doesn't fit the columnar
// array summary (AnalyzeJSONArray/ColumnarSummary): there is no row set to
// aggregate, just one object with a handful of top-level keys, some scalar,
// some nested. SummarizeJSONObject instead reduces it to a key schema —
// scalars verbatim (long strings truncated), arrays reduced to a
// name/count/shape line, nested objects inlined or counted.
const (
	jsonObjectMinBytes  = 1024 // below this a full parse+summary can't pay off
	jsonObjectGuardScan = 64   // leading bytes scanned for the cheap '{' guard
	jsonObjectMaxKeys   = 200  // bound pathological objects with huge key counts
	jsonObjectScalarMax = 64   // scalars at or under this many bytes print verbatim
	jsonObjectTruncRune = 48   // longer strings/numbers truncate to this many runes
	jsonObjectInlineMax = 4    // nested objects with at most this many keys inline
)

// looksLikeJSONObject is the cheap pre-parse guard: content must be at least
// jsonObjectMinBytes long and its first non-space byte (scanned over at most
// jsonObjectGuardScan leading bytes) must be '{'. No allocation — no
// strings.TrimSpace, no decoder — so a non-matching call (the common case for
// arbitrary tool output) costs nothing.
func looksLikeJSONObject(s string) bool {
	if len(s) < jsonObjectMinBytes {
		return false
	}
	n := min(len(s), jsonObjectGuardScan)
	for i := range n {
		switch s[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			return true
		default:
			return false
		}
	}
	return false
}

// SummarizeJSONObject reduces a large top-level JSON object to a key schema:
// scalars printed verbatim (long strings/numbers rune-safe truncated), arrays
// reduced to a shape line, small nested objects inlined. Returns "" when the
// cheap guard rejects content, the content doesn't parse as a JSON object, or
// the summary would not be strictly smaller than the input (never-worse).
func SummarizeJSONObject(content string) string {
	if !looksLikeJSONObject(content) {
		return ""
	}
	dec := json.NewDecoder(strings.NewReader(content))
	dec.UseNumber() // print numbers as written, not float-mangled
	var m map[string]any
	if err := dec.Decode(&m); err != nil {
		return ""
	}
	out := renderJSONObjectSummary(m)
	if len(out) >= len(content) {
		return ""
	}
	return out
}

func renderJSONObjectSummary(m map[string]any) string {
	keys := slices.Sorted(maps.Keys(m))
	var b strings.Builder
	fmt.Fprintf(&b, "[JSON OBJECT — %d top-level keys]\n", len(keys))

	shown, extra := keys, 0
	if len(keys) > jsonObjectMaxKeys {
		shown = keys[:jsonObjectMaxKeys]
		extra = len(keys) - jsonObjectMaxKeys
	}
	for _, k := range shown {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(formatValue(m[k]))
		b.WriteByte('\n')
	}
	if extra > 0 {
		fmt.Fprintf(&b, "… +%d more keys\n", extra)
	}
	return b.String()
}

// formatValue renders one JSON value per the summary's display rules.
func formatValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case bool:
		if t {
			return "true"
		}
		return "false"
	case json.Number:
		return truncateScalar(string(t), false)
	case string:
		return truncateScalar(t, true)
	case map[string]any:
		return formatNestedObject(t)
	case []any:
		return formatArray(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// truncateScalar prints s verbatim (quoted, if it's a string) when it's at
// most jsonObjectScalarMax bytes; longer values are rune-safe truncated to
// jsonObjectTruncRune runes with a "…(+n bytes)" suffix, n being the bytes
// dropped from the original.
func truncateScalar(s string, quoted bool) string {
	if len(s) <= jsonObjectScalarMax {
		if quoted {
			return strconv.Quote(s)
		}
		return s
	}
	runes := []rune(s)
	cut := min(jsonObjectTruncRune, len(runes))
	prefix := string(runes[:cut])
	removed := len(s) - len(prefix)
	if removed <= 0 {
		if quoted {
			return strconv.Quote(s)
		}
		return s
	}
	suffix := "…(+" + strconv.Itoa(removed) + " bytes)"
	if quoted {
		return `"` + prefix + suffix + `"`
	}
	return prefix + suffix
}

// formatNestedObject inlines an object of at most jsonObjectInlineMax keys as
// "{k:v, …}"; larger objects collapse to "object{K keys}" so a deep/wide
// config doesn't blow the summary back up to original size.
func formatNestedObject(m map[string]any) string {
	if len(m) > jsonObjectInlineMax {
		return fmt.Sprintf("object{%d keys}", len(m))
	}
	keys := slices.Sorted(maps.Keys(m))
	var b strings.Builder
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(k)
		b.WriteByte(':')
		b.WriteString(formatValue(m[k]))
	}
	b.WriteByte('}')
	return b.String()
}

// formatArray reduces an array to a shape line. Arrays of at least 5 objects
// get a schema derived from the first element's keys/kinds
// ("array[N] — {key:kind,…}"); everything else (short arrays, arrays of
// scalars) is just "array[N]".
func formatArray(arr []any) string {
	n := len(arr)
	if n >= 5 {
		if first, ok := arr[0].(map[string]any); ok && len(first) > 0 {
			keys := slices.Sorted(maps.Keys(first))
			var b strings.Builder
			fmt.Fprintf(&b, "array[%d] — {", n)
			for i, k := range keys {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(k)
				b.WriteByte(':')
				b.WriteString(kindOf(first[k]))
			}
			b.WriteByte('}')
			return b.String()
		}
	}
	return fmt.Sprintf("array[%d]", n)
}

// kindOf names a decoded JSON value's type for the array-schema line.
func kindOf(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case string:
		return "string"
	case json.Number:
		if strings.ContainsAny(string(t), ".eE") {
			return "float"
		}
		return "int"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return "string"
	}
}
