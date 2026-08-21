package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Triple is one recorded tool call: the tool that ran, the input it was given,
// and the text result the model received back. It is the unit the replay
// harness feeds to the hook pipeline.
type Triple struct {
	Session string          `json:"session"`
	Tool    string          `json:"tool"`
	Result  string          `json:"result"`
	Input   json.RawMessage `json:"input"`
}

// Session is one transcript file's worth of triples, in transcript order.
//
// Order matters and so does the boundary: §ref dedup, the Read delta cache and
// the rerun-delta store are all stateful and all keyed within a session, so a
// replay that reorders triples or merges two sessions measures something that
// never happened.
type Session struct {
	Path    string   `json:"path"`
	Triples []Triple `json:"triples"`
}

// Fingerprint identifies the corpus a report was produced from. Comparing two
// reports taken over different corpora is meaningless, so --baseline refuses
// to do it; this is what it checks.
//
// Hash is SHA-256 over the sorted (relative path, size) pairs. Size rather
// than content: hashing a gigabyte archive on every run would dominate the
// replay, and a transcript that changes size is the only way a transcript
// changes at all (they are append-only).
type Fingerprint struct {
	Hash       string `json:"hash"`
	Files      int    `json:"files"`
	TotalBytes int64  `json:"total_bytes"`
}

// Corpus is the result of walking a transcript directory.
//
// Skipped/Filtered/Unpaired are carried, not discarded, because a corrupt or
// half-read archive must not produce a report that looks exactly like a clean
// one. A run whose Skipped count jumps is not a run whose savings improved.
type Corpus struct {
	Sessions    []Session
	Fingerprint Fingerprint
	// Skipped counts lines that looked like transcript entries but could not be
	// parsed as JSON.
	Skipped int
	// Filtered counts triples dropped because the recorded result already
	// carried qdf-hook markers.
	Filtered int
	// Unpaired counts tool_result blocks whose tool_use was not in the same
	// file (a truncated or resumed transcript).
	Unpaired int
}

// Triples returns the total number of triples across every session.
func (c *Corpus) Triples() int {
	n := 0
	for _, s := range c.Sessions {
		n += len(s.Triples)
	}
	return n
}

// scanWindow is how much of a result IsHookOutput inspects at each end.
const scanWindow = 4 << 10

// hookMarkers are the literal strings qdf-hook emits. Any result containing
// one was already compressed by an earlier run of the tool.
//
// TWO generations are listed. The § markers are current — measurement showed
// § costs exactly one token, the same as an ASCII sigil, so there was nothing
// to win by renaming them and they stayed. The block-dedup and path-prefix
// markers did move to ASCII, because ⟦ ⟧ cost three tokens each and §P§ cost
// three per folded line; their retired forms stay listed here because a
// transcript archive spans months, and therefore spans both.
//
// Precision matters more than brevity here in one direction only: a marker
// this list MISSES lets already-compressed output back into the corpus, where
// the pipeline compresses it a second time and manufactures a win that no real
// session ever saw. A marker that over-matches merely shrinks the corpus.
var hookMarkers = []string{
	// Current.
	"§ref:",
	"§unchanged:",
	"§unchanged-window:",
	"§delta:",
	"§rerun-delta§",
	"  ⨯",
	"[repeat",
	"[^=",

	// Retired: block dedup and path-prefix folding before the ASCII switch.
	"⟦↑ repeat",
	"§P=",

	// Generation-independent: the recovery footer and the structural summary
	// headers, whose text is not marker-dependent.
	"qdf-hook expand",
	"[expand ",
	"[TABLE ",
	"[grep: ",
	"[JSON OBJECT",
	"[go test PASS]",
	"[go test FAIL]",
	"[go bench ",
	"[git log ",
	"[qdf-hook SESSION RESTORE",
}

