package hook

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alex60217101990/qdf-hook/internal/hookcore"
)

// HandleGlob is retained for backward compatibility; Glob is handled by the
// generic pipeline via Dispatch (buildGlobTree is its tool-specific step).
func HandleGlob(r io.Reader, w io.Writer) error { return Dispatch(hookcore.NewDiskStore(), r, w) }

// buildGlobTree converts a newline-separated list of file paths into a compact
// directory-tree summary grouped by the top two path components.
func buildGlobTree(content string) string {
	type dirGroup struct {
		dir   string
		files []string
	}
	groups := make(map[string]*dirGroup)
	var order []string
	total := 0

	// SplitSeq: single forward pass, no []string materialized.
	for path := range strings.SplitSeq(strings.TrimSpace(content), "\n") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		dir := topGlobDir(path)
		if _, ok := groups[dir]; !ok {
			groups[dir] = &dirGroup{dir: dir}
			order = append(order, dir)
		}
		groups[dir].files = append(groups[dir].files, filepath.Base(path))
		total++
	}
	sort.Strings(order)

	var sb strings.Builder
	for _, dir := range order {
		g := groups[dir]
		sort.Strings(g.files)
		if len(g.files) <= 6 {
			fmt.Fprintf(&sb, "%-30s %d files (%s)\n",
				dir+"/", len(g.files), strings.Join(g.files, " "))
		} else {
			fmt.Fprintf(&sb, "%-30s %d files\n", dir+"/", len(g.files))
		}
	}
	fmt.Fprintf(&sb, "[%d total files, %d dirs]", total, len(groups))
	return sb.String()
}

// topGlobDir returns up to two slash-separated path components as the group key.
func topGlobDir(path string) string {
	parts := strings.SplitN(filepath.ToSlash(path), "/", 3)
	switch len(parts) {
	case 1:
		return "."
	case 2:
		return parts[0]
	default:
		return parts[0] + "/" + parts[1]
	}
}
