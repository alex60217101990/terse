package main

import (
	"os"

	"github.com/alex60217101990/qdf-hook/internal/analytics"
	"github.com/spf13/cobra"
)

func cmdBarsDemo() *cobra.Command {
	return &cobra.Command{
		Use:    "barsdemo",
		Short:  "Preview progress-bar styles (with color) to pick one",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			analytics.RenderBarGallery(os.Stdout)
			return nil
		},
	}
}