// IsHookOutput reports whether s already carries qdf-hook markers.
//
// It inspects the first and last scanWindow bytes rather than only the head:
// the "[full output: qdf-hook expand HASH]" recovery footer is appended at the
// END of a lossy summary, and a table or grep summary can easily run past 4 KB
// before reaching it. Scanning only the head would let exactly the lossiest —
// and therefore most misleading — outputs back into the corpus.
func IsHookOutput(s string) bool {
	if s == "" {
		return false
	}
	head := s
	if len(head) > scanWindow {
		head = head[:scanWindow]
	}
	var tail string
	if len(s) > scanWindow {
		tail = s[len(s)-scanWindow:]
	}
	for _, m := range hookMarkers {
		if strings.Contains(head, m) {
			return true
		}
		if tail != "" && strings.Contains(tail, m) {
			return true
		}
	}
	return false
}

// LoadSessions walks dir for *.jsonl transcripts and returns one Session per
// file, in path order, with triples in transcript order.
func LoadSessions(dir string) (Corpus, error) {
	var corpus Corpus

	paths, err := transcriptPaths(dir)
	if err != nil {
		return corpus, err
	}
	if len(paths) == 0 {
		return corpus, fmt.Errorf("no *.jsonl transcripts under %s", dir)
	}

	corpus.Fingerprint, err = fingerprint(dir, paths)
	if err != nil {
		return corpus, err
	}

	for _, p := range paths {
		sess, err := parseSession(p, &corpus)
		if err != nil {
			return corpus, fmt.Errorf("%s: %w", p, err)
		}
		if len(sess.Triples) == 0 {
			// A transcript with nothing to replay (all prose, or every result
			// already compressed) still counts toward the fingerprint, but an
			// empty session would only add noise to the report.
			continue
		}
		corpus.Sessions = append(corpus.Sessions, sess)
	}
	return corpus, nil
}

// transcriptPaths returns every *.jsonl file under dir, sorted.
//
// Walk errors on individual entries are tolerated (a real archive can contain
// a directory the user cannot read); an error opening the root is not, because
// that means the whole corpus is missing and a silent empty run would be far
// worse than a message.
func transcriptPaths(dir string) ([]string, error) {
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("corpus dir: %w", err)
	}
	var paths []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == dir {
				return err
			}
			return nil // unreadable entry: skip it, keep walking
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".jsonl") {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

// fingerprint hashes the sorted (relative path, size) pairs of the corpus.
func fingerprint(dir string, paths []string) (Fingerprint, error) {
	type entry struct {
		rel  string
		size int64
	}
	entries := make([]entry, 0, len(paths))
	var total int64
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return Fingerprint{}, fmt.Errorf("stat %s: %w", p, err)
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			rel = p
		}
		entries = append(entries, entry{filepath.ToSlash(rel), info.Size()})
		total += info.Size()
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].rel < entries[j].rel })

	h := sha256.New()
	for _, e := range entries {
		// NUL-separated: a path cannot contain a NUL, so a rename can never be
		// disguised as a size change or vice versa.
		fmt.Fprintf(h, "%s\x00%d\x00", e.rel, e.size)
	}
	return Fingerprint{
		Hash:       hex.EncodeToString(h.Sum(nil)),
		Files:      len(entries),
		TotalBytes: total,
	}, nil
}

