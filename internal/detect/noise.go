package detect

import "strings"

// noisePrefixes are line prefixes that carry no information for the model —
// pure shell/build/tooling chatter. Matched with a cheap byte-level prefix
// check (no regexp).
var noisePrefixes = []string{
	"zsh: command not found: _", // gvm _encode/_decode shim spam
	"go: downloading ",
	"go: finding ",
	"go: extracting ",
	"npm notice",
	"npm warn ",
	"npm fund",
	"npm WARN",
	"Cloning into ",
	"Download complete",
	"Pull complete",
	"Pulling from ",
	"Verifying Checksum",
	"Already exists",
}

// StripNoise drops known junk lines (shell/build chatter) from command output.
// It returns the input unchanged — the same string, no copy — when nothing
// matched, so the common (clean) case is allocation-free and callers can gate
// on len/identity. Kept lines are appended as zero-copy slices of the original.
func StripNoise(content string) string {
	// Fast reject: if none of the cheapest discriminators appear anywhere, skip.
	if !strings.Contains(content, "command not found: _") &&
		!strings.Contains(content, "go: ") &&
		!strings.Contains(content, "npm ") &&
		!strings.Contains(content, "complete") &&
		!strings.Contains(content, "Pulling") {
		return content
	}

	var b strings.Builder
	b.Grow(len(content))
	dropped := false
	start := 0
	for start <= len(content) {
		nl := strings.IndexByte(content[start:], '\n')
		var line string
		var end int
		if nl < 0 {
			line = content[start:]
			end = len(content)
		} else {
			line = content[start : start+nl]
			end = start + nl + 1
		}
		switch {
		case isNoise(line):
			dropped = true
		case nl < 0:
			b.WriteString(line)
		default:
			b.WriteString(content[start:end]) // line incl. its '\n', zero-copy slice
		}
		if nl < 0 {
			break
		}
		start = end
	}
	if !dropped {
		return content // nothing removed — avoid returning a rebuilt copy
	}
	return b.String()
}

// isNoise reports whether a line (without its trailing newline) is junk.
func isNoise(line string) bool {
	// Trim leading ASCII space/tab cheaply without allocating.
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	trimmed := line[i:]
	for _, p := range noisePrefixes {
		if strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}
