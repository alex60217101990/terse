package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alex60217101990/terse/internal/tokens"
)

// Throwaway probe: what the cap is worth in AMPLIFIED tokens, the only unit the
// bill is paid in. A tool result is re-sent with every later turn of its
// session, so a byte cut once is cut once per remaining turn.
//
// QDF_CORPUS=<dir> go test -run TestValueProbe -v -timeout 30m.
func TestValueProbe(t *testing.T) {
	dir := os.Getenv("QDF_CORPUS")
	if dir == "" {
		t.Skip("set QDF_CORPUS")
	}
	const capAt = 1600 // bytes, the shipped capBytes
	const notice = "\n...    23893 bytes elided, full output: qdf-hook expand ac1f1163110cc96ecd50aba387e472bc\n"
	noticeTok := tokens.Count(notice)

	type call struct {
		cmd      string
		body     string
		turnsIdx int // index of the turn this result landed on
	}

	var bashAmp, savedAmp, capped, total, overCap int
	var perCallSaved []int

	walk := func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".jsonl" {
			return nil //nolint:nilerr // an unreadable corpus file is not a failure
		}
		f, err := os.Open(path)
		if err != nil {
			return nil //nolint:nilerr // same
		}
		defer f.Close()

		// One pass to collect Bash results in order plus the turn count, a
		// second to price them: amplification is "turns after this one".
		var calls []call
		turns := 0
		cmdByID := map[string]string{}
		dec := json.NewDecoder(f)
		for {
			var rec map[string]any
			if dec.Decode(&rec) != nil {
				break
			}
			msg, _ := rec["message"].(map[string]any)
			if msg == nil {
				continue
			}
			if rec["type"] == "assistant" {
				turns++
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
					var body strings.Builder
					switch v := m["content"].(type) {
					case string:
						body.WriteString(v)
					case []any:
						for _, e := range v {
							if em, _ := e.(map[string]any); em != nil {
								if s, _ := em["text"].(string); s != "" {
									body.WriteString(s)
								}
							}
						}
					}
					calls = append(calls, call{cmd, body.String(), turns})
				}
			}
		}
		for _, c := range calls {
			amp := max(turns-c.turnsIdx, 1)
			full := tokens.Count(c.body)
			bashAmp += full * amp
			total++
			if len(c.body) <= capAt || !cappable(c.cmd) {
				continue
			}
			overCap++
			half := capAt / 2
			kept := tokens.Count(c.body[:half]) + tokens.Count(c.body[len(c.body)-half:]) + noticeTok
			if kept >= full {
				continue
			}
			capped++
			saved := (full - kept) * amp
			savedAmp += saved
			perCallSaved = append(perCallSaved, saved)
		}
		return nil
	}
	if err := filepath.WalkDir(dir, walk); err != nil {
		t.Fatal(err)
	}

	// Anchors from the spec's bill decomposition: Bash output is 8.51% of the
	// bill, and one recovery turn costs 217,263 amplified tokens.
	const bashShareOfBill = 8.51
	const recoveryTurn = 217263
	share := float64(savedAmp) / float64(bashAmp)
	var avg int
	if len(perCallSaved) > 0 {
		avg = savedAmp / len(perCallSaved)
	}
	t.Logf("bash results: %d, over cap and cappable: %d (%.1f%%)", total, overCap, 100*float64(overCap)/float64(total))
	t.Logf("bash amplified tokens: %d", bashAmp)
	t.Logf("saved amplified tokens: %d (%.1f%% of the Bash pool, %.2f%% of the bill)",
		savedAmp, 100*share, share*bashShareOfBill)
	t.Logf("avg saved per capped call: %d amplified tokens", avg)
	t.Logf("break-even recovery rate: %.1f%% (one recovery turn = %d)",
		100*float64(avg)/float64(recoveryTurn), recoveryTurn)
}
