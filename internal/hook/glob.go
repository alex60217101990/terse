package hook

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/protocol"
)

// HandleGlob compresses Glob tool output from a flat file list to a compact directory tree.
func HandleGlob(r io.Reader, w io.Writer) error {
	start := time.Now()
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	if inp.ToolResponse == nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	content := inp.ToolResponse.Content
	// Short content: passthrough — no benefit from compression.
	if len(content) <= 256 {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	tree := buildGlobTree(content)
	// Only replace if the tree is actually shorter than the original.
	if len(tree) >= len(content) {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	_ = analytics.Record(analytics.Event{
		TS:      time.Now().UnixNano(),
		SID:     inp.SessionID,
		Hook:    "glob",
		Action:  "tree",
		BytesIn: len(content),
		BytesOut: len(tree),
		DurNS:   time.Since(start).Nanoseconds(),
	})
	return protocol.EncodeOutput(w, protocol.Replace(tree))
}

// buildGlobTree converts a newline-separated list of file paths into a compact
// directory-tree summary grouped by the top two path components.
func buildGlobTree(content string) string {
	lines := strings.Split(strings.TrimSpace(content), "\n")

	type dirGroup struct {
		dir   string
		files []string
	}
	groups := make(map[string]*dirGroup)
	var order []string
	total := 0

	for _, path := range lines {
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
