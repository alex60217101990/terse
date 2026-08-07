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
	"github.com/alex60217101990/terse/internal/tokens"
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

	record := func(action string, bytesOut, tokensOut, tokensIn int) {
		_ = analytics.Record(analytics.Event{
			TS:        time.Now().UnixNano(),
			SID:       inp.SessionID,
			Hook:      inp.ToolName, // canonical Claude tool name (Write/Edit/MultiEdit)
			Action:    action,
			BytesIn:   echoLen,
			BytesOut:  bytesOut,
			TokensIn:  tokensIn,
			TokensOut: tokensOut,
			DurNS:     time.Since(start).Nanoseconds(),
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
				if state := store.LoadSession(ContextKey(inp)); state != nil {
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
					store.SaveSession(ContextKey(inp), state)
					cached = true
				}
			}
		}
	}

	// Replace the echo with a compact cache-ref marker, but only when we cached
	// the file AND the marker actually costs the model fewer tokens than the
	// echo it replaces.
	//
	// The gate used to be len(marker) < EchoLen, and it was wrong twice over.
	// EchoLen is the raw tool_response JSON — for Edit that embeds the entire
	// pre-edit originalFile, so it is enormous, while what the model is shown
	// is one short rendered sentence. And bytes are not what the model pays: a
	// 32-char hex hash is byte-cheap and token-expensive, roughly one token per
	// two or three characters. Measured over 1251 real Write/Edit calls, the
	// old gate fired almost every time and made the output BIGGER — Edit −3.4%,
	// Write −9.4%.
	//
	// So the comparison is against Text(), which is what the model actually
	// sees, in tokens. When Text() is empty the rendered echo is not visible to
	// this process at all (the real Edit shape), and a marker cannot be shown
	// to be cheaper than something we cannot measure — so it passes through.
	// The cache priming above already happened either way; that, not the
	// marker, is this hook's purpose.
	echo := inp.ToolResponse.Text()
	if cached && echo != "" {
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
		if markerTok, echoTok := tokens.Count(marker), tokens.Count(echo); markerTok < echoTok {
			record("compressed", len(marker), markerTok, echoTok)
			return protocol.EncodeOutput(w, protocol.Replace(marker))
		}
	}
	echoTok := tokens.Count(echo)
	record("passthrough", echoLen, echoTok, echoTok)
	return protocol.EncodeOutput(w, protocol.Passthrough())
}
