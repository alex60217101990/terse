package main

import (
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alex60217101990/terse/internal/cache"
	"github.com/alex60217101990/terse/internal/daemon"
)

// pingTimeout bounds how long `daemon --ping` waits for a reply before
// treating the daemon as down — generous enough for a live daemon under
// normal load, short enough for a container HEALTHCHECK to fail fast.
const pingTimeout = 2 * time.Second

// isLoopbackAddr reports whether addr is a host:port pair that only ever
// binds to the local machine. --pprof exposes runtime internals (heap
// dumps, goroutine stacks) with no auth, so a non-loopback bind would be a
// silent, undocumented exposure — this is checked and rejected before the
// server ever starts (see cmdDaemon's --serve path).
func isLoopbackAddr(addr string) bool {
	return strings.HasPrefix(addr, "127.0.0.1:") ||
		strings.HasPrefix(addr, "localhost:") ||
		strings.HasPrefix(addr, "[::1]:")
}

// daemonIdleTimeout is how long qdf-hookd sits idle (no connections) before
// it exits on its own. Long enough to survive gaps between hook calls within
// a session, short enough not to linger forever after the session ends.
const daemonIdleTimeout = 30 * time.Minute

// restoreDaemonRuntime undoes the one-shot-CLI tuning that main.init() applies
// (GOMAXPROCS(1) + SetGCPercent(-1)). The daemon is long-lived and concurrent:
// it MUST garbage-collect (or it leaks for its whole life) and use all Ps (or
// it serializes connections).
func restoreDaemonRuntime() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(100)
}

func cmdDaemon() *cobra.Command {
	var serve bool
	var ensure bool
	var ping bool
	var maxSize string
	var ttl string
	var pprofAddr string

	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the qdf-hookd background daemon",
		Long: `daemon manages qdf-hookd, the long-lived process that answers hook
requests over a unix socket instead of qdf-hook's per-invocation disk round
trip.

  --serve   run the serve loop in the foreground (blocks until idle-exit or
            QUIT)
  --ensure  make sure a live, current daemon is running: no-op if one already
            is, start one if none is, replace it if it's a stale version
  --ping    probe a running daemon and exit 0/1 (for a container HEALTHCHECK)
  --pprof   (with --serve) also serve net/http/pprof on a loopback-only
            addr, e.g. 127.0.0.1:6060 (or $QDF_PPROF)`,
		RunE: func(_ *cobra.Command, _ []string) error {
			switch {
			case ping:
				// Exit 0/1 explicitly (rather than returning an error for
				// main() to report) since this flag exists specifically for a
				// container HEALTHCHECK, which keys off the process exit
				// code, not stderr text.
				reply, err := daemon.Ping(daemon.SockPath(), pingTimeout)
				if err != nil {
					fmt.Fprintln(os.Stderr, "daemon: ping:", err)
					os.Exit(1)
				}
				fmt.Println(reply)
				return nil
			case serve:
				restoreDaemonRuntime()
				if maxSize != "" {
					_ = os.Setenv("QDF_CACHE_MAX_SIZE", maxSize)
				}
				if ttl != "" {
					_ = os.Setenv("QDF_CACHE_TTL", ttl)
				}
				if pprofAddr == "" {
					pprofAddr = os.Getenv("QDF_PPROF")
				}
				if pprofAddr != "" {
					if !isLoopbackAddr(pprofAddr) {
						return fmt.Errorf("daemon: --pprof %q must be loopback-only (127.0.0.1:PORT, localhost:PORT, or [::1]:PORT)", pprofAddr)
					}
					go func() {
						// ListenAndServe only returns on error (e.g. the port is
						// taken); it never blocks Serve below since it runs on its
						// own goroutine. There is intentionally no way to stop it
						// short of the whole daemon exiting — it lives exactly as
						// long as the daemon does.
						_ = http.ListenAndServe(pprofAddr, nil)
					}()
				}
				return daemon.Serve(daemon.SockPath(), daemonIdleTimeout, appVersion)
			case ensure:
				// SessionStart runs `daemon --ensure`, so this is the throttled
				// (≤once/24h) automatic disk-cache prune point — it fires even
				// if the daemon can't start, which is the only automatic prune a
				// CLI-only user gets (no daemon sweep ticker for them).
				cache.AutoSweep(time.Now().Unix())
				exe, err := os.Executable()
				if err != nil {
					return fmt.Errorf("daemon: resolve executable path: %w", err)
				}
				return daemon.Ensure(daemon.SockPath(), exe, appVersion)
			default:
				return errors.New("daemon: specify --serve, --ensure, or --ping")
			}
		},
	}
	cmd.Flags().BoolVar(&serve, "serve", false, "run the daemon serve loop in the foreground")
	cmd.Flags().BoolVar(&ensure, "ensure", false, "ensure a live, current daemon is running")
	cmd.Flags().BoolVar(&ping, "ping", false, "probe a running daemon and print its version reply (exit 0/1)")
	cmd.Flags().StringVar(&maxSize, "cache-max-size", "", "override refs/last size cap in bytes (default 128MiB / $QDF_CACHE_MAX_SIZE)")
	cmd.Flags().StringVar(&ttl, "cache-ttl", "", "override cache TTL, e.g. 720h (default / $QDF_CACHE_TTL)")
	cmd.Flags().StringVar(&pprofAddr, "pprof", "", "with --serve, also serve net/http/pprof on this loopback-only addr (default \"\" / $QDF_PPROF)")
	return cmd
}
