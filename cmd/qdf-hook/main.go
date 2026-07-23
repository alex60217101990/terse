package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"

	"github.com/alex60217101990/qdf-hook/internal/hook"
	"github.com/spf13/cobra"
)

func init() {
	// qdf-hook is a one-shot CLI: it does a single unit of work and exits in
	// microseconds. Pin to a single P and disable the GC so the runtime doesn't
	// spin up extra scheduler/GC machinery it will never need — a measurable
	// cut in per-invocation startup. Safe: nothing runs long enough to GC, and
	// peak memory is bounded by the one input being processed.
	runtime.GOMAXPROCS(1)
	debug.SetGCPercent(-1)
}

var (
	cpuprofile string
	memprofile string
	cpuFile    *os.File
)

var rootCmd = &cobra.Command{
	Use:   "qdf-hook",
	Short: "Claude Code hook to reduce token consumption via compression",
	Long: `qdf-hook is a PostToolUse/PreCompact/PostCompact hook for Claude Code.
It intercepts tool output and compresses it to reduce token consumption.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// CPU profiling starts here because cobra only parses the
	// --cpuprofile/--memprofile flags during rootCmd.Execute(); reading them
	// in main() before Execute() would always see "". Stopping the profile
	// and writing the heap profile happens back in main() after Execute()
	// returns — not in a PersistentPostRunE, which cobra skips when the
	// command's RunE returns an error (we still want the profile then).
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cpuprofile == "" {
			return nil
		}
		f, err := os.Create(cpuprofile)
		if err != nil {
			return fmt.Errorf("cpuprofile: %w", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			f.Close()
			return fmt.Errorf("cpuprofile: %w", err)
		}
		cpuFile = f
		return nil
	},
}

// stopProfiling stops any in-progress CPU profile and writes the heap profile.
// Called from main() after Execute() so profiles are flushed even when a hook
// command returns an error.
func stopProfiling() {
	if cpuFile != nil {
		pprof.StopCPUProfile()
		cpuFile.Close()
	}
	if memprofile == "" {
		return
	}
	f, err := os.Create(memprofile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "qdf-hook: memprofile:", err)
		return
	}
	defer f.Close()
	if err := pprof.WriteHeapProfile(f); err != nil {
		fmt.Fprintln(os.Stderr, "qdf-hook: memprofile:", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cpuprofile, "cpuprofile", "", "write CPU profile to file")
	rootCmd.PersistentFlags().StringVar(&memprofile, "memprofile", "", "write memory profile to file")
	rootCmd.AddCommand(
		cmdVersion(),
		cmdPost(),
		cmdRead(),
		cmdBash(),
		cmdGlob(),
		cmdGrep(),
		cmdWrite(),
		cmdPreToolUse(),
		cmdPreCompact(),
		cmdPostCompact(),
		cmdSessionStart(),
		cmdStats(),
		cmdGC(),
		cmdExpand(),
		cmdInit(),
	)
}

func cmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion()
		},
	}
}

func cmdRead() *cobra.Command {
	return &cobra.Command{
		Use:   "read",
		Short: "Handle PostToolUse hook for the Read tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRead()
		},
	}
}

func cmdBash() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Handle PostToolUse hook for the Bash tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBash()
		},
	}
}

func cmdPreToolUse() *cobra.Command {
	return &cobra.Command{
		Use:   "pretooluse",
		Short: "Handle PreToolUse hook for the Read tool (mtime fast-path)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPreToolUse()
		},
	}
}

func cmdPreCompact() *cobra.Command {
	return &cobra.Command{
		Use:   "precompact",
		Short: "Handle PreCompact hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPreCompact()
		},
	}
}

func cmdPostCompact() *cobra.Command {
	return &cobra.Command{
		Use:   "postcompact",
		Short: "Handle PostCompact hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPostCompact()
		},
	}
}

func cmdGlob() *cobra.Command {
	return &cobra.Command{
		Use:   "glob",
		Short: "Handle PostToolUse hook for Glob tool — compress file list to tree",
		RunE:  func(cmd *cobra.Command, args []string) error { return hook.HandleGlob(os.Stdin, os.Stdout) },
	}
}

func cmdPost() *cobra.Command {
	return &cobra.Command{
		Use:   "post",
		Short: "Universal PostToolUse hook — routes any tool through the pipeline",
		RunE:  func(cmd *cobra.Command, args []string) error { return hook.Dispatch(os.Stdin, os.Stdout) },
	}
}

func cmdGrep() *cobra.Command {
	return &cobra.Command{
		Use:   "grep",
		Short: "Handle PostToolUse hook for Grep tool — group matches by file",
		RunE:  func(cmd *cobra.Command, args []string) error { return hook.HandleGrep(os.Stdin, os.Stdout) },
	}
}

func cmdWrite() *cobra.Command {
	return &cobra.Command{
		Use:   "write",
		Short: "Handle PostToolUse hook for Write/Edit/MultiEdit — suppress content echo",
		RunE:  func(cmd *cobra.Command, args []string) error { return hook.HandleWrite(os.Stdin, os.Stdout) },
	}
}

func cmdSessionStart() *cobra.Command {
	return &cobra.Command{
		Use:   "sessionstart",
		Short: "Handle session start hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSessionStart()
		},
	}
}

func main() {
	err := rootCmd.Execute()
	stopProfiling()
	if err != nil {
		fmt.Fprintln(os.Stderr, "qdf-hook:", err)
	}
}

// runVersion prints the current version.
func runVersion() error {
	fmt.Println("qdf-hook v0.1.0")
	return nil
}

// stubs — filled in by later tasks

func runRead() error         { return hook.HandleRead(os.Stdin, os.Stdout) }
func runBash() error         { return hook.HandleBash(os.Stdin, os.Stdout) }
func runPreToolUse() error   { return hook.HandlePreToolUse(os.Stdin, os.Stdout) }
func runPreCompact() error   { return hook.HandlePreCompact(os.Stdin, os.Stdout) }
func runPostCompact() error  { return hook.HandlePostCompact(os.Stdin, os.Stdout) }
func runSessionStart() error { return hook.HandleSessionStart(os.Stdin, os.Stdout) }
