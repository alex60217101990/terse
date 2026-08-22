package hook

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// The corpus median Bash command is about 50 bytes; this is that shape.
const benchCmd = "GOWORK=off go test ./internal/hook/ -run TestSomething"

// BenchmarkCappable measures the cost of the capping decision gate that runs
// before every Bash command. A single forward pass, no allocation or regexp.
// Measured (M5 Pro, count=10): ~24.8 ns/op, 0 B/op, 0 allocs/op.
// Well under the 5 µs budget; no further optimization warranted.
func BenchmarkCappable(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = cappable(benchCmd)
	}
}

// BenchmarkWrapCommand measures the cost of rewriting a Bash command to add
// output capping scaffolding. The cap size is written straight into the
// buffer via strconv.AppendInt rather than through an intermediate
// strconv.Itoa string, so this path allocates nothing. Measured (M5 Pro,
// count=10): ~36.4 ns/op, 0 B/op, 0 allocs/op.
// Well under the 5 µs budget; no further optimization warranted.
func BenchmarkWrapCommand(b *testing.B) {
	buf := make([]byte, 0, 1024)
	result := wrapCommand(buf[:0], benchCmd, "/Users/x/.qdf-hook/captures/abc.out", "abc", 1600)
	if result == nil {
		b.Fatal("wrapCommand returned nil")
	}
	b.ReportAllocs()
	for b.Loop() {
		buf = wrapCommand(buf[:0], benchCmd, "/Users/x/.qdf-hook/captures/abc.out", "abc", 1600)
	}
	_ = buf
}

// BenchmarkHandleBashPreToolUse measures the whole per-Bash-call path the
// daemon runs: decode the tool input, decide, hash the capture id, ensure the
// capture directory, rewrite, and encode the response.
func BenchmarkHandleBashPreToolUse(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	raw, err := json.Marshal(map[string]string{"command": benchCmd})
	if err != nil {
		b.Fatal(err)
	}
	inp := &protocol.HookInput{
		SessionID: "bench-session-0123456789abcdef",
		ToolName:  "Bash",
		ToolInput: raw,
	}
	if err := handleBashPreToolUse(inp, io.Discard); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := handleBashPreToolUse(inp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
