package main

import (
	"fmt"
	"os"

	"github.com/alex60217101990/terse/internal/analytics"
	"github.com/alex60217101990/terse/internal/cache"
	"github.com/spf13/cobra"
)

func cmdGC() *cobra.Command {
	var dryRun bool
	var minScore float64
	var maxSize, ttl string

	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Remove stale session state files by utility score",
		RunE: func(cmd *cobra.Command, args []string) error {
			if maxSize != "" {
				_ = os.Setenv("QDF_CACHE_MAX_SIZE", maxSize)
			}
			if ttl != "" {
				_ = os.Setenv("QDF_CACHE_TTL", ttl)
			}
			result, err := cache.RunGC(dryRun, minScore)
			if err != nil {
				return fmt.Errorf("gc: %w", err)
			}
			if dryRun {
				fmt.Printf("dry-run: would remove %d sessions, keep %d\n", result.Removed, result.Kept)
			} else {
				fmt.Printf("removed %d sessions + %d blobs, freed %s\n",
					result.Removed, result.BlobsRemoved,
					analytics.FormatBytes(int(result.FreedBytes+result.BlobBytesFreed)))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print what would be removed without deleting")
	cmd.Flags().Float64Var(&minScore, "min-score", 0.01, "sessions below this utility score are removed")
	cmd.Flags().StringVar(&maxSize, "cache-max-size", "", "override refs/last size cap in bytes (default 128MiB / $QDF_CACHE_MAX_SIZE)")
	cmd.Flags().StringVar(&ttl, "cache-ttl", "", "override cache TTL, e.g. 720h (default / $QDF_CACHE_TTL)")
	return cmd
}
