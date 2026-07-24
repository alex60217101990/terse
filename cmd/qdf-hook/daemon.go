package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/alex60217101990/qdf-hook/internal/daemon"
	"github.com/spf13/cobra"
)

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
	var maxSize string
	var ttl string

	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the qdf-hookd background daemon",
		Long: `daemon manages qdf-hookd, the long-lived process that answers hook
requests over a unix socket instead of qdf-hook's per-invocation disk round
trip.

  --serve   run the serve loop in the foreground (blocks until idle-exit or
            QUIT)
  --ensure  make sure a live, current daemon is running: no-op if one already
            is, start one if none is, replace it if it's a stale version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case serve:
				restoreDaemonRuntime()
				if maxSize != "" {
					_ = os.Setenv("QDF_CACHE_MAX_SIZE", maxSize)
				}
				if ttl != "" {
					_ = os.Setenv("QDF_CACHE_TTL", ttl)
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
				return fmt.Errorf("daemon: specify --serve or --ensure")
			}
		},
	}
	cmd.Flags().BoolVar(&serve, "serve", false, "run the daemon serve loop in the foreground")
	cmd.Flags().BoolVar(&ensure, "ensure", false, "ensure a live, current daemon is running")
	cmd.Flags().StringVar(&maxSize, "cache-max-size", "", "override refs/last size cap in bytes (default 128MiB / $QDF_CACHE_MAX_SIZE)")
	cmd.Flags().StringVar(&ttl, "cache-ttl", "", "override cache TTL, e.g. 720h (default / $QDF_CACHE_TTL)")
	return cmd
}
