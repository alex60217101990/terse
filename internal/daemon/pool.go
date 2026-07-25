package daemon

import (
	"bytes"
	"io"
	"sync"
)

// reqBufInit sizes pooled request buffers to hold the large majority of tool
// outputs without a grow-realloc. Larger requests still work — the buffer
// grows for them; it just isn't returned to the pool if it grew past
// reqBufCap (see readRequest), so one giant payload can't pin megabytes in
// the pool for the daemon's whole life.
const (
	reqBufInit = 16 << 10 // 16 KiB initial capacity
	reqBufCap  = 1 << 20  // 1 MiB: don't pool buffers grown past this
)

var reqBufPool = sync.Pool{
	New: func() any {
		b := new(bytes.Buffer)
		b.Grow(reqBufInit)
		return b
	},
}

// readRequest reads all of r into a pooled buffer and returns it. The caller
// must return the buffer with putRequest when done (typically `defer
// putRequest(buf)`). It replaces a plain io.ReadAll so the daemon reuses one
// buffer across requests instead of allocating (and growing) a fresh one
// every time — cutting steady-state allocations and GC pressure under
// sustained load.
//
// The pair is deliberately two plain functions rather than returning a
// release closure: a captured closure escapes to the heap (one alloc per
// call), whereas `defer putRequest(buf)` running once per handler is
// open-coded by the compiler with no allocation.
//
// SAFETY: the returned buffer's bytes alias pooled storage, so the caller
// must not retain them past putRequest. This is sound in handleConn because
// protocol.DecodeInput uses encoding/json, which copies every string out of
// the input into fresh Go strings — nothing the store retains aliases the
// request bytes. Switching to a zero-copy JSON decoder would break this
// invariant and this pooling with it.
func readRequest(r io.Reader) *bytes.Buffer {
	buf := reqBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	_, _ = buf.ReadFrom(r)
	return buf
}

// putRequest returns a buffer from readRequest to the pool, dropping it if it
// grew past reqBufCap so one giant payload can't pin memory for the daemon's
// whole life.
func putRequest(buf *bytes.Buffer) {
	if buf.Cap() <= reqBufCap {
		reqBufPool.Put(buf)
	}
}
