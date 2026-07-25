package daemon

import (
	"strings"
	"testing"
)

// TestSTATS_ReturnsRuntimeSnapshot proves the STATS control word (peer of
// PING/QUIT in handleConn's switch) replies with a one-metric-per-line
// runtime snapshot, without ever reaching the hook dispatch pipeline.
func TestSTATS_ReturnsRuntimeSnapshot(t *testing.T) {
	sock := startTestDaemon(t)
	reply := rawRoundtrip(t, sock, []byte("STATS\n"))
	for _, want := range []string{"heap_alloc", "num_goroutine", "uptime"} {
		if !strings.Contains(string(reply), want) {
			t.Errorf("STATS reply missing %q: %s", want, reply)
		}
	}
}
