package hook

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alex60217101990/terse/internal/analytics"
	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/hookcore"
	"github.com/alex60217101990/terse/internal/protocol"
)

// HandleWrite compresses Write/Edit/MultiEdit tool responses.
// Claude just wrote this content — no need to see the full file (Edit's echo
// even embeds the entire pre-edit originalFile). Also caches the written file
// for delta tracking on the next Read.
func HandleWrite(r io.Reader, w io.Writer) error {
	inp, err := protocol.DecodeInput(r)
	if err != nil {
		return fmt.Errorf("DecodeInput: %w", err)
	}
	return handleWrite(hookcore.NewDiskStore(), inp, w)
}

// handleWrite is the Write/Edit logic over an already-decoded input.
//
// The Write/Edit/MultiEdit tool_response carries no plain-text echo in a field
// this code reads (Edit's is {filePath, oldString, newString, originalFile,
// structuredPatch, ...}) — so echo size is taken from the raw tool_response
// length (EchoLen), and the cached content is the ACTUAL file read from disk,
// never the tool echo (which is a diff/original, not the post-edit file).
func handleWrite(store hookcore.StateStore, inp *protocol.HookInput, w io.Writer) error {
	start := time.Now()
	if inp.ToolResponse == nil {
		return protocol.EncodeOutput(w, protocol.Passthrough())
	}

	// Parse file path from tool_input (Write and Edit both have file_path).
	var ti protocol.ReadInput // reuse: same {file_path} shape
	_ = json.Unmarshal(inp.ToolInput, &ti)

	echoLen := inp.ToolResponse.EchoLen()

	record := func(action string, bytesOut int) {
		_ = analytics.Record(analytics.Event{
			TS:       time.Now().UnixNano(),
			SID:      inp.SessionID,
			Hook:     inp.ToolName, // canonical Claude tool name (Write/Edit/MultiEdit)
			Action:   action,
			BytesIn:  echoLen,
			BytesOut: bytesOut,
			DurNS:    time.Since(start).Nanoseconds(),
		})
	}

	// Prime the delta cache with the actual post-edit file bytes, read from a
	// single open fd (os.Open + f.Stat gives content size + ModTime in one
	// stat). This is the core purpose — it makes the NEXT Read of this path
	// return §delta/§unchanged — and runs whenever the file is readable and we
	// have a session, independent of echo size.
	var (
		hashHex   string
		lineCount int
		cached    bool
	)
	if ti.FilePath != "" && inp.SessionID != "" {
		if f, err := os.Open(ti.FilePath); err == nil {
			var modTime, ctimeNS int64
			if fi, e := f.Stat(); e == nil {
				modTime = fi.ModTime().UnixNano()
				ctimeNS = statCtimeNS(fi)
			}
			fileBytes, rerr := io.ReadAll(f)
			_ = f.Close()
			if rerr == nil {
				hash := sha256.Sum256(fileBytes)
				hashHex = cache.ShortHex(hash[:8])
				lineCount = bytes.Count(fileBytes, []byte("\n"))
				if state := store.LoadSession(inp.SessionID); state != nil {
					state.Turn++
					state.Files[ti.FilePath] = cache.FileEntry{
						Hash:       hash,
						Turn:       state.Turn,
						Content:    fileBytes,
						ModTime:    modTime,
						CtimeNS:    ctimeNS,
						LastReadAt: time.Now().Unix(),
						ReadCount:  1,
					}
					store.SaveSession(inp.SessionID, state)
					cached = true
				}
			}
		}
	}

	// Replace the (often large) echo with a compact cache-ref marker, but only
	// when we cached the file and the marker is actually shorter than the echo
	// it replaces (never-worse). Otherwise pass the response through unchanged.
	if cached {
		var mb strings.Builder
		mb.Grow(len(hashHex) + len(ti.FilePath) + 48)
		mb.WriteString("[WRITE §ref:")
		mb.WriteString(hashHex)
		mb.WriteString("§ ")
		mb.WriteString(ti.FilePath)
		mb.WriteString(" — ")
		mb.WriteString(strconv.Itoa(lineCount))
		mb.WriteString(" lines, cached for delta tracking]")
		marker := mb.String()
		if len(marker) < echoLen {
			record("compressed", len(marker))
			return protocol.EncodeOutput(w, protocol.Replace(marker))
		}
	}
	record("passthrough", echoLen)
	return protocol.EncodeOutput(w, protocol.Passthrough())
}
