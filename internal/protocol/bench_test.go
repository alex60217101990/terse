package protocol_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// BenchmarkToolResponseUnmarshalContent10KB measures decoding a tool_response
// envelope whose "content" is a 10KB plain string — the dominant shape on
// non-Read traffic. It exists to gate the RawMessage-copy removal in
// ToolResponse.UnmarshalJSON: the anon-struct decode path used to route
// "content" through json.RawMessage (a full copy of the escaped bytes) before
// contentText re-Unmarshaled them into a string. Custom field decoding should
// cut B/op materially versus that baseline.
func BenchmarkToolResponseUnmarshalContent10KB(b *testing.B) {
	// 10KB of content that needs JSON escaping (quotes, newlines, backslash)
	// so the benchmark also exercises the unescape path, not just raw copies.
	var raw strings.Builder
	for raw.Len() < 10*1024 {
		raw.WriteString(`line with "quotes" and \backslash\ and a tab\tend` + "\n")
	}
	content := raw.String()

	buf, err := json.Marshal(content)
	if err != nil {
		b.Fatalf("marshal content: %v", err)
	}
	envelope := []byte(`{"content":` + string(buf) + `,"stdout":"","stderr":"","output":""}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tr protocol.ToolResponse
		if err := tr.UnmarshalJSON(envelope); err != nil {
			b.Fatalf("UnmarshalJSON: %v", err)
		}
		if tr.Text() == "" {
			b.Fatal("empty text")
		}
	}
}
