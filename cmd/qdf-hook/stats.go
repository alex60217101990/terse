package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alex60217101990/terse/internal/analytics"
)

func cmdStats() *cobra.Command {
	var jsonOut bool
	var days int
	var style string
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show token savings analytics",
		RunE: func(_ *cobra.Command, args []string) error {
			events, err := analytics.LoadEvents(days)
			if err != nil {
				return fmt.Errorf("load events: %w", err)
			}
			if len(events) == 0 {
				fmt.Fprintln(os.Stderr, "qdf-hook: no analytics data found. Run some hooks first.")
				return nil
			}
			stats := analytics.ComputeStats(events)
			analytics.PrintStats(stats, jsonOut, style, os.Stdout)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	cmd.Flags().IntVar(&days, "days", 7, "number of days to include")
	cmd.Flags().StringVar(&style, "style", "line", "bar style: blocks|line|shade|braille")
	return cmd
}
