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
// One pass answers both the background and the interactive-session question.
// Measured (M5 Pro, count=8): ~60 ns/op, 0 B/op, 0 allocs/op.
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
// count=8): ~37 ns/op, 0 B/op, 0 allocs/op.
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
// daemon runs: decode the tool input, decide, hash the capture id, resolve the
// capture directory, rewrite, and encode the response.
//
// Measured (M5 Pro, count=8): 2,840 ns/op and 13 allocs before the hot path was
// reworked, ~297 ns/op and 0 allocs after — 9.6x, against a 5 µs budget and a
// 59.7 µs hook roundtrip. What moved: os.MkdirAll ran per call and was 92% of
// the CPU (now once per HOME), encoding/json was the rest of it (now a fused
// emitter that escapes only the command), and the id, path and response all
// share one pooled scratch.
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
