package hook

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// Corpus command sizes, measured over 23,857 cappable Bash calls: p50 160
// bytes, mean 362, p90 793. benchCmd is the p50 shape; benchCmdMean the mean.
var (
	benchCmd     = cmdOfSize(160)
	benchCmdMean = cmdOfSize(362)
)

func cmdOfSize(n int) string {
	var b strings.Builder
	for b.Len() < n {
		b.WriteString(`GOWORK=off go test ./internal/hook/ -run 'TestSomething' -count=1 && echo "step done"; `)
	}
	return b.String()[:n]
}

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
// Measured (M5 Pro, count=6) at the corpus p50 command: 620 ns/op, 0 allocs;
// at the mean command, 1,253 ns/op, 0 allocs. Both are well under the 5 µs
// budget and are 1-2% of the 59.7 µs hook roundtrip.
//
// The first version of this path cost 2,840 ns and 13 allocations on a
// 53-byte command — a size the corpus does not actually have. What moved
// since: os.MkdirAll ran per call and was 92% of the CPU (now once per HOME),
// encoding/json's reflection wrote the response (now a fused emitter over the
// same segment table wrapCommand uses), and the capture id was a sha256 over
// the whole command (now the session hash plus a counter, since the id only
// has to be unique).
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

// BenchmarkHandleBashPreToolUse_Mean is the same path at the corpus mean
// command size, where the wrapper embeds 362 bytes twice.
func BenchmarkHandleBashPreToolUse_Mean(b *testing.B) {
	b.Setenv("HOME", b.TempDir())
	raw, err := json.Marshal(map[string]string{"command": benchCmdMean})
	if err != nil {
		b.Fatal(err)
	}
	inp := &protocol.HookInput{
		SessionID: "bench-session-0123456789abcdef",
		ToolName:  "Bash",
		ToolInput: raw,
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := handleBashPreToolUse(inp, io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
