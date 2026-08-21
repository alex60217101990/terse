package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

// defaultPprofAddr is where `daemon --serve --pprof` listens if the operator
// doesn't override it — matches --pprof's own flag default in daemon.go.
const defaultPprofAddr = "127.0.0.1:6060"

// defaultProfileDuration is how long a sampling profile (cpu/block/mutex)
// collects before the daemon returns it, absent -d.
const defaultProfileDuration = 30 * time.Second

// profileEndpoints maps each supported `qdf-hook profile <type>` argument to
// its net/http/pprof path, relative to /debug/pprof/. cpu is special-cased in
// runProfile (it needs a ?seconds= query param); the rest are plain named
// profiles served as-is by net/http/pprof's index handler.
var profileEndpoints = map[string]string{
	"cpu":       "profile",
	"heap":      "heap",
	"goroutine": "goroutine",
	"allocs":    "allocs",
	"mutex":     "mutex",
	"block":     "block",
}

func cmdProfile() *cobra.Command {
	var out string
	var dur time.Duration
	var addr string

	cmd := &cobra.Command{
		Use:   "profile [cpu|heap|goroutine|allocs|mutex|block]",
		Short: "Fetch a pprof profile from the running daemon for tuning",
		Long: `profile fetches a pprof profile from a daemon started with
--pprof and either writes it to -o or, absent -o, opens it directly in
"go tool pprof". Requires the daemon to be running with:

  qdf-hook daemon --serve --pprof 127.0.0.1:6060`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if addr == "" {
				addr = os.Getenv("QDF_PPROF")
			}
			if addr == "" {
				addr = defaultPprofAddr
			}
			return runProfile(args[0], addr, out, dur)
		},
	}
	cmd.Flags().StringVarP(&out, "output", "o", "", "write the profile to this file instead of opening it in go tool pprof")
	cmd.Flags().DurationVarP(&dur, "duration", "d", defaultProfileDuration, "sample duration for cpu/block/mutex profiles")
	cmd.Flags().StringVar(&addr, "addr", "", "daemon --pprof addr to fetch from (default 127.0.0.1:6060 / $QDF_PPROF)")
	return cmd
}

// runProfile fetches the named pprof profile from the daemon's pprof server
// at addr, writes it to outPath (or a temp file if outPath is empty), and —
// only when outPath was empty, meaning the caller wants to look at it now
// rather than just capture it — execs `go tool pprof` on the result.
func runProfile(kind, addr, outPath string, dur time.Duration) error {
	path, ok := profileEndpoints[kind]
	if !ok {
		return fmt.Errorf("profile: unknown type %q (want one of cpu, heap, goroutine, allocs, mutex, block)", kind)
	}

	url := fmt.Sprintf("http://%s/debug/pprof/%s", addr, path)
	switch kind {
	case "cpu", "block", "mutex":
		url = fmt.Sprintf("%s?seconds=%d", url, int(dur.Seconds()))
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("profile: build request for %s: %w", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("profile: fetch %s: %w (is the daemon running with --pprof %s?)", url, err, addr)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("profile: fetch %s: HTTP %d: %s", url, resp.StatusCode, body)
	}

	wroteTemp := outPath == ""
	if !wroteTemp {
		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("profile: create %s: %w", outPath, err)
		}
		_, cerr := io.Copy(f, resp.Body)
		_ = f.Close()
		if cerr != nil {
			return fmt.Errorf("profile: write %s: %w", outPath, cerr)
		}
		fmt.Printf("wrote %s profile to %s\n", kind, outPath)
		return nil
	}

	f, err := os.CreateTemp("", "qdf-hook-profile-*.pb.gz")
	if err != nil {
		return fmt.Errorf("profile: create temp file: %w", err)
	}
	outPath = f.Name()
	_, cerr := io.Copy(f, resp.Body)
	_ = f.Close()
	if cerr != nil {
		return fmt.Errorf("profile: write %s: %w", outPath, cerr)
	}

	pprofCmd := exec.CommandContext(context.Background(), "go", "tool", "pprof", outPath)
	pprofCmd.Stdin = os.Stdin
	pprofCmd.Stdout = os.Stdout
	pprofCmd.Stderr = os.Stderr
	return pprofCmd.Run()
}
