package main

import (
	"fmt"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/alex60217101990/qdf-hook/internal/cache"
	"github.com/spf13/cobra"
)

func cmdGC() *cobra.Command {
	var dryRun bool
	var minScore float64

	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Remove stale session state files by utility score",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := cache.RunGC(dryRun, minScore)
			if err != nil {
				return fmt.Errorf("gc: %w", err)
			}
			if dryRun {
				fmt.Printf("dry-run: would remove %d sessions, keep %d\n", result.Removed, result.Kept)
			} else {
				fmt.Printf("removed %d sessions, kept %d, freed %s\n",
					result.Removed, result.Kept, analytics.FormatBytes(int(result.FreedBytes)))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would be removed without deleting")
	cmd.Flags().Float64Var(&minScore, "min-score", 0.01, "sessions below this utility score are removed")
	return cmd
}
