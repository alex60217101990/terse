package protocol

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"sync"

	"github.com/alex60217101990/terse/internal/bytesconv"
)

// scanner is a reusable jsontext decoder plus the reader it reads through.
// Both are pooled: the hook decodes one request per tool call, and neither the
// decoder's buffer nor the reader carries anything between requests once Reset.
type scanner struct {
	dec *jsontext.Decoder
	rd  *bytes.Reader
}

var scanPool = sync.Pool{New: func() any {
	rd := bytes.NewReader(nil)
	return &scanner{
		// Claude Code's own encoder decides what reaches us, and the v1 decoder
		// this replaced accepted both of these: a duplicate key (last wins) and
		// a byte that is not valid UTF-8. Rejecting them here would turn a
		// decodable request into a hook error.
		dec: jsontext.NewDecoder(rd,
			jsontext.AllowDuplicateNames(true),
			jsontext.AllowInvalidUTF8(true)),
		rd: rd,
	}
}}

var errNotObject = errors.New("protocol: request is not a JSON object")

// scanObject walks the members of the top-level JSON object in data, calling fn
// with each key (quotes stripped, escapes left as-is) and the member's raw
// value.
//
// Both slices ALIAS data — jsontext hands back values that live in the
// decoder's own buffer and are invalidated by the next read, so the offsets are
// used to point back into the caller's bytes instead. A caller that keeps a
// value beyond data's lifetime must copy it.
func scanObject(data []byte, fn func(key, val []byte) error) error {
	s, _ := scanPool.Get().(*scanner)
	if s == nil {
		rd := bytes.NewReader(nil)
		s = &scanner{dec: jsontext.NewDecoder(rd), rd: rd}
	}
	defer scanPool.Put(s)

	s.rd.Reset(data)
	s.dec.Reset(s.rd)

	tok, err := s.dec.ReadToken()
	if err != nil {
		return err
	}
	if tok.Kind() != '{' {
		return errNotObject
	}
	for s.dec.PeekKind() != '}' {
		key, err := s.dec.ReadValue()
		if err != nil {
			return err
		}
		k := alias(data, s.dec.InputOffset(), len(key))
		val, err := s.dec.ReadValue()
		if err != nil {
			return err
		}
		v := alias(data, s.dec.InputOffset(), len(val))
		if len(k) >= 2 {
			k = k[1 : len(k)-1] // strip the key's quotes
		}
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

// alias returns the n bytes of data that end at the decoder's current offset.
func alias(data []byte, end int64, n int) []byte {
	hi := int(end)
	lo := hi - n
	if lo < 0 || hi > len(data) {
		return nil
	}
	return data[lo:hi]
}

// joinStrings materializes raw JSON string values into one allocation and
// slices it, so a request pays for one string however many fields it carries.
// An absent field stays empty; a value that is not a JSON string is dropped,
// matching what a struct decode would have left behind.
func joinStrings(out *[numStrFields]string, raw *[numStrFields][]byte) {
	total := 0
	for _, r := range raw {
		if len(r) >= 2 && r[0] == '"' {
			total += len(r) - 2
		}
	}
	if total == 0 {
		return
	}
	buf := make([]byte, 0, total)
	var bounds [numStrFields][2]int
	for i, r := range raw {
		if len(r) < 2 || r[0] != '"' {
			continue
		}
		start := len(buf)
		body := r[1 : len(r)-1]
		if bytes.IndexByte(body, '\\') < 0 {
			buf = append(buf, body...)
		} else if unquoted, err := jsontext.AppendUnquote(buf, r); err == nil {
			buf = unquoted
		}
		bounds[i] = [2]int{start, len(buf)}
	}
	packed := bytesconv.B2S(buf)
	for i, b := range bounds {
		if b[1] > b[0] {
			out[i] = packed[b[0]:b[1]]
		}
	}
}
