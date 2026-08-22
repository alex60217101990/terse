package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alex60217101990/terse/internal/cache"
)

func cmdExpand() *cobra.Command {
	return &cobra.Command{
		Use:   "expand <hash>",
		Short: "Print the full content behind a §ref:HASH§ token",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runExpand(args[0])
		},
	}
}

// runExpand resolves a §ref token, a bare hash, or a capped command's capture
// id to its stored content.
//
// Both stores are searched because both hand the model an id and tell it to
// come back here: §ref for a deduplicated tool result, and the capture id the
// output-capping wrapper prints in its elision line. A capped command's full
// output is worthless if the handle printed next to it does not resolve.
func runExpand(arg string) error {
	hash := strings.TrimSuffix(strings.TrimPrefix(arg, "§ref:"), "§")
	if content, ok := cache.RefGet(hash); ok {
		fmt.Print(content)
		return nil
	}
	if content, ok := cache.CaptureGet(hash); ok {
		fmt.Print(content)
		return nil
	}
	return fmt.Errorf("nothing stored for %q: no ref blob and no capture", hash)
}
