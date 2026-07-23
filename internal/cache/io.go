package cache

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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
