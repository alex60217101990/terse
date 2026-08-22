package protocol_test

import (
	"encoding/json/jsontext"
	"strconv"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// AppendJSONString runs the standard library's encoder with the two options
// that reproduce encoding/json (EscapeForJS, AllowInvalidUTF8). AppendQuote is
// the same quoting WITHOUT options — kept here as the floor it costs to skip
// them, and as the reason they are worth paying for: without EscapeForJS the
// bytes fork from encoding/json on U+2028/U+2029.
var escapeCases = []struct {
	name string
	s    string
}{
	{"plain", "GOWORK=off go test ./internal/hook/ -run TestSomething"},
	{"quoted", `git commit -m "fix: the thing" && echo 'done'`},
	{"wrapper", strings.Repeat(`if : 2>/dev/null > '/x/y.out'; then { ls`+"\n", 8)},
	{"unicode", "echo ünïcødé 日本語 🎉 " + strings.Repeat("x", 200)},
}

func BenchmarkAppendJSONString(b *testing.B) {
	for _, c := range escapeCases {
		b.Run(c.name, func(b *testing.B) {
			buf := make([]byte, 0, 4096)
			b.ReportAllocs()
			for b.Loop() {
				buf = protocol.AppendJSONString(buf[:0], c.s)
			}
			_ = buf
		})
	}
}

func BenchmarkJsontextAppendQuote(b *testing.B) {
	for _, c := range escapeCases {
		b.Run(c.name, func(b *testing.B) {
			buf := make([]byte, 0, 4096)
			b.ReportAllocs()
			for b.Loop() {
				var err error
				buf, err = jsontext.AppendQuote(buf[:0], c.s)
				if err != nil {
					b.Fatal(err)
				}
			}
			_ = buf
		})
	}
}

// Command sizes as they actually occur: measured over 23,857 cappable Bash
// calls, p50 is 160 bytes, the mean 362, p90 793 and p99 3,141. The wrapper
// embeds the command twice, so this is what the response encoder really pays.
func BenchmarkAppendJSONString_Sizes(b *testing.B) {
	for _, n := range []int{160, 362, 793, 3141} {
		s := cmdOf(n)
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			buf := make([]byte, 0, 8192)
			b.ReportAllocs()
			for b.Loop() {
				buf = protocol.AppendJSONString(buf[:0], s)
			}
			_ = buf
		})
	}
}

// cmdOf builds a command-shaped string of about n bytes.
func cmdOf(n int) string {
	var b strings.Builder
	for b.Len() < n {
		b.WriteString(`GOWORK=off go test ./internal/hook/ -run 'TestSomething' -count=1 && echo "step done"; `)
	}
	return b.String()[:n]
}
