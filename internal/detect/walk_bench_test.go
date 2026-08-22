package detect

import (
	"encoding/json/jsontext"
	"fmt"
	"strings"
	"testing"

	"github.com/buger/jsonparser"

	"github.com/alex60217101990/terse/internal/protocol"
)

// bigArray is the shape the JSON summarizer actually meets: a tool that dumped
// a few hundred records.
func bigArray(rows int) []byte {
	var b strings.Builder
	b.WriteByte('[')
	for i := range rows {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"id":%d,"name":"item-%d","state":"running","cpu":%d.%d,`+
			`"mem":%d,"zone":"eu-central-1a","owner":"team-platform","ok":true}`,
			i, i, i%97, i%10, i*1024)
	}
	b.WriteByte(']')
	return []byte(b.String())
}

var arrayPayload = bigArray(500)

// BenchmarkWalkArray_Jsonparser is what the summarizer does today: a raw scan,
// values handed back as slices of the input, no validation.
func BenchmarkWalkArray_Jsonparser(b *testing.B) {
	b.SetBytes(int64(len(arrayPayload)))
	b.ReportAllocs()
	for b.Loop() {
		cells := 0
		_, err := jsonparser.ArrayEach(arrayPayload, func(row []byte, dt jsonparser.ValueType, _ int, _ error) {
			if dt != jsonparser.Object {
				return
			}
			_ = jsonparser.ObjectEach(row, func(_, val []byte, _ jsonparser.ValueType, _ int) error {
				cells += len(val)
				return nil
			})
		})
		if err != nil || cells == 0 {
			b.Fatal("miss")
		}
	}
}

// BenchmarkWalkArray_Jsontext is the same walk through the standard library.
func BenchmarkWalkArray_Jsontext(b *testing.B) {
	b.SetBytes(int64(len(arrayPayload)))
	b.ReportAllocs()
	for b.Loop() {
		cells := 0
		dec := jsontext.NewDecoder(strings.NewReader(string(arrayPayload)))
		if _, err := dec.ReadToken(); err != nil {
			b.Fatal(err)
		}
		for dec.PeekKind() != ']' {
			row, err := dec.ReadValue()
			if err != nil {
				b.Fatal(err)
			}
			if err := protocol.ScanObject(row, func(_, val []byte) error {
				cells += len(val)
				return nil
			}); err != nil {
				b.Fatal(err)
			}
		}
		if cells == 0 {
			b.Fatal("miss")
		}
	}
}
