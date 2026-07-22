package detect

import (
	"strings"
)

// IsGitLogOutput returns true if s looks like `git log --oneline` output.
// Heuristic: at least 2 non-empty lines each starting with a 7–40 char hex hash.
func IsGitLogOutput(s string) bool {
	lines := strings.SplitN(strings.TrimSpace(s), "\n", 6)
	if len(lines) < 2 {
		return false
	}
	matched := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		hash := fields[0]
		if len(hash) < 7 || len(hash) > 40 {
			return false
		}
		isHex := true
		for _, c := range hash {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				isHex = false
				break
			}
		}
		if isHex {
			matched++
		}
	}
	return matched >= 2
}

// SummarizeGitLog returns a compact table of commits.
// Keeps hash (short), and truncates long commit messages to 72 characters.
func SummarizeGitLog(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	var body strings.Builder
	commits := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		hash := fields[0]
		if len(hash) > 8 {
			hash = hash[:8]
		}
		msg := strings.Join(fields[1:], " ")
		// Truncate long messages to keep output compact. Cut on a rune
		// boundary so multi-byte characters are never split.
		if r := []rune(msg); len(r) > 72 {
			msg = string(r[:69]) + "..."
		}
		body.WriteString(hash)
		body.WriteString(" ")
		body.WriteString(msg)
		body.WriteString("\n")
		commits++
	}

	var sb strings.Builder
	sb.WriteString("[git log — ")
	sb.WriteString(itoa(commits))
	sb.WriteString(" commits]\n")
	sb.WriteString(body.String())
	return sb.String()
}