// transcriptLine is the sliver of a Claude Code transcript entry this tool
// needs. Everything else on the line (uuid, timestamps, cwd, attachments) is
// left undecoded — it is both irrelevant and, in a real archive, private.
type transcriptLine struct {
	Type    string `json:"type"`
	Message struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// contentBlock is one element of a message's content array. tool_use and
// tool_result are different shapes sharing one array, so both sets of fields
// live here and Type says which are populated.
type contentBlock struct {
	Type string `json:"type"`

	// tool_use
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`

	// tool_result
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
}

type pendingCall struct {
	name  string
	input json.RawMessage
}

// parseSession reads one transcript file and pairs each tool_result with the
// tool_use that produced it, via tool_use_id.
//
// Pairing is what makes a triple replayable: the hook pipeline routes on the
// tool NAME and several handlers (Read above all) parse the tool INPUT, so a
// result without its call cannot be dispatched the way it originally was.
func parseSession(path string, c *Corpus) (Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer f.Close()

	sess := Session{Path: path}
	name := filepath.Base(path)
	pending := make(map[string]pendingCall)
	r := bufio.NewReaderSize(f, 1<<20)

	for {
		line, err := readLine(r)
		if len(line) > 0 {
			c.consumeLine(&sess, name, pending, line)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return sess, nil
			}
			return sess, err
		}
	}
}

// consumeLine folds one transcript line into sess.
func (c *Corpus) consumeLine(sess *Session, name string, pending map[string]pendingCall, line []byte) {
	if len(strings.TrimSpace(string(line))) == 0 {
		return
	}
	var tl transcriptLine
	if json.Unmarshal(line, &tl) != nil {
		c.Skipped++
		return
	}
	// Meta entries (attachment, file-history-snapshot, queue-operation, ...)
	// carry no message content. They are not malformed, so they are not
	// skipped — counting them would drown the signal Skipped exists to give.
	raw := trimJSONSpace(tl.Message.Content)
	if len(raw) == 0 || raw[0] != '[' {
		return
	}
	var blocks []contentBlock
	if json.Unmarshal(raw, &blocks) != nil {
		c.Skipped++
		return
	}
	for _, b := range blocks {
		switch b.Type {
		case "tool_use":
			if b.ID != "" {
				pending[b.ID] = pendingCall{name: b.Name, input: b.Input}
			}
		case "tool_result":
			call, ok := pending[b.ToolUseID]
			if !ok {
				c.Unpaired++
				continue
			}
			delete(pending, b.ToolUseID)
			text := resultText(b.Content)
			if text == "" {
				continue
			}
			if IsHookOutput(text) {
				c.Filtered++
				continue
			}
			sess.Triples = append(sess.Triples, Triple{
				Session: name,
				Tool:    call.name,
				Input:   call.input,
				Result:  text,
			})
		}
	}
}

// resultText extracts the text of a tool_result's content, which is either a
// plain JSON string or an array of content blocks (what MCP tools and a
// finished Agent produce). Non-text blocks contribute nothing.
func resultText(raw json.RawMessage) string {
	raw = trimJSONSpace(raw)
	if len(raw) == 0 {
		return ""
	}
	switch raw[0] {
	case '"':
		var s string
		if json.Unmarshal(raw, &s) != nil {
			return ""
		}
		return s
	case '[':
		var blocks []struct {
			Text string `json:"text"`
		}
		if json.Unmarshal(raw, &blocks) != nil {
			return ""
		}
		var sb strings.Builder
		for _, b := range blocks {
			if b.Text == "" {
				continue
			}
			if sb.Len() > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(b.Text)
		}
		return sb.String()
	default:
		return ""
	}
}

// readLine returns one newline-terminated line with NO length cap.
//
// bufio.Scanner cannot be used here: its token limit is a hard error that
// abandons the rest of the file, and a single transcript line legitimately
// runs to many megabytes when a tool dumped a large result. Losing the tail of
// a transcript because one line was long would silently shrink the corpus.
//
// The returned slice may alias the reader's buffer and is only valid until the
// next call.
func readLine(r *bufio.Reader) ([]byte, error) {
	var buf []byte
	for {
		chunk, err := r.ReadSlice('\n')
		if errors.Is(err, bufio.ErrBufferFull) {
			buf = append(buf, chunk...)
			continue
		}
		if buf == nil {
			return chunk, err
		}
		return append(buf, chunk...), err
	}
}

// trimJSONSpace strips the whitespace JSON allows before a value, so the first
// byte can be used to tell a string from an array.
func trimJSONSpace(b []byte) []byte { return bytes.TrimLeft(b, " \t\r\n") }
