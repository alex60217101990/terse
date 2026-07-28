package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	qdf "github.com/alex60217101990/qdf"
)

// writeFileLazy writes data to path, creating the parent directory only if the
// write fails because it doesn't exist. This avoids an os.MkdirAll (a stat, and
// often a mkdir attempt) on every call — the directory almost always already
// exists after the first write of a session.
func writeFileLazy(path string, data []byte) error {
	err := os.WriteFile(path, data, 0o600)
	if errors.Is(err, fs.ErrNotExist) {
		if mkErr := os.MkdirAll(filepath.Dir(path), 0o700); mkErr != nil {
			return mkErr
		}
		err = os.WriteFile(path, data, 0o600)
	}
	return err
}

// blobEncPool holds the []byte scratch buffers the content-addressed blob
// writers (RefPut, LastOutputPut) qdf-encode into, mirroring store.go's
// savePool. Without it each writer fed qdf.Marshal a nil dst, forcing qdf's
// geometric grow-chain to reallocate from scratch on every call; pooling means
// the backing array grows on the first few writes and is reused thereafter.
var blobEncPool = sync.Pool{New: func() any { b := make([]byte, 0, 4096); return &b }}

// marshalBlobPooled qdf-encodes v with OptBalanced into a pooled buffer and
// writes it to path. The buffer is returned to blobEncPool only after
// writeFileLazy returns — os.WriteFile copies the bytes synchronously, so this
// is race-safe against a concurrent Get that would overwrite the backing array.
// A buffer that grew past MaxPooledBufSize is dropped rather than pinned (same
// rationale as savePool). Returns the encode/write error for the caller to
// handle (both blob stores are rebuildable caches and ignore it).
func marshalBlobPooled(path string, v any) error {
	bufPtr := blobEncPool.Get().(*[]byte)
	data, err := qdf.AppendMarshal((*bufPtr)[:0], v, qdf.OptBalanced)
	if err != nil {
		if cap(data) <= MaxPooledBufSize {
			*bufPtr = data
			blobEncPool.Put(bufPtr)
		}
		return err
	}
	werr := writeFileLazy(path, data)
	if cap(data) <= MaxPooledBufSize {
		*bufPtr = data
		blobEncPool.Put(bufPtr)
	}
	return werr
}
