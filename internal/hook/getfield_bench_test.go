package hook

import (
	"testing"

	"github.com/buger/jsonparser"

	"github.com/alex60217101990/terse/internal/protocol"
)

var toolInput = []byte(`{"command":"GOWORK=off go test ./internal/hook/ -run TestSomething",` +
	`"description":"Run the hook tests","timeout":120000}`)

// One field out of a small object is the shape both the capping hook and the
// summarizers need. jsonparser is a dependency; jsontext ships with the
// toolchain — this is the price comparison.
func BenchmarkGetField_Jsonparser(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		v, typ, _, err := jsonparser.Get(toolInput, "command")
		if err != nil || typ != jsonparser.String || len(v) == 0 {
			b.Fatal("miss")
		}
	}
}

func BenchmarkGetField_Jsontext(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		v, ok := jsonField(toolInput, "command")
		if !ok || len(v) == 0 {
			b.Fatal("miss")
		}
	}
}

// jsonField is the candidate replacement: walk the object and stop at the
// wanted key, returning its raw value as a slice of data.
func jsonField(data []byte, want string) ([]byte, bool) {
	var found []byte
	err := protocol.ScanObject(data, func(key, val []byte) error {
		if string(key) == want {
			found = val
			return protocol.ErrStopScan
		}
		return nil
	})
	return found, err == nil && found != nil
}
