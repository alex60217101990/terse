package main

import (
	"fmt"
	"strings"

	"github.com/alex60217101990/terse/internal/cache"
	"github.com/spf13/cobra"
)

func cmdExpand() *cobra.Command {
	return &cobra.Command{
		Use:   "expand <hash>",
		Short: "Print the full content behind a §ref:HASH§ token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExpand(args[0])
		},
	}
}

// runExpand resolves a §ref token or bare hash to its stored content.
func runExpand(arg string) error {
	hash := strings.TrimSuffix(strings.TrimPrefix(arg, "§ref:"), "§")
	content, ok := cache.RefGet(hash)
	if !ok {
		return fmt.Errorf("no ref blob for %q", hash)
	}
	fmt.Print(content)
	return nil
}
