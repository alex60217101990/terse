package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alex60217101990/qdf-hook/internal/daemon"
	"github.com/spf13/cobra"
)

// daemonIdleTimeout is how long qdf-hookd sits idle (no connections) before
// it exits on its own. Long enough to survive gaps between hook calls within
// a session, short enough not to linger forever after the session ends.
const daemonIdleTimeout = 30 * time.Minute

func cmdDaemon() *cobra.Command {
	var serve bool
	var ensure bool

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
				return daemon.Serve(daemon.SockPath(), daemonIdleTimeout, appVersion)
			case ensure:
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
	return cmd
}
