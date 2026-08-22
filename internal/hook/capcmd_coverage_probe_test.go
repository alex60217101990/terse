package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Throwaway probe: of the bytes real Bash calls put into the transcript, how
// many sit behind each capping disqualifier. Run explicitly with
// QDF_CORPUS=<dir> go test -run TestCoverageProbe -v.
func TestCoverageProbe(t *testing.T) {
	dir := os.Getenv("QDF_CORPUS")
	if dir == "" {
		t.Skip("set QDF_CORPUS")
	}

	classify := func(cmd string) string {
		if cmd == "" {
			return "empty"
		}
		if cappable(cmd) {
			return "cappable"
		}
		onlyMerge := true // every redirect seen so far is a harmless 2>&1
		for i := 0; i < len(cmd); i++ {
			switch c := cmd[i]; c {
			case '\\':
				i++
			case '|':
				if i+1 < len(cmd) && cmd[i+1] == '|' {
					i++
					continue
				}
				return "pipe"
			case '&':
				if i+1 < len(cmd) && cmd[i+1] == '&' {
					i++
					continue
				}
				return "background"
			case '>':
				if i >= 1 && cmd[i-1] == '2' && i+2 < len(cmd) && cmd[i+1] == '&' && cmd[i+2] == '1' {
					i += 2
					continue
				}
				onlyMerge = false
			case '<':
				onlyMerge = false
			case '`':
				return "substitution"
			case '$':
				if i+1 < len(cmd) && cmd[i+1] == '(' {
					return "substitution"
				}
			}
		}
		if !onlyMerge {
			return "redirect"
		}
		return "cappable" // includes commands whose only redirect is 2>&1
	}

	type stat struct{ n, bytes, over, nOver, maxB int }
	var cmdLens []int
	byClass := map[string]*stat{}
	add := func(class string, n int) {
		s := byClass[class]
		if s == nil {
			s = &stat{}
			byClass[class] = s
		}
		s.n++
		s.bytes += n
		if n > s.maxB {
			s.maxB = n
		}
		if n > 1600 {
			s.over += n - 1600
			s.nOver++
		}
	}

	resultLen := func(m map[string]any) int {
		c := m["content"]
		switch v := c.(type) {
		case string:
			return len(v)
		case []any:
			n := 0
			for _, e := range v {
				if em, _ := e.(map[string]any); em != nil {
					if s, _ := em["text"].(string); s != "" {
						n += len(s)
					}
				}
			}
			return n
		}
		return 0
	}

	walk := func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".jsonl" {
			return nil //nolint:nilerr // an unreadable corpus file is not a failure
		}
		f, err := os.Open(path)
		if err != nil {
			return nil //nolint:nilerr // same
		}
		defer f.Close()
		cmdByID := map[string]string{}
		dec := json.NewDecoder(f)
		for {
			var rec map[string]any
			if dec.Decode(&rec) != nil {
				return nil
			}
			msg, _ := rec["message"].(map[string]any)
			if msg == nil {
				continue
			}
			content, _ := msg["content"].([]any)
			for _, it := range content {
				m, _ := it.(map[string]any)
				if m == nil {
					continue
				}
				switch m["type"] {
				case "tool_use":
					if m["name"] != "Bash" {
						continue
					}
					in, _ := m["input"].(map[string]any)
					cmd, _ := in["command"].(string)
					id, _ := m["id"].(string)
					cmdByID[id] = cmd
				case "tool_result":
					id, _ := m["tool_use_id"].(string)
					cmd, seen := cmdByID[id]
					if !seen {
						continue
					}
					delete(cmdByID, id)
					add(classify(cmd), resultLen(m))
					if cappable(cmd) {
						cmdLens = append(cmdLens, len(cmd))
					}
				}
			}
		}
	}
	if err := filepath.WalkDir(dir, walk); err != nil {
		t.Fatal(err)
	}

	var totalN, totalB, totalOver int
	keys := make([]string, 0, len(byClass))
	for k, v := range byClass {
		keys = append(keys, k)
		totalN += v.n
		totalB += v.bytes
		totalOver += v.over
	}
	sort.Slice(keys, func(i, j int) bool { return byClass[keys[i]].bytes > byClass[keys[j]].bytes })
	sort.Ints(cmdLens)
	if n := len(cmdLens); n > 0 {
		sum := 0
		for _, v := range cmdLens {
			sum += v
		}
		t.Logf("cappable command length: n=%d mean=%d p50=%d p90=%d p99=%d max=%d",
			n, sum/n, cmdLens[n/2], cmdLens[n*90/100], cmdLens[n*99/100], cmdLens[n-1])
	}
	t.Logf("paired Bash results: %d, output %d bytes, over-cap %d bytes", totalN, totalB, totalOver)
	for _, k := range keys {
		v := byClass[k]
		t.Logf("  %-13s calls %6d (%4.1f%%)  over-cap calls %5d (%4.1f%% of class)  bytes %9d (%4.1f%%)  over-cap %9d (%4.1f%%)  max %8d",
			k, v.n, 100*float64(v.n)/float64(totalN),
			v.nOver, 100*float64(v.nOver)/float64(v.n),
			v.bytes, 100*float64(v.bytes)/float64(totalB),
			v.over, 100*float64(v.over)/float64(totalOver), v.maxB)
	}
	_ = strings.TrimSpace
}
