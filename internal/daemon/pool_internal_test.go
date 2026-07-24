package daemon

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// microPayload is a ~12 KiB request, the size class that made io.ReadAll grow
// through several reallocs in the baseline roundtrip benchmark.
var microPayload = []byte(strings.Repeat("bench log line here\n", 640))

// BenchmarkReadPath_ReadAll and BenchmarkReadPath_Pooled A/B the read buffer
// lever in isolation, free of socket/client noise: the only difference is
// io.ReadAll (fresh growing buffer each call) vs readRequest (pooled, reused).
// The pooled variant should report fewer allocs/op and lower B/op.
func BenchmarkReadPath_ReadAll(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = io.ReadAll(bytes.NewReader(microPayload))
	}
}

func BenchmarkReadPath_Pooled(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		buf := readRequest(bytes.NewReader(microPayload))
		_ = buf.Bytes()
		putRequest(buf)
	}
}
