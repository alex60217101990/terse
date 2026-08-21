package cache

import (
	"os"
	"path/filepath"
)

// CaptureDir is where a wrapped command's full output is parked so a capped
// view stays recoverable without re-running the command. Peer of RefsDir.
func CaptureDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".qdf-hook", "captures")
}

// CapturePath is the capture file for id. filepath.Base pins the result inside
// CaptureDir: the id reaches a shell command line, so traversal is not
// hypothetical.
func CapturePath(id string) string {
	return filepath.Join(CaptureDir(), filepath.Base(id)+".out")
}
