package protocol_test

import (
	"encoding/json/jsontext"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/protocol"
)

// Go 1.27 exports jsontext.AppendQuote, which does the same job as the
// hand-rolled escaper. Kept as the record of a rejected swap.
//
// In isolation the stdlib wins by about 26% on every shape below. In place it
// was worth 2.4% of the hook call (297 -> 290 ns), because the quotes it adds
// have to be memmoved away — this escaper writes the *body* of a string that
// is still being built. Against that it does not escape U+2028/U+2029, which
// encoding/json does, so adopting it would fork the response bytes from the
// encoder the differential test pins, and needs a fallback branch for the
// invalid-UTF-8 case it rejects. Not worth 7 ns; revisit if the fused emitter
// ever escapes more than one command per call.
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
