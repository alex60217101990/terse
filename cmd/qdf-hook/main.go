package main

import (
	"fmt"
	"os"

	"github.com/alex60217101990/qdf-hook/internal/protocol"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qdf-hook",
	Short: "Claude Code hook to reduce token consumption via compression",
	Long: `qdf-hook is a PostToolUse/PreCompact/PostCompact hook for Claude Code.
It intercepts tool output and compresses it to reduce token consumption.`,
}

func init() {
	rootCmd.AddCommand(
		cmdVersion(),
		cmdRead(),
		cmdBash(),
		cmdPreCompact(),
		cmdPostCompact(),
		cmdSessionStart(),
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "qdf-hook:", err)
	}
}

// runVersion prints the current version.
func runVersion() error {
	fmt.Println("qdf-hook v0.1.0")
	return nil
}

// stubs — filled in by later tasks

func runRead() error         { return protocol.EncodeOutput(os.Stdout, protocol.Passthrough()) }
func runBash() error         { return protocol.EncodeOutput(os.Stdout, protocol.Passthrough()) }
func runPreCompact() error   { return protocol.EncodeOutput(os.Stdout, protocol.Passthrough()) }
func runPostCompact() error  { return protocol.EncodeOutput(os.Stdout, protocol.Passthrough()) }
func runSessionStart() error { return protocol.EncodeOutput(os.Stdout, protocol.Passthrough()) }
